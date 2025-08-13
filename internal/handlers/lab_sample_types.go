package handlers

import (
	"database/sql"
	"log/slog"
	"strconv"

	"case/internal/models"

	"github.com/gofiber/fiber/v2"
)

// HandlerAPIGetLabSwabTypes returns all swab types for dropdown
func HandlerAPIGetLabSwabTypes(c *fiber.Ctx, db *sql.DB, sl *slog.Logger) error {
	swabTypes, err := models.LabSwabTypes(c.Context(), db)
	if err != nil {
		sl.Error("Failed to get swab types", "error", err)
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to get swab types",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    swabTypes,
	})
}

// HandlerAPIGetLabUrineTypes returns all urine test types for dropdown
func HandlerAPIGetLabUrineTypes(c *fiber.Ctx, db *sql.DB, sl *slog.Logger) error {
	urineTypes, err := models.LabUrineTypes(c.Context(), db)
	if err != nil {
		sl.Error("Failed to get urine types", "error", err)
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to get urine types",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    urineTypes,
	})
}

// HandlerAPIGetLabBloodTypes returns all blood test types for dropdown
func HandlerAPIGetLabBloodTypes(c *fiber.Ctx, db *sql.DB, sl *slog.Logger) error {
	bloodTypes, err := models.LabBloodTypes(c.Context(), db)
	if err != nil {
		sl.Error("Failed to get blood types", "error", err)
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to get blood types",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    bloodTypes,
	})
}

// HandlerAPIGetLabBloodTypesByCategory returns blood test types by category
func HandlerAPIGetLabBloodTypesByCategory(c *fiber.Ctx, db *sql.DB, sl *slog.Logger) error {
	category := c.Params("category")
	if category == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Category parameter is required",
		})
	}

	bloodTypes, err := models.LabBloodTypesByCategory(c.Context(), db, category)
	if err != nil {
		sl.Error("Failed to get blood types by category", "error", err, "category", category)
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to get blood types by category",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    bloodTypes,
	})
}

// HandlerAPISaveLabSampleSelections saves lab sample selections
func HandlerAPISaveLabSampleSelections(c *fiber.Ctx, db *sql.DB, sl *slog.Logger) error {
	var request struct {
		LabID      int    `json:"lab_id"`
		SampleType string `json:"sample_type"`
		Selections []struct {
			SelectedTypeID int    `json:"selected_type_id"`
			OtherSpecify   string `json:"other_specify,omitempty"`
		} `json:"selections"`
	}

	if err := c.BodyParser(&request); err != nil {
		sl.Error("Failed to parse request body", "error", err)
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Delete existing selections for this lab and sample type
	_, err := db.ExecContext(c.Context(),
		"DELETE FROM public.lab_sample_selections WHERE lab_id = $1 AND sample_type = $2",
		request.LabID, request.SampleType)
	if err != nil {
		sl.Error("Failed to delete existing selections", "error", err)
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to save selections",
		})
	}

	// Insert new selections
	for _, selection := range request.Selections {
		sampleSelection := &models.LabSampleSelection{
			LabID:          request.LabID,
			SampleType:     request.SampleType,
			SelectedTypeID: sql.NullInt64{Int64: int64(selection.SelectedTypeID), Valid: true},
			OtherSpecify:   sql.NullString{String: selection.OtherSpecify, Valid: selection.OtherSpecify != ""},
		}

		if err := sampleSelection.Insert(c.Context(), db); err != nil {
			sl.Error("Failed to insert sample selection", "error", err)
			return c.Status(500).JSON(fiber.Map{
				"error": "Failed to save selections",
			})
		}
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Sample selections saved successfully",
	})
}

// HandlerAPIGetLabSampleSelections returns sample selections for a lab
func HandlerAPIGetLabSampleSelections(c *fiber.Ctx, db *sql.DB, sl *slog.Logger) error {
	labIDStr := c.Params("lab_id")
	if labIDStr == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Lab ID parameter is required",
		})
	}

	labID, err := strconv.Atoi(labIDStr)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid lab ID",
		})
	}

	selections, err := models.LabSampleSelectionsByLabID(c.Context(), db, labID)
	if err != nil {
		sl.Error("Failed to get sample selections", "error", err)
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to get sample selections",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    selections,
	})
}
