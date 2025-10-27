package handlers

import (
	"database/sql"
	"log/slog"
	"strconv"
	"fmt"

	"case/internal/models"

	"github.com/gofiber/fiber/v2"
)

// HandlerGetSurveys returns all surveys
func HandlerGetSurveys(c *fiber.Ctx, db *sql.DB, sl *slog.Logger) error {
	fmt.Println("here!!!!!")
	surveys, err := models.GetSurveys(db)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(surveys)
}

// HandlerGetQuestionsBySurvey returns survey_questions for a given survey
func HandlerGetQuestionsBySurvey(c *fiber.Ctx, db *sql.DB, sl *slog.Logger) error {
	surveyIDStr := c.Params("survey_id")
	if surveyIDStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Survey ID is required",
		})
	}

	surveyID, err := strconv.Atoi(surveyIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid survey ID",
		})
	}

	survey_questions, err := models.GetQuestionsBySurvey(db, surveyID)
	if err != nil {
		sl.Error("Error getting survey_questions", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get survey_questions",
		})
	}

	return c.JSON(survey_questions)
}
