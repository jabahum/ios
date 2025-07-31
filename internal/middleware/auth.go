package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
)

// AuthRequired is a middleware that checks if the user is authenticated
func AuthRequired(store *session.Store) fiber.Handler {
	return func(c *fiber.Ctx) error {
		sess, err := store.Get(c)
		if err != nil {
			return c.Redirect("/login")
		}

		// Check if user is authenticated
		if sess.Get("isAuthenticated") != true {
			return c.Redirect("/login")
		}

		return c.Next()
	}
}
