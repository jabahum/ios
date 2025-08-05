package handlers

import (
	"case/internal/models"
	"database/sql"
	"log/slog"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
)

// HandlerOutbreakList handles the outbreak list page
func HandlerOutbreakList(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	// Get current user ID
	userID := GetCurrentUser(c, store)
	if userID == 0 {
		return c.Redirect("/login")
	}

	// Get all active outbreaks
	outbreaks, err := models.GetActiveOutbreaks(c.Context(), db)
	if err != nil {
		sl.Error("Failed to get active outbreaks: " + err.Error())
		return c.Status(500).SendString("Failed to get outbreaks")
	}

	// Get default outbreak for this user
	var defaultOutbreak *models.Outbreak
	if len(outbreaks) > 0 {
		defaultOutbreak = outbreaks[0]
	} else {
		// If no accessible outbreaks, create a default one or redirect
		sl.Error("No outbreaks accessible to user", "user_id", userID)
		return c.Status(403).SendString("No outbreaks accessible to your account. Please contact your administrator.")
	}

	// Convert outbreaks to interface slice
	items := make([]interface{}, len(outbreaks))
	for i, v := range outbreaks {
		items[i] = v
	}

	data := NewTemplateData(c, store)
	data.Items = items
	data.Form = defaultOutbreak

	// Check if user has case-related roles
	hasCaseRole, err := hasCaseRole(c, db, userID)
	if err != nil {
		sl.Error("Failed to check case role", "user_id", userID, "error", err)
		hasCaseRole = false
	}

	// Add user management permissions to data
	canManageOutbreak := make(map[int]bool)
	for _, outbreak := range outbreaks {
		canManage, err := models.CanUserManageOutbreak(c.Context(), db, userID, outbreak.ID)
		if err != nil {
			sl.Error("Failed to check outbreak management permission", "outbreak_id", outbreak.ID, "error", err)
			canManage = false
		}
		canManageOutbreak[outbreak.ID] = canManage
	}
	data.Optionz = map[string]map[string]string{
		"can_manage_outbreak": make(map[string]string),
		"has_case_role":       make(map[string]string),
	}
	for outbreakID, canManage := range canManageOutbreak {
		data.Optionz["can_manage_outbreak"][strconv.Itoa(outbreakID)] = strconv.FormatBool(canManage)
	}

	// Add case role flag
	data.Optionz["has_case_role"]["value"] = strconv.FormatBool(hasCaseRole)

	// Debug logging
	sl.Info("Template data debug",
		"optionz_initialized", data.Optionz != nil,
		"optionz_keys", len(data.Optionz),
		"has_case_role_value", data.Optionz["has_case_role"]["value"])

	return GenerateHTML(c, db, data, "outbreaks")
}

// HandlerOutbreakForm handles the outbreak form page
func HandlerOutbreakForm(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	// Get current user ID
	userID := GetCurrentUser(c, store)
	if userID == 0 {
		return c.Redirect("/login")
	}

	var outbreak *models.Outbreak
	id := c.Params("id")
	if id != "0" {
		idInt, err := strconv.Atoi(id)
		if err != nil {
			sl.Error("Invalid outbreak ID: " + err.Error())
			return c.Status(400).SendString("Invalid outbreak ID")
		}

		// Check if user has access to this outbreak
		hasAccess, err := models.CheckUserOutbreakAccess(c.Context(), db, userID, idInt)
		if err != nil {
			sl.Error("Failed to check outbreak access: " + err.Error())
			return c.Status(500).SendString("Failed to check access")
		}
		if !hasAccess {
			return c.Status(403).SendString("Access denied to this outbreak")
		}

		outbreak, err = models.OutbreakByID(c.Context(), db, idInt)
		if err != nil {
			sl.Error("Failed to get outbreak: " + err.Error())
			return c.Status(500).SendString("Failed to get outbreak")
		}
	} else {
		outbreak = &models.Outbreak{}
	}

	data := NewTemplateData(c, store)
	data.Form = outbreak

	// Add outbreak type options
	data.Optionz = map[string]map[string]string{
		"outbreak_types": {
			"vhf":     "VHF (Viral Hemorrhagic Fever)",
			"mpox":    "MPOX (Monkeypox)",
			"general": "General",
		},
	}

	return GenerateHTML(c, db, data, "form_outbreak")
}

// HandlerOutbreakSubmit handles the outbreak form submission
func HandlerOutbreakSubmit(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	// Get current user ID
	userID := GetCurrentUser(c, store)
	if userID == 0 {
		return c.Redirect("/login")
	}

	var outbreak models.Outbreak
	if err := DecodeFormData(c, &outbreak); err != nil {
		sl.Error("Failed to decode form data: " + err.Error())
		return c.Status(400).SendString("Invalid form data")
	}

	// If editing existing outbreak, check access
	if outbreak.ID != 0 {
		hasAccess, err := models.CheckUserOutbreakAccess(c.Context(), db, userID, outbreak.ID)
		if err != nil {
			sl.Error("Failed to check outbreak access: " + err.Error())
			return c.Status(500).SendString("Failed to check access")
		}
		if !hasAccess {
			return c.Status(403).SendString("Access denied to this outbreak")
		}

		// Check if user can manage this outbreak
		canManage, err := models.CanUserManageOutbreak(c.Context(), db, userID, outbreak.ID)
		if err != nil {
			sl.Error("Failed to check outbreak management permission: " + err.Error())
			return c.Status(500).SendString("Failed to check permissions")
		}
		if !canManage {
			return c.Status(403).SendString("You don't have permission to edit this outbreak")
		}
	}

	// Parse start date
	startDate, err := time.Parse("2006-01-02", c.FormValue("start_date"))
	if err != nil {
		sl.Error("Failed to parse start date: " + err.Error())
		return c.Status(400).SendString("Invalid start date format")
	}
	outbreak.StartDate = sql.NullTime{Time: startDate, Valid: true}

	// Set outbreak type and category
	outbreakType := c.FormValue("outbreak_type")
	if outbreakType != "" {
		outbreak.OutbreakType = sql.NullString{String: outbreakType, Valid: true}
		outbreak.OutbreakCategory = sql.NullString{String: outbreakType, Valid: true}
	}

	// Set audit fields
	now := time.Now()
	if !outbreak.Exists() {
		outbreak.EnterOn = sql.NullTime{Time: now, Valid: true}
		outbreak.EnterBy = sql.NullInt64{Int64: int64(userID), Valid: true}
	}
	outbreak.EditOn = sql.NullTime{Time: now, Valid: true}
	outbreak.EditBy = sql.NullInt64{Int64: int64(userID), Valid: true}

	// Save outbreak
	if err := outbreak.Save(c.Context(), db); err != nil {
		sl.Error("Failed to save outbreak: " + err.Error())
		return c.Status(500).SendString("Failed to save outbreak")
	}

	return c.Redirect("/outbreaks")
}

// HandlerOutbreakClose handles closing an outbreak
func HandlerOutbreakClose(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	// Get current user ID
	userID := GetCurrentUser(c, store)
	if userID == 0 {
		return c.Redirect("/login")
	}

	id, err := strconv.Atoi(c.Params("i"))
	if err != nil {
		sl.Error("Invalid outbreak ID: " + err.Error())
		return c.Status(400).SendString("Invalid outbreak ID")
	}

	// Check if user has access to this outbreak
	hasAccess, err := models.CheckUserOutbreakAccess(c.Context(), db, userID, id)
	if err != nil {
		sl.Error("Failed to check outbreak access: " + err.Error())
		return c.Status(500).SendString("Failed to check access")
	}
	if !hasAccess {
		return c.Status(403).SendString("Access denied to this outbreak")
	}

	// Check if user can manage this outbreak
	canManage, err := models.CanUserManageOutbreak(c.Context(), db, userID, id)
	if err != nil {
		sl.Error("Failed to check outbreak management permission: " + err.Error())
		return c.Status(500).SendString("Failed to check permissions")
	}
	if !canManage {
		return c.Status(403).SendString("You don't have permission to close this outbreak")
	}

	// Get the outbreak
	outbreak, err := models.OutbreakByID(c.Context(), db, id)
	if err != nil {
		sl.Error("Failed to get outbreak: " + err.Error())
		return c.Status(500).SendString("Failed to get outbreak")
	}

	// Close the outbreak
	outbreak.Status = sql.NullString{String: "closed", Valid: true}
	outbreak.EditOn = sql.NullTime{Time: time.Now(), Valid: true}
	outbreak.EditBy = sql.NullInt64{Int64: int64(userID), Valid: true}

	if err := outbreak.Save(c.Context(), db); err != nil {
		sl.Error("Failed to close outbreak: " + err.Error())
		return c.Status(500).SendString("Failed to close outbreak")
	}

	return c.Redirect("/outbreaks")
}

// HandlerOutbreakSelect handles selecting an outbreak
func HandlerOutbreakSelect(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	// Get current user ID
	userID := GetCurrentUser(c, store)
	if userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	id, err := strconv.Atoi(c.Params("i"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid outbreak ID"})
	}

	// Check if user has access to this outbreak
	hasAccess, err := models.CheckUserOutbreakAccess(c.Context(), db, userID, id)
	if err != nil {
		sl.Error("Failed to check outbreak access: " + err.Error())
		return c.Status(500).JSON(fiber.Map{"error": "Failed to check access"})
	}
	if !hasAccess {
		return c.Status(403).JSON(fiber.Map{"error": "Access denied to this outbreak"})
	}

	// Set the selected outbreak in session
	if err := SetSelectedOutbreak(c, store, id); err != nil {
		sl.Error("Failed to set selected outbreak: " + err.Error())
		return c.Status(500).JSON(fiber.Map{"error": "Failed to select outbreak"})
	}

	return c.JSON(fiber.Map{"message": "Outbreak selected successfully"})
}

// HandlerGetOutbreaksAPI handles the API endpoint for getting all outbreaks
func HandlerGetOutbreaksAPI(c *fiber.Ctx, db *sql.DB, sl *slog.Logger) error {
	// Get all active outbreaks
	outbreaks, err := models.GetActiveOutbreaks(c.Context(), db)
	if err != nil {
		sl.Error("Failed to get active outbreaks: " + err.Error())
		return c.Status(500).JSON(fiber.Map{"error": "Failed to get outbreaks"})
	}

	// Convert to JSON response
	var response []fiber.Map
	for _, outbreak := range outbreaks {
		response = append(response, fiber.Map{
			"id":                outbreak.ID,
			"name":              outbreak.Name,
			"description":       outbreak.Description,
			"start_date":        outbreak.StartDate,
			"end_date":          outbreak.EndDate,
			"status":            outbreak.Status,
			"outbreak_type":     outbreak.OutbreakType,
			"outbreak_category": outbreak.OutbreakCategory,
		})
	}

	return c.JSON(response)
}

// SetSelectedOutbreak sets the selected outbreak in the session
func SetSelectedOutbreak(c *fiber.Ctx, store *session.Store, outbreakID int) error {
	sess, err := store.Get(c)
	if err != nil {
		return err
	}
	sess.Set("selected_outbreak", outbreakID)
	sess.Set("outbreak_id", outbreakID)
	return sess.Save()
}

// GetSelectedOutbreak gets the selected outbreak from the session
func GetSelectedOutbreak(c *fiber.Ctx, store *session.Store) int {
	sess, err := store.Get(c)
	if err != nil {
		return 0
	}
	// Try both keys
	for _, key := range []string{"selected_outbreak", "outbreak_id"} {
		val := sess.Get(key)
		switch v := val.(type) {
		case int:
			return v
		case int64:
			return int(v)
		case float64:
			return int(v)
		case string:
			i, _ := strconv.Atoi(v)
			return i
		}
	}
	return 0
}
