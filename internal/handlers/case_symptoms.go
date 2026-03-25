package handlers

import (
	"case/internal/config"
	"database/sql"
	"log/slog"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
)

func HandlerSymptomsForm(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config config.Config) error {
	return nil
}
func HandlerSymptomsSubmit(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config config.Config) error {
	return nil
}
func HandlerSymptomsList(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config config.Config) error {
	return nil
}
