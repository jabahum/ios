package handlers

import (
	"database/sql"
	"log/slog"
	"strconv"

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
	districtIDStr := c.Params("district_id")
	if districtIDStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "District ID is required",
		})
	}

	districtID, err := strconv.Atoi(districtIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid district ID",
		})
	}

	subcounties, err := models.GetSubcountiesByDistrict(db, districtID)
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
	subcountyIDStr := c.Params("subcounty_id")
	if subcountyIDStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Subcounty ID is required",
		})
	}

	subcountyID, err := strconv.Atoi(subcountyIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid subcounty ID",
		})
	}

	parishes, err := models.GetParishesBySubcounty(db, subcountyID)
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
	districtIDStr := c.Params("district_id")
	if districtIDStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "District ID is required",
		})
	}

	districtID, err := strconv.Atoi(districtIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid district ID",
		})
	}

	parishes, err := models.GetParishesByDistrict(db, districtID)
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
	parishIDStr := c.Params("parish_id")
	if parishIDStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Parish ID is required",
		})
	}

	parishID, err := strconv.Atoi(parishIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid parish ID",
		})
	}

	villages, err := models.GetVillagesByParish(db, parishID)
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
	districtIDStr := c.Params("district_id")
	if districtIDStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "District ID is required",
		})
	}

	districtID, err := strconv.Atoi(districtIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid district ID",
		})
	}

	villages, err := models.GetVillagesByDistrict(db, districtID)
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
	subcountyIDStr := c.Params("subcounty_id")
	if subcountyIDStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Subcounty ID is required",
		})
	}

	subcountyID, err := strconv.Atoi(subcountyIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid subcounty ID",
		})
	}

	villages, err := models.GetVillagesBySubcounty(db, subcountyID)
	if err != nil {
		sl.Error("Error getting villages", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get villages",
		})
	}

	return c.JSON(villages)
}

// HandlerGetFacilities returns all facilities
func HandlerGetFacilities(c *fiber.Ctx, db *sql.DB, sl *slog.Logger) error {
	facilities, err := models.GetAllFacilities(db)
	if err != nil {
		sl.Error("Error getting facilities", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get facilities",
		})
	}

	return c.JSON(facilities)
}
