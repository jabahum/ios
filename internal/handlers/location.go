package handlers

import (
	"database/sql"
	"log/slog"

	"case/internal/models"

	"github.com/gofiber/fiber/v2"
)

// HandlerGetDistricts returns all districts
func HandlerGetDistricts(c *fiber.Ctx, db *sql.DB, sl *slog.Logger) error {
	districts, err := models.GetDistricts(db)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(districts)
}

// HandlerGetSubcountiesByDistrict returns subcounties for a given district
func HandlerGetSubcountiesByDistrict(c *fiber.Ctx, db *sql.DB, sl *slog.Logger) error {
	districtCode := c.Params("district_code")
	if districtCode == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "District code is required",
		})
	}

	subcounties, err := models.GetSubcountiesByDistrict(db, districtCode)
	if err != nil {
		sl.Error("Error getting subcounties", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get subcounties",
		})
	}

	return c.JSON(subcounties)
}

// HandlerGetParishesBySubcounty returns parishes for a given subcounty
func HandlerGetParishesBySubcounty(c *fiber.Ctx, db *sql.DB, sl *slog.Logger) error {
	subcountyCode := c.Params("subcounty_code")
	if subcountyCode == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Subcounty code is required",
		})
	}

	parishes, err := models.GetParishesBySubcounty(db, subcountyCode)
	if err != nil {
		sl.Error("Error getting parishes", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get parishes",
		})
	}

	return c.JSON(parishes)
}

// HandlerGetParishesByDistrict returns parishes for a given district
func HandlerGetParishesByDistrict(c *fiber.Ctx, db *sql.DB, sl *slog.Logger) error {
	districtCode := c.Params("district_code")
	if districtCode == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "District code is required",
		})
	}

	parishes, err := models.GetParishesByDistrict(db, districtCode)
	if err != nil {
		sl.Error("Error getting parishes", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get parishes",
		})
	}

	return c.JSON(parishes)
}

// HandlerGetVillagesByParish returns villages for a given parish
func HandlerGetVillagesByParish(c *fiber.Ctx, db *sql.DB, sl *slog.Logger) error {
	parishCode := c.Params("parish_code")
	if parishCode == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Parish code is required",
		})
	}

	villages, err := models.GetVillagesByParish(db, parishCode)
	if err != nil {
		sl.Error("Error getting villages", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get villages",
		})
	}

	return c.JSON(villages)
}

// HandlerGetVillagesByDistrict returns villages for a given district
func HandlerGetVillagesByDistrict(c *fiber.Ctx, db *sql.DB, sl *slog.Logger) error {
	districtCode := c.Params("district_code")
	if districtCode == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "District code is required",
		})
	}

	villages, err := models.GetVillagesByDistrict(db, districtCode)
	if err != nil {
		sl.Error("Error getting villages", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get villages",
		})
	}

	return c.JSON(villages)
}

// HandlerGetVillagesBySubcounty returns villages for a given subcounty
func HandlerGetVillagesBySubcounty(c *fiber.Ctx, db *sql.DB, sl *slog.Logger) error {
	subcountyCode := c.Params("subcounty_code")
	if subcountyCode == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Subcounty code is required",
		})
	}

	villages, err := models.GetVillagesBySubcounty(db, subcountyCode)
	if err != nil {
		sl.Error("Error getting villages", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get villages",
		})
	}

	return c.JSON(villages)
}
