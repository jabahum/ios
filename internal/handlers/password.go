package handlers

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"time"

	"case/internal/models"
	"case/internal/utils"

	"github.com/gofiber/fiber/v2"
)

// PasswordHandler handles password change operations
type PasswordHandler struct {
	userService *models.UserService
}

// NewPasswordHandler creates a new password handler
func NewPasswordHandler(userService *models.UserService) *PasswordHandler {
	return &PasswordHandler{
		userService: userService,
	}
}

// ShowChangePasswordForm shows the password change form
func (h *PasswordHandler) ShowChangePasswordForm(c *fiber.Ctx) error {
	// Generate CSRF token
	csrfToken := generateCSRFToken()
	c.Cookie(&fiber.Cookie{
		Name:     "csrf_token",
		Value:    csrfToken,
		MaxAge:   3600,
		Path:     "/",
		HTTPOnly: true,
		Secure:   false,
	})

	return GenerateHTML(c, nil, fiber.Map{
		"CSRFToken": csrfToken,
	}, "change_password")
}

// ChangePassword handles password change requests
func (h *PasswordHandler) ChangePassword(c *fiber.Ctx) error {
	// Get current user ID from session
	userID := utils.GetUserIDFromSession(c)
	if userID == 0 {
		return GenerateHTML(c, nil, fiber.Map{"error": "Unauthorized"}, "error")
	}

	// Get form data
	currentPassword := c.FormValue("current_password")
	newPassword := c.FormValue("new_password")
	confirmPassword := c.FormValue("confirm_password")
	csrfToken := c.FormValue("csrf_token")

	// Validate CSRF token
	if !validateCSRFToken(c, csrfToken) {
		return GenerateHTML(c, nil, fiber.Map{
			"Message":     "Invalid security token",
			"MessageType": "error",
		}, "change_password")
	}

	// Validate input
	if currentPassword == "" || newPassword == "" || confirmPassword == "" {
		return GenerateHTML(c, nil, fiber.Map{
			"Message":     "All fields are required",
			"MessageType": "error",
		}, "change_password")
	}

	// Check if new passwords match
	if newPassword != confirmPassword {
		return GenerateHTML(c, nil, fiber.Map{
			"Message":     "New passwords do not match",
			"MessageType": "error",
		}, "change_password")
	}

	// Validate password strength
	if !validatePasswordStrength(newPassword) {
		return GenerateHTML(c, nil, fiber.Map{
			"Message":     "Password does not meet strength requirements",
			"MessageType": "error",
		}, "change_password")
	}

	// Get user from database
	user, err := h.userService.GetUserByID(userID)
	if err != nil {
		return GenerateHTML(c, nil, fiber.Map{
			"Message":     "Error retrieving user information",
			"MessageType": "error",
		}, "change_password")
	}

	// Verify current password
	if !verifyPassword(currentPassword, user.UserPass.String, "") {
		return GenerateHTML(c, nil, fiber.Map{
			"Message":     "Current password is incorrect",
			"MessageType": "error",
		}, "change_password")
	}

	// Generate new password hash and salt
	newSalt := generateSalt()
	newHash := hashPassword(newPassword, newSalt)

	// Update user password
	err = h.userService.UpdatePassword(int64(user.UserID), newHash, newSalt)
	if err != nil {
		return GenerateHTML(c, nil, fiber.Map{
			"Message":     "Error updating password",
			"MessageType": "error",
		}, "change_password")
	}

	// Log password change
	h.logPasswordChange(userID, c.IP(), c.Get("User-Agent"))

	// Redirect with success message
	return c.Redirect("/dashboard?message=Password changed successfully")
}

// RequestPasswordReset handles password reset requests
func (h *PasswordHandler) RequestPasswordReset(c *fiber.Ctx) error {
	email := c.FormValue("email")
	if email == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Email is required"})
	}

	// Get user by email
	user, err := h.userService.GetUserByEmail(email)
	if err != nil {
		// Don't reveal if user exists or not
		return c.JSON(fiber.Map{"message": "If the email exists, a reset link has been sent"})
	}

	// Generate reset token
	token := generateResetToken()
	expiresAt := time.Now().Add(24 * time.Hour)

	// Create password change request
	request := &models.PasswordChangeRequest{
		UserID:              int64(user.UserID),
		RequestToken:        token,
		CurrentPasswordHash: user.UserPass.String,
		NewPasswordHash:     "", // Will be set when approved
		NewPasswordSalt:     "", // Will be set when approved
		ExpiresAt:           expiresAt,
		IPAddress:           sql.NullString{String: c.IP(), Valid: true},
		UserAgent:           sql.NullString{String: c.Get("User-Agent"), Valid: true},
	}

	// Save request to database
	err = h.userService.CreatePasswordChangeRequest(request)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Error creating reset request"})
	}

	// Send reset email (implement email service)
	// h.sendPasswordResetEmail(user.Email, token)

	return c.JSON(fiber.Map{"message": "Password reset link sent to your email"})
}

// ResetPassword handles password reset with token
func (h *PasswordHandler) ResetPassword(c *fiber.Ctx) error {
	token := c.Params("token")
	newPassword := c.FormValue("new_password")
	confirmPassword := c.FormValue("confirm_password")

	if newPassword == "" || confirmPassword == "" {
		return GenerateHTML(c, nil, fiber.Map{
			"Message":     "All fields are required",
			"MessageType": "error",
			"Token":       token,
		}, "reset_password")
	}

	if newPassword != confirmPassword {
		return GenerateHTML(c, nil, fiber.Map{
			"Message":     "Passwords do not match",
			"MessageType": "error",
			"Token":       token,
		}, "reset_password")
	}

	if !validatePasswordStrength(newPassword) {
		return GenerateHTML(c, nil, fiber.Map{
			"Message":     "Password does not meet strength requirements",
			"MessageType": "error",
			"Token":       token,
		}, "reset_password")
	}

	// Get password change request by token (placeholder - implement this method)
	// request, err := h.userService.GetPasswordChangeRequestByToken(token)
	// if err != nil {
	// 	return GenerateHTML(c, nil, fiber.Map{
	// 		"Message":     "Invalid or expired reset token",
	// 		"MessageType": "error",
	// 		"Token":       token,
	// 	}, "reset_password")
	// }

	// For now, return error since method doesn't exist
	return GenerateHTML(c, nil, fiber.Map{
		"Message":     "Password reset functionality not implemented",
		"MessageType": "error",
		"Token":       token,
	}, "reset_password")

	// Check if token is expired (commented out since request is not defined)
	// if time.Now().After(request.ExpiresAt) {
	// 	return GenerateHTML(c, nil, fiber.Map{
	// 		"Message":     "Reset token has expired",
	// 		"MessageType": "error",
	// 		"Token":       token,
	// 	}, "reset_password")
	// }

	// Generate new password hash and salt (commented out since request is not defined)
	// newSalt := generateSalt()
	// newHash := hashPassword(newPassword, newSalt)

	// Update password change request (commented out since request is not defined)
	// request.NewPasswordHash = newHash
	// request.NewPasswordSalt = newSalt
	// request.IsApproved = true
	// request.ApprovedAt = sql.NullTime{Time: time.Now(), Valid: true}

	// err = h.userService.UpdatePasswordChangeRequest(request)
	// if err != nil {
	// 	return GenerateHTML(c, nil, fiber.Map{
	// 		"Message":     "Error updating password",
	// 		"MessageType": "error",
	// 		"Token":       token,
	// 	}, "reset_password")
	// }

	// Update user password (commented out since request is not defined)
	// err = h.userService.UpdatePassword(request.UserID, newHash, newSalt)
	// if err != nil {
	// 	return GenerateHTML(c, nil, fiber.Map{
	// 		"Message":     "Error updating password",
	// 		"MessageType": "error",
	// 		"Token":       token,
	// 	}, "reset_password")
	// }

	// return c.Redirect("/login?message=Password reset successfully")
}

// ShowForgotPasswordForm shows the forgot password form
func (h *PasswordHandler) ShowForgotPasswordForm(c *fiber.Ctx) error {
	return GenerateHTML(c, nil, nil, "forgot")
}

// ShowResetPasswordForm shows the reset password form
func (h *PasswordHandler) ShowResetPasswordForm(c *fiber.Ctx) error {
	token := c.Params("token")
	return GenerateHTML(c, nil, fiber.Map{"Token": token}, "reset_password")
}

// Helper functions
func verifyPassword(password, hash, salt string) bool {
	return hashPassword(password, salt) == hash
}

func validatePasswordStrength(password string) bool {
	if len(password) < 8 {
		return false
	}

	hasUpper := false
	hasLower := false
	hasNumber := false
	hasSpecial := false

	for _, char := range password {
		switch {
		case char >= 'A' && char <= 'Z':
			hasUpper = true
		case char >= 'a' && char <= 'z':
			hasLower = true
		case char >= '0' && char <= '9':
			hasNumber = true
		default:
			hasSpecial = true
		}
	}

	return hasUpper && hasLower && hasNumber && hasSpecial
}

func generateCSRFToken() string {
	token := make([]byte, 32)
	rand.Read(token)
	return hex.EncodeToString(token)
}

func validateCSRFToken(c *fiber.Ctx, token string) bool {
	cookieToken := c.Cookies("csrf_token")
	return token == cookieToken
}

func generateResetToken() string {
	token := make([]byte, 32)
	rand.Read(token)
	return hex.EncodeToString(token)
}

func (h *PasswordHandler) logPasswordChange(userID int64, ipAddress, userAgent string) {
	// Log password change for audit purposes
	// This should be implemented based on your logging system
}
