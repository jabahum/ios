package handlers

import (
	"database/sql"
	"errors"
	"log/slog"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/lib/pq"
)

// HandlerDepartmentListAPI returns all departments ordered by name.
func HandlerDepartmentListAPI(c *fiber.Ctx, db *sql.DB, sl *slog.Logger) error {
	rows, err := db.QueryContext(c.Context(), `
		SELECT id, name, description, is_active, created_at, updated_at
		FROM departments
		ORDER BY name`)
	if err != nil {
		sl.Error("department list", "error", err)
		return c.Status(500).JSON(fiber.Map{"error": "Database error"})
	}
	defer rows.Close()

	out := make([]fiber.Map, 0)
	for rows.Next() {
		var id int
		var name string
		var desc sql.NullString
		var active bool
		var createdAt, updatedAt time.Time
		if err := rows.Scan(&id, &name, &desc, &active, &createdAt, &updatedAt); err != nil {
			continue
		}
		out = append(out, fiber.Map{
			"id":          id,
			"name":        name,
			"description": desc.String,
			"is_active":   active,
			"created_at":  createdAt,
			"updated_at":  updatedAt,
		})
	}
	return c.JSON(fiber.Map{"departments": out})
}

// HandlerDepartmentGetAPI returns one department by id.
func HandlerDepartmentGetAPI(c *fiber.Ctx, db *sql.DB, sl *slog.Logger) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id <= 0 {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid department id"})
	}
	var name string
	var desc sql.NullString
	var active bool
	var createdAt, updatedAt time.Time
	err = db.QueryRowContext(c.Context(), `
		SELECT name, description, is_active, created_at, updated_at
		FROM departments WHERE id = $1`, id).Scan(&name, &desc, &active, &createdAt, &updatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return c.Status(404).JSON(fiber.Map{"error": "Department not found"})
	}
	if err != nil {
		sl.Error("department get", "error", err)
		return c.Status(500).JSON(fiber.Map{"error": "Database error"})
	}
	return c.JSON(fiber.Map{
		"department": fiber.Map{
			"id":          id,
			"name":        name,
			"description": desc.String,
			"is_active":   active,
			"created_at":  createdAt,
			"updated_at":  updatedAt,
		},
	})
}

// HandlerDepartmentCreateAPI inserts a department.
func HandlerDepartmentCreateAPI(c *fiber.Ctx, db *sql.DB, sl *slog.Logger) error {
	var body APIDepartmentWriteRequest
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid JSON"})
	}
	if body.Name == "" {
		return c.Status(400).JSON(fiber.Map{"error": "name is required"})
	}
	var newID int
	err := db.QueryRowContext(c.Context(), `
		INSERT INTO departments (name, description, is_active)
		VALUES ($1, $2, $3)
		RETURNING id`,
		body.Name, body.Description, body.IsActive).Scan(&newID)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return c.Status(400).JSON(fiber.Map{"error": "Department name already exists"})
		}
		sl.Error("department create", "error", err)
		return c.Status(500).JSON(fiber.Map{"error": "Failed to create department"})
	}
	return c.Status(201).JSON(fiber.Map{"message": "Department created", "id": newID})
}

// HandlerDepartmentUpdateAPI updates a department.
func HandlerDepartmentUpdateAPI(c *fiber.Ctx, db *sql.DB, sl *slog.Logger) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id <= 0 {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid department id"})
	}
	var body APIDepartmentWriteRequest
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid JSON"})
	}
	if body.Name == "" {
		return c.Status(400).JSON(fiber.Map{"error": "name is required"})
	}
	res, err := db.ExecContext(c.Context(), `
		UPDATE departments SET name = $1, description = $2, is_active = $3, updated_at = CURRENT_TIMESTAMP
		WHERE id = $4`,
		body.Name, body.Description, body.IsActive, id)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return c.Status(400).JSON(fiber.Map{"error": "Department name already exists"})
		}
		sl.Error("department update", "error", err)
		return c.Status(500).JSON(fiber.Map{"error": "Database error"})
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return c.Status(404).JSON(fiber.Map{"error": "Department not found"})
	}
	return c.JSON(fiber.Map{"message": "Department updated"})
}

// HandlerDepartmentDeleteAPI deletes a department (fails if referenced by users).
func HandlerDepartmentDeleteAPI(c *fiber.Ctx, db *sql.DB, sl *slog.Logger) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id <= 0 {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid department id"})
	}
	res, err := db.ExecContext(c.Context(), `DELETE FROM departments WHERE id = $1`, id)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23503" {
			return c.Status(409).JSON(fiber.Map{"error": "Department is still referenced (e.g. users.department_id)"})
		}
		sl.Error("department delete", "error", err)
		return c.Status(500).JSON(fiber.Map{"error": "Database error"})
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return c.Status(404).JSON(fiber.Map{"error": "Department not found"})
	}
	return c.JSON(fiber.Map{"message": "Department deleted"})
}
