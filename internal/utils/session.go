package utils

import (
	"github.com/gofiber/fiber/v2"
)

// GetUserIDFromSession gets the user ID from the current session
func GetUserIDFromSession(c *fiber.Ctx) int64 {
	// This should be implemented based on your session management
	// For now, returning a placeholder
	if userID := c.Locals("user_id"); userID != nil {
		if id, ok := userID.(int64); ok {
			return id
		}
	}
	return 0
}
