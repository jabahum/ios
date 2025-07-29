package utils

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
)

// GetUserIDFromSession gets the user ID from the current session
func GetUserIDFromSession(c *fiber.Ctx) int64 {
	// Get the session store from Fiber locals
	store := c.Locals("store")
	if store == nil {
		return 0
	}

	sessionStore, ok := store.(*session.Store)
	if !ok {
		return 0
	}

	// Get session
	sess, err := sessionStore.Get(c)
	if err != nil {
		return 0
	}

	// Try to get user ID from session
	userID := sess.Get("user")
	if userID == nil {
		return 0
	}

	// Convert to int64
	switch v := userID.(type) {
	case int:
		return int64(v)
	case int64:
		return v
	case float64:
		return int64(v)
	default:
		return 0
	}
}
