package handlers

import (
	"case/internal/config"
	"case/internal/models"
	"database/sql"
	"log/slog"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
)

func VerifyDischarge2(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config config.Config) error {
	//data := map[string]string{"Title": "Help Page"}
	//return GenerateHTML(c, data, "help")
	id, err := strconv.Atoi(c.Params("i"))

	if err != nil {
		sl.Error("No ID provided: " + err.Error())
		return c.Status(fiber.StatusOK).SendString("Invalid Discharge Certificate")
	}

	// Step 2: Fetch discharge record
	discharge, err := models.DischargeByDischargeID(c.Context(), db, id) // Assuming this function exists
	if err != nil {
		sl.Error("failed to retrieve discharge details: " + err.Error())
		return c.Status(fiber.StatusOK).SendString("Invalid Discharge Certificate")
	}

	if discharge.ClientID.Int64 > 0 {

		client, er := models.ClientByID(c.Context(), db, int(discharge.ClientID.Int64))
		if er != nil {
			sl.Error("failed to retrieve client details: " + er.Error())
			return c.Status(fiber.StatusOK).SendString("Invalid Discharge Certificate")
		}

		msg := "Valid Discharge Certificate \n" +
			"Patient #: " + client.EtuNo.String + "\n" +
			"Patient Name: " + client.Firstname.String + " " + client.Lastname.String + "\n" +
			"Doctor: " + discharge.DischargingOfficer.String + "\n" +
			"Date of Discharge: " + discharge.DischargeDate.String + "\n"

		return c.Status(fiber.StatusOK).SendString(msg)
	} else {
		sl.Error("No client associeted with provided discharge details")
		return c.Status(fiber.StatusOK).SendString("Invalid Discharge Certificate")
	}
}

func VerifyDischarge(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config config.Config) error {

	id, err := strconv.Atoi(c.Params("i"))

	if err != nil {
		sl.Error("failed to retrieve supplied ID: " + err.Error())
		return c.Status(fiber.StatusOK).JSON(fiber.Map{"error": "Invalid Discharge Certificate"})
	}

	// Step 2: Fetch discharge record
	discharge, err := models.DischargeByDischargeID(c.Context(), db, id) // Assuming this function exists
	if err != nil {
		sl.Error("failed to retrieve discharge details: " + err.Error())
		return c.Status(fiber.StatusOK).JSON(fiber.Map{"error": "Invalid Discharge Certificate"})
	}

	// Step 3: Check if discharge record exists
	if discharge == nil {
		sl.Error("no discharge details associated with id")
		return c.Status(fiber.StatusOK).JSON(fiber.Map{"error": "Invalide Discharge Certificate"})
	}

	client, er := models.ClientByID(c.Context(), db, int(discharge.ClientID.Int64))
	if er != nil {
		sl.Error("failed to retrieve client details: " + er.Error())
		return c.Status(fiber.StatusOK).JSON(fiber.Map{"error": "Invalid Discharge Certificate"})
	}

	// Step 4: Return confirmation message
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":        "Valid Discharge Certificate",
		"Patient":        client.Firstname.String + " " + client.Lastname.String,
		"Patient #":      client.EtuNo.String,
		"Discharge Date": discharge.DischargeDate.String,
	})

}
