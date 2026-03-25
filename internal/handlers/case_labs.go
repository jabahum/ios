package handlers

import (
	"case/internal/config"
	"database/sql"
	"log/slog"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
)

func HandlerLabForm(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config config.Config) error {
	return nil
}
func HandlerLabSubmit(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config config.Config) error {
	return nil
}
func HandlerLabList(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config config.Config) error {
	return nil
}
