package middleware

import (
	"github.com/gin-gonic/gin"
)

// GinAuthRequired creates a Gin middleware that requires authentication
func GinAuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		// For now, just pass through - you can implement actual auth logic here
		// This is a placeholder that allows all requests through
		c.Next()
	}
}
