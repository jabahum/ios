package utils

import (
	"github.com/gin-gonic/gin"
)

// GetUserIDFromSession gets the user ID from the current session
func GetUserIDFromSession(c *gin.Context) int64 {
	// This should be implemented based on your session management
	// For now, returning a placeholder
	if userID, exists := c.Get("user_id"); exists {
		if id, ok := userID.(int64); ok {
			return id
		}
	}
	return 0
}
