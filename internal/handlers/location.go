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

// HandlerGetSubcounties returns subcounties for a given district
func HandlerGetSubcounties(c *fiber.Ctx, db *sql.DB, sl *slog.Logger) error {
	districtID, err := strconv.Atoi(c.Params("district_id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid district ID",
		})
	}

	subcounties, err := models.GetSubcountiesByDistrict(db, districtID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(subcounties)
}

// HandlerGetParishes returns parishes for a given subcounty
func HandlerGetParishes(c *fiber.Ctx, db *sql.DB, sl *slog.Logger) error {
	subcountyID, err := strconv.Atoi(c.Params("subcounty_id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid subcounty ID",
		})
	}

	parishes, err := models.GetParishesBySubcounty(db, subcountyID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(parishes)
}

// HandlerGetParishesByDistrict returns parishes for a given district
func HandlerGetParishesByDistrict(c *fiber.Ctx, db *sql.DB, sl *slog.Logger) error {
	districtID, err := strconv.Atoi(c.Params("district_id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid district ID",
		})
	}

	parishes, err := models.GetParishesByDistrict(db, districtID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(parishes)
}

// HandlerGetVillages returns villages for a given parish
func HandlerGetVillages(c *fiber.Ctx, db *sql.DB, sl *slog.Logger) error {
	parishID, err := strconv.Atoi(c.Params("parish_id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid parish ID",
		})
	}

	villages, err := models.GetVillagesByParish(db, parishID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(villages)
}

// HandlerGetVillagesByDistrict returns villages for a given district
func HandlerGetVillagesByDistrict(c *fiber.Ctx, db *sql.DB, sl *slog.Logger) error {
	districtID, err := strconv.Atoi(c.Params("district_id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid district ID",
		})
	}

	villages, err := models.GetVillagesByDistrict(db, districtID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(villages)
}

// HandlerGetVillagesBySubcounty returns villages for a given subcounty
func HandlerGetVillagesBySubcounty(c *fiber.Ctx, db *sql.DB, sl *slog.Logger) error {
	subcountyID, err := strconv.Atoi(c.Params("subcounty_id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid subcounty ID",
		})
	}

	villages, err := models.GetVillagesBySubcounty(db, subcountyID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(villages)
}
