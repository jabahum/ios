from pathlib import Path

root = Path(__file__).resolve().parents[1] / "internal" / "routes"
body = (root / "app_json_api_routes_body.go.txt").read_text(encoding="utf-8")
header = r"""package routes

import (
	"database/sql"
	"log/slog"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"

	"case/internal/handlers"
	"case/internal/models"
	"case/internal/services"
)

// JSONAPIAuthRequired returns 401 JSON when the session is not authenticated (no HTML redirect).
// Accepts session keys "user" or "user_id" as int, int64, or float64.
func JSONAPIAuthRequired(store *session.Store) fiber.Handler {
	return func(c *fiber.Ctx) error {
		sess, err := store.Get(c)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Authentication required"})
		}
		if sess.Get("isAuthenticated") != true {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Authentication required"})
		}
		if apiSessionUserID(sess) == 0 {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid user session"})
		}
		return c.Next()
	}
}

func apiSessionUserID(sess *session.Session) int {
	for _, key := range []string{"user_id", "user"} {
		v := sess.Get(key)
		if v == nil {
			continue
		}
		switch n := v.(type) {
		case int:
			if n > 0 {
				return n
			}
		case int64:
			if n > 0 {
				return int(n)
			}
		case float64:
			if n > 0 {
				return int(n)
			}
		}
	}
	return 0
}

// registerAuthenticatedJSONAPIRoutes registers JSON /api/* routes on the root app with session auth.
// This avoids relying only on app.Group("/", AuthRequired(store)), which can behave differently per deploy
// and only checks sess.Get("user") (not user_id).
func registerAuthenticatedJSONAPIRoutes(app *fiber.App, db *sql.DB, sl *slog.Logger, store *session.Store, config handlers.Config, smsService *services.SMSService, inventoryHandler *handlers.InventoryHandler) {
	auth := JSONAPIAuthRequired(store)
"""
(root / "app_json_api.go").write_text(header + "\n" + body + "\n}\n", encoding="utf-8")
print("OK", root / "app_json_api.go")
