package middleware

import (
	"case/internal/models"
	"context"
	"database/sql"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
)

// PermissionRequired creates middleware that checks if user has specific permission
func PermissionRequired(store *session.Store, db *sql.DB, sl *slog.Logger, resource, action string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get user session
		sess, err := store.Get(c)
		if err != nil {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
				"error": "Authentication required",
			})
		}

		// Check if user is authenticated
		if sess.Get("isAuthenticated") != true {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
				"error": "Authentication required",
			})
		}

		// Get user ID from session (try both keys for compatibility)
		var userID int
		var ok bool
		// Try user_id first (RBAC standard)
		if userID, ok = sess.Get("user_id").(int); !ok {
			// Fallback to user (legacy authentication)
			if userID, ok = sess.Get("user").(int); !ok {
				return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
					"error": "Invalid user session",
				})
			}
		}

		// Check permission
		hasPermission, err := models.HasPermission(c.Context(), db, userID, resource, action)
		if err != nil {
			sl.Error("Error checking permission", "error", err, "user_id", userID, "resource", resource, "action", action)
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"error": "Internal server error",
			})
		}

		if !hasPermission {
			// Log unauthorized access attempt
			auditLog := &models.AuditLog{
				UserID:    sql.NullInt64{Int64: int64(userID), Valid: true},
				Action:    "unauthorized_access",
				Resource:  resource,
				Details:   "Permission denied: " + action,
				IPAddress: c.IP(),
				UserAgent: c.Get("User-Agent"),
				CreatedAt: time.Now(),
			}
			models.LogAuditEvent(c.Context(), db, auditLog)

			return c.Status(http.StatusForbidden).JSON(fiber.Map{
				"error":   "Access denied",
				"message": "You don't have permission to perform this action",
			})
		}

		// Add user info to context for handlers
		c.Locals("current_user_id", userID)
		c.Locals("current_user_permissions", getCachedPermissions(c, db, store, userID))

		return c.Next()
	}
}

// RoleRequired creates middleware that checks if user has specific role
func RoleRequired(store *session.Store, db *sql.DB, sl *slog.Logger, roleName string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get user session
		sess, err := store.Get(c)
		if err != nil {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
				"error": "Authentication required",
			})
		}

		// Check if user is authenticated
		if sess.Get("isAuthenticated") != true {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
				"error": "Authentication required",
			})
		}

		// Get user ID from session (try both keys for compatibility)
		var userID int
		var ok bool
		// Try user_id first (RBAC standard)
		if userID, ok = sess.Get("user_id").(int); !ok {
			// Fallback to user (legacy authentication)
			if userID, ok = sess.Get("user").(int); !ok {
				return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
					"error": "Invalid user session",
				})
			}
		}

		// Get user roles
		roles, err := models.GetUserRoles(c.Context(), db, userID)
		if err != nil {
			sl.Error("Error getting user roles", "error", err, "user_id", userID)
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"error": "Internal server error",
			})
		}

		// Check if user has the required role
		hasRole := false
		for _, role := range roles {
			if role.Name == roleName {
				hasRole = true
				break
			}
		}

		if !hasRole {
			// Log unauthorized access attempt
			auditLog := &models.AuditLog{
				UserID:    sql.NullInt64{Int64: int64(userID), Valid: true},
				Action:    "unauthorized_access",
				Resource:  "role_check",
				Details:   "Role denied: " + roleName,
				IPAddress: c.IP(),
				UserAgent: c.Get("User-Agent"),
				CreatedAt: time.Now(),
			}
			models.LogAuditEvent(c.Context(), db, auditLog)

			return c.Status(http.StatusForbidden).JSON(fiber.Map{
				"error":   "Access denied",
				"message": "You don't have the required role to access this resource",
			})
		}

		return c.Next()
	}
}

// EnhancedAuthRequired combines authentication with basic permission checking
func EnhancedAuthRequired(store *session.Store, db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get user session
		sess, err := store.Get(c)
		if err != nil {
			return c.Redirect("/login")
		}

		// Check if user is authenticated
		if sess.Get("isAuthenticated") != true {
			return c.Redirect("/login")
		}

		// Get user ID from session
		userID, ok := sess.Get("user_id").(int)
		if !ok {
			return c.Redirect("/login")
		}

		// Check if user is active (you might want to add this check)
		// For now, we'll just add user info to context
		c.Locals("current_user_id", userID)
		c.Locals("current_user_permissions", getCachedPermissions(c, db, store, userID))

		return c.Next()
	}
}

// AuditMiddleware logs all requests for audit purposes
func AuditMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		// Process request
		err := c.Next()

		// Log the request
		duration := time.Since(start)

		// Get user info if available
		var userID sql.NullInt64
		if currentUserID := c.Locals("current_user_id"); currentUserID != nil {
			if uid, ok := currentUserID.(int); ok {
				userID = sql.NullInt64{Int64: int64(uid), Valid: true}
			}
		}

		// Create audit log entry
		auditLog := &models.AuditLog{
			UserID:    userID,
			Action:    c.Method(),
			Resource:  c.Path(),
			Details:   createAuditDetails(c, duration),
			IPAddress: c.IP(),
			UserAgent: c.Get("User-Agent"),
			CreatedAt: time.Now(),
		}

		// Only log if we have a database connection
		if db := c.Locals("db"); db != nil {
			if database, ok := db.(*sql.DB); ok {
				go models.LogAuditEvent(context.Background(), database, auditLog)
			}
		}

		return err
	}
}

// Helper functions
func getCachedPermissions(c *fiber.Ctx, db *sql.DB, store *session.Store, userID int) []models.Permission {
	// Check if permissions are already cached in session
	sess, err := store.Get(c)
	if err != nil {
		return nil
	}

	// Try to get cached permissions
	if cached := sess.Get("user_permissions"); cached != nil {
		if permissions, ok := cached.([]models.Permission); ok {
			return permissions
		}
	}

	// If not cached, fetch from database
	permissions, err := models.GetUserPermissions(c.Context(), db, userID)
	if err != nil {
		return nil
	}

	// Cache permissions in session (with expiration)
	sess.Set("user_permissions", permissions)
	sess.Set("permissions_cached_at", time.Now())

	return permissions
}

func createAuditDetails(c *fiber.Ctx, duration time.Duration) string {
	details := map[string]interface{}{
		"method":     c.Method(),
		"path":       c.Path(),
		"status":     c.Response().StatusCode(),
		"duration":   duration.String(),
		"user_agent": c.Get("User-Agent"),
		"ip":         c.IP(),
	}

	// Add query parameters if any
	if len(c.Query("")) > 0 {
		details["query_params"] = c.Query("")
	}

	// Add request body for POST/PUT requests (be careful with sensitive data)
	if c.Method() == "POST" || c.Method() == "PUT" {
		// Only log non-sensitive endpoints
		path := c.Path()
		if !isSensitiveEndpoint(path) {
			body := string(c.Body())
			if len(body) > 0 && len(body) < 1000 { // Limit body size for logging
				details["request_body"] = body
			}
		}
	}

	jsonDetails, _ := json.Marshal(details)
	return string(jsonDetails)
}

func isSensitiveEndpoint(path string) bool {
	sensitivePaths := []string{
		"/login",
		"/users/password",
		"/auth",
		"/api/auth",
	}

	for _, sensitivePath := range sensitivePaths {
		if path == sensitivePath || len(path) >= len(sensitivePath) && path[:len(sensitivePath)] == sensitivePath {
			return true
		}
	}
	return false
}

// Utility function to get current user ID from context
func GetCurrentUserID(c *fiber.Ctx) (int, bool) {
	if userID := c.Locals("current_user_id"); userID != nil {
		if uid, ok := userID.(int); ok {
			return uid, true
		}
	}
	return 0, false
}

// Utility function to check if user has permission
func UserHasPermission(c *fiber.Ctx, resource, action string) bool {
	if permissions := c.Locals("current_user_permissions"); permissions != nil {
		if perms, ok := permissions.([]models.Permission); ok {
			for _, perm := range perms {
				if perm.Resource == resource && perm.Action == action {
					return true
				}
			}
		}
	}
	return false
}
