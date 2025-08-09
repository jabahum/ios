package handlers

import (
	"case/internal/models"
	"database/sql"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/lib/pq"
)

func HandlerMpoxDailyFollowUpForm(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	clientIDStr := c.Params("i")
	clientID, err := strconv.Atoi(clientIDStr)
	if err != nil {
		return c.Status(http.StatusBadRequest).SendString("Invalid client ID")
	}

	// Validate that the client exists and check facility access
	var clientSite sql.NullInt64
	err = db.QueryRow("SELECT site FROM clients WHERE id = $1", clientID).Scan(&clientSite)
	if err != nil {
		sl.Error("Failed to check client existence", "error", err, "client_id", clientID)
		return c.Status(500).SendString("Failed to get client details")
	}

	if !clientSite.Valid {
		sl.Error("Client not found", "client_id", clientID)
		return c.Status(404).SendString("Client not found")
	}

	// Check facility-based access control
	userID := GetCurrentUser(c, store)
	userFacility := GetCurrentFacility(c, db, sl, store)
	if userFacility > 0 {
		// User has a facility assigned, check if they can access this case
		if clientSite.Int64 != int64(userFacility) {
			sl.Error("User attempted to access mpox daily follow-up for case from different facility",
				"user_id", userID, "user_facility", userFacility, "case_site", clientSite.Int64, "case_id", clientID)
			return c.Status(403).SendString("Access denied: You can only access cases from your assigned facility.")
		}
	}

	dateStr := c.Query("dte")
	var encounterDate time.Time
	if dateStr == "" || dateStr == "0000-00-00" {
		encounterDate = time.Now()
	} else {
		encounterDate, _ = time.Parse("2006-01-02", dateStr)
	}

	data := NewTemplateData(c, store)
	data.Title = "Mpox Daily Follow-Up Form"
	data.Form = fiber.Map{
		"ClientID":      clientID,
		"EncounterDate": encounterDate.Format("2006-01-02"),
	}
	return GenerateHTML(c, db, data, "mpox_daily_followup_new")
}

func HandlerMpoxDailyFollowUpSubmit(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	clientID, err := strconv.Atoi(c.FormValue("client_id"))
	if err != nil {
		sl.Error("Invalid client ID", "error", err, "client_id", c.FormValue("client_id"))
		return c.Status(400).SendString("Invalid client ID")
	}

	// Validate that the client exists and check facility access
	var clientSite sql.NullInt64
	err = db.QueryRow("SELECT site FROM clients WHERE id = $1", clientID).Scan(&clientSite)
	if err != nil {
		sl.Error("Failed to check client existence", "error", err, "client_id", clientID)
		return c.Status(500).SendString("Failed to get client details")
	}

	if !clientSite.Valid {
		sl.Error("Client not found", "client_id", clientID)
		return c.Status(404).SendString("Client not found")
	}

	// Check facility-based access control
	userID := GetCurrentUser(c, store)
	userFacility := GetCurrentFacility(c, db, sl, store)
	if userFacility > 0 {
		// User has a facility assigned, check if they can access this case
		if clientSite.Int64 != int64(userFacility) {
			sl.Error("User attempted to submit mpox daily follow-up for case from different facility",
				"user_id", userID, "user_facility", userFacility, "case_site", clientSite.Int64, "case_id", clientID)
			return c.Status(403).SendString("Access denied: You can only access cases from your assigned facility.")
		}
	}

	dateStr := c.FormValue("followup_date")
	var followupDate sql.NullTime
	if dateStr != "" {
		dt, err := time.Parse("2006-01-02", dateStr)
		if err == nil {
			followupDate = sql.NullTime{Time: dt, Valid: true}
		}
	}
	// Handle encounter_type as array from multiple select
	var encounterTypeArray pq.StringArray
	// Get all form values for encounter_type (multiple select)
	form, err := c.MultipartForm()
	if err == nil && form != nil {
		if values, exists := form.Value["encounter_type"]; exists {
			encounterTypeArray = pq.StringArray(values)
		}
	}
	// Fallback to single value if multipart form fails
	if len(encounterTypeArray) == 0 {
		if singleValue := c.FormValue("encounter_type"); singleValue != "" {
			encounterTypeArray = pq.StringArray{singleValue}
		}
	}
	followup := &models.MpoxDailyFollowUp{
		ClientID:           clientID,
		FollowUpDate:       followupDate,
		EncounterType:      encounterTypeArray,
		OtherSite:          parseNullString(c.FormValue("other_site")),
		SpO2:               parseNullInt64(c.FormValue("spo2")),
		NewLesions:         parseNullBool(c.FormValue("new_lesions")),
		DiseaseProgression: parseNullString(c.FormValue("disease_progression")),
		ProgressionOther:   parseNullString(c.FormValue("progression_other")),
		LesionFace:         parseNullString(c.FormValue("lesion_face")),
		LesionMouth:        parseNullString(c.FormValue("lesion_mouth")),
		LesionChest:        parseNullString(c.FormValue("lesion_chest")),
		LesionAbdomen:      parseNullString(c.FormValue("lesion_abdomen")),
		LesionBack:         parseNullString(c.FormValue("lesion_back")),
		LesionArms:         parseNullString(c.FormValue("lesion_arms")),
		LesionPalms:        parseNullString(c.FormValue("lesion_palms")),
		LesionForearms:     parseNullString(c.FormValue("lesion_forearms")),
		LesionThighs:       parseNullString(c.FormValue("lesion_thighs")),
		LesionLegs:         parseNullString(c.FormValue("lesion_legs")),
		LesionSoles:        parseNullString(c.FormValue("lesion_soles")),
		LesionGenitalia:    parseNullString(c.FormValue("lesion_genitalia")),
		LesionPerianal:     parseNullString(c.FormValue("lesion_perianal")),
		LesionOther:        parseNullString(c.FormValue("lesion_other")),
		LesionSpecifyWhere: parseNullString(c.FormValue("lesion_specify_where")),
		TypeMacule:         parseNullString(c.FormValue("type_macule")),
		TypePapule:         parseNullString(c.FormValue("type_papule")),
		TypeVesicle:        parseNullString(c.FormValue("type_vesicle")),
		TypePustule:        parseNullString(c.FormValue("type_pustule")),
		TypeUmbilicated:    parseNullString(c.FormValue("type_umbilicated")),
		TypeUlcerated:      parseNullString(c.FormValue("type_ulcerated")),
		TypeCrusting:       parseNullString(c.FormValue("type_crusting")),
		TypePartialScab:    parseNullString(c.FormValue("type_partialscab")),
		TypeOther:          parseNullString(c.FormValue("type_other")),
		PainPresent:        parseNullBool(c.FormValue("pain_present")),
		PainScore:          parseNullInt64(c.FormValue("pain_score")),
		PainDescription:    parseNullString(c.FormValue("pain_description")),
		Temperature:        parseNullFloat64(c.FormValue("temperature")),
		HeartRate:          parseNullInt64(c.FormValue("heart_rate")),
		RespiratoryRate:    parseNullInt64(c.FormValue("respiratory_rate")),
		BpSystolic:         parseNullInt64(c.FormValue("bp_systolic")),
		BpDiastolic:        parseNullInt64(c.FormValue("bp_diastolic")),
		Consciousness:      parseNullString(c.FormValue("consciousness")),
		DataEntrant:        parseNullString(c.FormValue("data_entrant")),
	}

	if err := followup.Insert(db); err != nil {
		sl.Error("Failed to save follow-up", "error", err)
		return c.Status(500).SendString("Failed to save follow-up")
	}

	return c.Redirect("/cases/encounters/list/" + strconv.Itoa(clientID))
}

// Helper parse functions
func parseNullFloat64(val string) sql.NullFloat64 {
	f, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return sql.NullFloat64{Valid: false}
	}
	return sql.NullFloat64{Float64: f, Valid: true}
}

func parseNullInt64(val string) sql.NullInt64 {
	i, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return sql.NullInt64{Valid: false}
	}
	return sql.NullInt64{Int64: i, Valid: true}
}

func parseNullString(val string) sql.NullString {
	if val == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: val, Valid: true}
}

func parseNullBool(val string) sql.NullBool {
	if val == "on" || val == "true" || val == "1" {
		return sql.NullBool{Bool: true, Valid: true}
	}
	if val == "off" || val == "false" || val == "0" {
		return sql.NullBool{Bool: false, Valid: true}
	}
	return sql.NullBool{Valid: false}
}

// Helper to join string slices
func joinStrings(arr []string, sep string) string {
	result := ""
	for i, s := range arr {
		if i > 0 {
			result += sep
		}
		result += s
	}
	return result
}
