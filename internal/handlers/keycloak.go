package handlers

import (
	"database/sql"
	"encoding/base64"
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"

	"case/internal/keycloak"
	"case/internal/models"
	"case/internal/utils"
)

/* -----------------------------
 * Handler
 * ----------------------------- */

type AuthHandler struct {
	kc          *keycloak.Client
	config      AuthConfig
	logger      *slog.Logger
	db          *sql.DB
	store       *session.Store
	userService *models.UserService
}

func NewAuthHandler(
	kc *keycloak.Client,
	db *sql.DB,
	store *session.Store,
	userService *models.UserService,
	config AuthConfig,
	logger *slog.Logger,
) *AuthHandler {
	return &AuthHandler{
		kc:          kc,
		db:          db,
		store:       store,
		userService: userService,
		config:      config,
		logger:      logger,
	}
}

/* -----------------------------
 * Routes
 * ----------------------------- */

// GET /api/v1/auth/login
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	state := utils.RandomString(32)
	codeVerifier := utils.RandomString(64)
	codeChallenge := base64.RawURLEncoding.EncodeToString(
		utils.Sha256Bytes(codeVerifier),
	)

	// Store PKCE + state in short-lived cookies
	setCookie(c, "auth_state", state, 5*time.Minute, h)
	setCookie(c, "code_verifier", codeVerifier, 5*time.Minute, h)

	loginURL := h.kc.BuildLoginURL(
		h.config.RedirectURI,
		state,
		codeChallenge,
	)

	return c.Redirect(loginURL, fiber.StatusFound)
}

// GET /api/v1/auth/callback
func (h *AuthHandler) Callback(c *fiber.Ctx) error {
	code := c.Query("code")
	state := c.Query("state")

	if code == "" || state == "" {
		return c.Status(fiber.StatusBadRequest).
			JSON(fiber.Map{"error": "missing code or state"})
	}

	storedState := c.Cookies("auth_state")
	if storedState == "" || storedState != state {
		return c.Status(fiber.StatusUnauthorized).
			JSON(fiber.Map{"error": "invalid state"})
	}

	codeVerifier := c.Cookies("code_verifier")
	if codeVerifier == "" {
		return c.Status(fiber.StatusUnauthorized).
			JSON(fiber.Map{"error": "missing code verifier"})
	}

	// 1. Exchange code for token
	token, err := h.kc.ExchangeCodeForToken(
		code,
		h.config.RedirectURI,
		codeVerifier,
	)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).
			JSON(fiber.Map{"error": err.Error()})
	}

	// 2. Fetch user info from Keycloak
	userInfo, err := h.kc.UserInfo(token.AccessToken)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).
			JSON(fiber.Map{"error": "failed to fetch userinfo"})
	}

	// 3. Resolve local user
	user, err := h.userService.GetUserByEmail(userInfo.Email)
	if err != nil || user == nil {
		h.logger.Warn(
			"Keycloak user not found locally",
			"email", userInfo.Email,
			"username", userInfo.PreferredUsername,
		)
		return c.Redirect("/login?error=user_not_registered")
	}

	// 4. Create Fiber session (THIS is the bridge)
	sess, err := h.store.Get(c)
	if err != nil {
		return c.Redirect("/login?error=session")
	}

	if err := sess.Regenerate(); err != nil {
		return err
	}

	sess.Set("user", user.UserID)
	sess.Set("user_id", user.UserID)
	sess.Set("userID", user.UserID)
	sess.Set("username", userInfo.PreferredUsername)
	sess.Set("isAuthenticated", true)
	sess.Set("authenticated", true)
	sess.Set("auth_provider", "keycloak")

	if err := sess.Save(); err != nil {
		return c.Redirect("/login?error=session_save")
	}

	// 5. Clear temporary cookies
	clearCookie(c, "auth_state", h)
	clearCookie(c, "code_verifier", h)

	// 6. Set auth cookies (keep existing frontend behavior)
	setCookie(
		c,
		"access_token",
		token.AccessToken,
		time.Duration(token.ExpiresIn)*time.Second,
		h,
	)
	setCookie(
		c,
		"refresh_token",
		token.RefreshToken,
		time.Duration(token.RefreshExpiresIn)*time.Second,
		h,
	)

	return c.Redirect("/home", fiber.StatusFound)
}

// POST /api/v1/auth/refresh
func (h *AuthHandler) Refresh(c *fiber.Ctx) error {
	refreshToken := c.Cookies("refresh_token")
	if refreshToken == "" {
		return c.Status(fiber.StatusUnauthorized).
			JSON(fiber.Map{"error": "missing refresh token"})
	}

	token, err := h.kc.AccessToken(refreshToken)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).
			JSON(fiber.Map{"error": err.Error()})
	}

	setCookie(
		c,
		"access_token",
		token.AccessToken,
		time.Duration(token.ExpiresIn)*time.Second,
		h,
	)
	setCookie(
		c,
		"refresh_token",
		token.RefreshToken,
		time.Duration(token.RefreshExpiresIn)*time.Second,
		h,
	)

	return c.JSON(fiber.Map{"success": true})
}

// POST /api/v1/auth/logout
func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	refreshToken := c.Cookies("refresh_token")
	if refreshToken != "" {
		_ = h.kc.LogOut(refreshToken)
	}

	// Destroy Fiber session
	sess, _ := h.store.Get(c)
	if sess != nil {
		sess.Destroy()
	}

	clearCookie(c, "access_token", h)
	clearCookie(c, "refresh_token", h)

	return c.Redirect("/login")
}

// GET /api/v1/auth/me
func (h *AuthHandler) Me(c *fiber.Ctx) error {
	accessToken := c.Cookies("access_token")
	if accessToken == "" {
		return c.Status(fiber.StatusUnauthorized).
			JSON(fiber.Map{"error": "unauthenticated"})
	}

	user, err := h.kc.UserInfo(accessToken)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).
			JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"user": user})
}

/* -----------------------------
 * Cookie helpers (Fiber)
 * ----------------------------- */

func setCookie(
	c *fiber.Ctx,
	name string,
	value string,
	maxAge time.Duration,
	h *AuthHandler,
) {
	c.Cookie(&fiber.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		Domain:   h.config.CookieDomain,
		MaxAge:   int(maxAge.Seconds()),
		Secure:   h.config.CookieSecure,
		HTTPOnly: true,
		SameSite: "Lax",
	})
}

func clearCookie(c *fiber.Ctx, name string, h *AuthHandler) {
	c.Cookie(&fiber.Cookie{
		Name:     name,
		Value:    "",
		Path:     "/",
		Domain:   h.config.CookieDomain,
		MaxAge:   -1,
		Secure:   h.config.CookieSecure,
		HTTPOnly: true,
		SameSite: "Lax",
	})
}
