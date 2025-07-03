package handlers

import (
	"case/internal/models"
	"case/internal/utils"
	"database/sql"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
)

// HandlerMpoxAdmissionForm renders the mpox admission form
func HandlerMpoxAdmissionForm(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	clientIDStr := c.Params("i")
	clientID, err := strconv.Atoi(clientIDStr)
	if err != nil {
		return c.Status(http.StatusBadRequest).SendString("Invalid client ID")
	}

	dateStr := c.Query("dte")
	if dateStr == "" || dateStr == "0000-00-00" {
		dateStr = time.Now().Format("2006-01-02")
	}

	// Get client details
	var client models.Client
	err = db.QueryRow("SELECT id, firstname, lastname FROM clients WHERE id = $1", clientID).Scan(&client.ID, &client.Firstname, &client.Lastname)
	if err != nil {
		sl.Error("Error fetching client", "error", err)
		return c.Status(http.StatusInternalServerError).SendString("Error fetching client details")
	}

	// Check if an admission exists for this client
	var admissionID int
	err = db.QueryRow("SELECT id FROM mpox_demographics WHERE client_id = $1", clientID).Scan(&admissionID)
	hasAdmission := err == nil

	// Get options for dropdowns
	options := map[string]map[string]string{
		"sex": {
			"M": "Male",
			"F": "Female",
		},
		"yn": {
			"Y": "Yes",
			"N": "No",
			"U": "Unknown",
		},
		"ppe_status": {
			"FULL":    "Full PPE",
			"PARTIAL": "Partial PPE",
			"NONE":    "No PPE",
		},
		"sex_of_partners": {
			"M": "Male",
			"F": "Female",
			"B": "Both",
			"N": "None",
		},
		"hiv_status": {
			"POS": "Positive",
			"NEG": "Negative",
			"UNK": "Unknown",
		},
		"rash_severity": {
			"MILD": "Mild",
			"MOD":  "Moderate",
			"SEV":  "Severe",
		},
	}

	data := fiber.Map{
		"ClientID":      clientID,
		"EncounterDate": dateStr,
		"Client":        client,
		"Optionz":       options,
		"HasAdmission":  hasAdmission,
		"AdmissionID":   admissionID,
	}

	// Use GenerateHTML helper to render the template
	return GenerateHTML(c, db, data, "mpox_admission")
}

// HandlerMpoxAdmissionSubmit handles the submission of the mpox admission form
func HandlerMpoxAdmissionSubmit(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	// Start a transaction
	tx, err := db.Begin()
	if err != nil {
		sl.Error("Error starting transaction", "error", err)
		return c.Status(http.StatusInternalServerError).SendString("Error saving admission data")
	}
	defer tx.Rollback()

	// Parse form data
	clientID, _ := strconv.Atoi(c.FormValue("client_id"))

	// Create and insert demographics
	demographics := models.MpoxDemographics{
		ClientID:             sql.NullInt64{Int64: int64(clientID), Valid: true},
		Sex:                  sql.NullString{String: c.FormValue("sex"), Valid: true},
		DateOfBirth:          sql.NullTime{Time: utils.ParseDate(c.FormValue("date_of_birth")), Valid: true},
		AgeYears:             sql.NullInt64{Int64: utils.ParseInt64(c.FormValue("age_years")), Valid: true},
		AgeMonths:            sql.NullInt64{Int64: utils.ParseInt64(c.FormValue("age_months")), Valid: true},
		AgeDays:              sql.NullInt64{Int64: utils.ParseInt64(c.FormValue("age_days")), Valid: true},
		HealthCareWorker:     sql.NullString{String: c.FormValue("health_care_worker"), Valid: true},
		LaboratoryWorker:     sql.NullString{String: c.FormValue("laboratory_worker"), Valid: true},
		PPEStatus:            sql.NullString{String: c.FormValue("ppe_status"), Valid: true},
		Tribe:                sql.NullString{String: c.FormValue("tribe"), Valid: true},
		Pregnant:             sql.NullBool{Bool: c.FormValue("pregnant") == "Y", Valid: true},
		GestationalWeeks:     sql.NullInt64{Int64: utils.ParseInt64(c.FormValue("gestational_weeks")), Valid: true},
		LMNP:                 sql.NullTime{Time: utils.ParseDate(c.FormValue("lmnp")), Valid: true},
		RecentlyPregnant:     sql.NullBool{Bool: c.FormValue("recently_pregnant") == "Y", Valid: true},
		Pregnant22_42:        sql.NullBool{Bool: c.FormValue("pregnant_22_42") == "Y", Valid: true},
		TetanusVaccination:   sql.NullBool{Bool: c.FormValue("tetanus_vaccination") == "Y", Valid: true},
		Occupation:           sql.NullString{String: c.FormValue("occupation"), Valid: true},
		SiteOfFirstEncounter: sql.NullString{String: c.FormValue("site_of_first_encounter"), Valid: true},
	}

	err = demographics.Insert(db)
	if err != nil {
		sl.Error("Error inserting demographics", "error", err)
		return c.Status(http.StatusInternalServerError).SendString("Error saving demographics")
	}

	// Create and insert exposure history
	exposure := models.MpoxExposureHistory{
		DemographicsID: demographics.ID,
		KnownLink:      sql.NullBool{Bool: c.FormValue("known_link") == "Y", Valid: true},
		SexuallyActive: sql.NullBool{Bool: c.FormValue("sexually_active") == "Y", Valid: true},
		SexOfPartners:  sql.NullString{String: c.FormValue("sex_of_partners"), Valid: true},
		RecentTravel:   sql.NullBool{Bool: c.FormValue("recent_travel") == "Y", Valid: true},
		TravelHighRisk: sql.NullBool{Bool: c.FormValue("travel_high_risk") == "Y", Valid: true},
		TravelDetails:  sql.NullString{String: c.FormValue("travel_details"), Valid: true},
	}

	err = exposure.Insert(db)
	if err != nil {
		sl.Error("Error inserting exposure history", "error", err)
		return c.Status(http.StatusInternalServerError).SendString("Error saving exposure history")
	}

	// Create and insert onset vitals
	vitals := models.MpoxOnsetVitals{
		DemographicsID:  demographics.ID,
		SymptomOnset:    sql.NullTime{Time: utils.ParseDate(c.FormValue("symptom_onset")), Valid: true},
		Fever:           sql.NullBool{Bool: c.FormValue("fever") == "Y", Valid: true},
		SoreThroat:      sql.NullBool{Bool: c.FormValue("sore_throat") == "Y", Valid: true},
		Headache:        sql.NullBool{Bool: c.FormValue("headache") == "Y", Valid: true},
		MuscleAches:     sql.NullBool{Bool: c.FormValue("muscle_aches") == "Y", Valid: true},
		Cough:           sql.NullBool{Bool: c.FormValue("cough") == "Y", Valid: true},
		Fatigue:         sql.NullBool{Bool: c.FormValue("fatigue") == "Y", Valid: true},
		OralPain:        sql.NullBool{Bool: c.FormValue("oral_pain") == "Y", Valid: true},
		Nausea:          sql.NullBool{Bool: c.FormValue("nausea") == "Y", Valid: true},
		Vomiting:        sql.NullBool{Bool: c.FormValue("vomiting") == "Y", Valid: true},
		Diarrhea:        sql.NullBool{Bool: c.FormValue("diarrhea") == "Y", Valid: true},
		RectalPain:      sql.NullBool{Bool: c.FormValue("rectal_pain") == "Y", Valid: true},
		Lesions:         sql.NullBool{Bool: c.FormValue("lesions") == "Y", Valid: true},
		Lymphadenopathy: sql.NullBool{Bool: c.FormValue("lymphadenopathy") == "Y", Valid: true},
		Temperature:     sql.NullFloat64{Float64: utils.ParseFloat64(c.FormValue("temperature")), Valid: true},
		HeartRate:       sql.NullInt64{Int64: utils.ParseInt64(c.FormValue("heart_rate")), Valid: true},
		RespiratoryRate: sql.NullInt64{Int64: utils.ParseInt64(c.FormValue("respiratory_rate")), Valid: true},
		BpSystolic:      sql.NullInt64{Int64: utils.ParseInt64(c.FormValue("bp_systolic")), Valid: true},
		BpDiastolic:     sql.NullInt64{Int64: utils.ParseInt64(c.FormValue("bp_diastolic")), Valid: true},
		Dehydration:     sql.NullBool{Bool: c.FormValue("dehydration") == "Y", Valid: true},
		AVPU:            sql.NullString{String: c.FormValue("avpu"), Valid: true},
		HeightCm:        sql.NullFloat64{Float64: utils.ParseFloat64(c.FormValue("height_cm")), Valid: true},
		WeightKg:        sql.NullFloat64{Float64: utils.ParseFloat64(c.FormValue("weight_kg")), Valid: true},
	}

	err = vitals.Insert(db)
	if err != nil {
		sl.Error("Error inserting onset vitals", "error", err)
		return c.Status(http.StatusInternalServerError).SendString("Error saving onset vitals")
	}

	// Create and insert comorbidities
	comorbidities := models.MpoxComorbidities{
		DemographicsID:       demographics.ID,
		CardiacDisease:       sql.NullBool{Bool: c.FormValue("cardiac_disease") == "Y", Valid: true},
		Hypertension:         sql.NullBool{Bool: c.FormValue("hypertension") == "Y", Valid: true},
		PulmonaryDisease:     sql.NullBool{Bool: c.FormValue("pulmonary_disease") == "Y", Valid: true},
		Asthma:               sql.NullBool{Bool: c.FormValue("asthma") == "Y", Valid: true},
		KidneyDisease:        sql.NullBool{Bool: c.FormValue("kidney_disease") == "Y", Valid: true},
		LiverDisease:         sql.NullBool{Bool: c.FormValue("liver_disease") == "Y", Valid: true},
		NeurologicalDisorder: sql.NullBool{Bool: c.FormValue("neurological_disorder") == "Y", Valid: true},
		Diabetes:             sql.NullBool{Bool: c.FormValue("diabetes") == "Y", Valid: true},
		TuberculosisActive:   sql.NullBool{Bool: c.FormValue("tuberculosis_active") == "Y", Valid: true},
		TuberculosisPrevious: sql.NullBool{Bool: c.FormValue("tuberculosis_previous") == "Y", Valid: true},
		Asplenia:             sql.NullBool{Bool: c.FormValue("asplenia") == "Y", Valid: true},
		Neoplasm:             sql.NullBool{Bool: c.FormValue("neoplasm") == "Y", Valid: true},
		AlcoholUseDisorder:   sql.NullBool{Bool: c.FormValue("alcohol_use_disorder") == "Y", Valid: true},
		Immunosuppressants:   sql.NullBool{Bool: c.FormValue("immunosuppressants") == "Y", Valid: true},
		STI:                  sql.NullBool{Bool: c.FormValue("sti") == "Y", Valid: true},
		HIVStatus:            sql.NullString{String: c.FormValue("hiv_status"), Valid: true},
		ARTRegimen:           sql.NullString{String: c.FormValue("art_regimen"), Valid: true},
		CD4:                  sql.NullInt64{Int64: utils.ParseInt64(c.FormValue("cd4")), Valid: true},
		ViralLoad:            sql.NullString{String: c.FormValue("viral_load"), Valid: true},
	}

	err = comorbidities.Insert(db)
	if err != nil {
		sl.Error("Error inserting comorbidities", "error", err)
		return c.Status(http.StatusInternalServerError).SendString("Error saving comorbidities")
	}

	// Create and insert rash evaluation
	rash := models.MpoxRashEvaluation{
		DemographicsID:       demographics.ID,
		Severity:             sql.NullString{String: c.FormValue("severity"), Valid: true},
		Face:                 sql.NullBool{Bool: c.FormValue("face") == "Y", Valid: true},
		Nares:                sql.NullBool{Bool: c.FormValue("nares") == "Y", Valid: true},
		Mouth:                sql.NullBool{Bool: c.FormValue("mouth") == "Y", Valid: true},
		Chest:                sql.NullBool{Bool: c.FormValue("chest") == "Y", Valid: true},
		Abdomen:              sql.NullBool{Bool: c.FormValue("abdomen") == "Y", Valid: true},
		Back:                 sql.NullBool{Bool: c.FormValue("back") == "Y", Valid: true},
		Perianal:             sql.NullBool{Bool: c.FormValue("perianal") == "Y", Valid: true},
		Genitals:             sql.NullBool{Bool: c.FormValue("genitals") == "Y", Valid: true},
		Palms:                sql.NullBool{Bool: c.FormValue("palms") == "Y", Valid: true},
		Arms:                 sql.NullBool{Bool: c.FormValue("arms") == "Y", Valid: true},
		Forearms:             sql.NullBool{Bool: c.FormValue("forearms") == "Y", Valid: true},
		Thighs:               sql.NullBool{Bool: c.FormValue("thighs") == "Y", Valid: true},
		Legs:                 sql.NullBool{Bool: c.FormValue("legs") == "Y", Valid: true},
		Soles:                sql.NullBool{Bool: c.FormValue("soles") == "Y", Valid: true},
		Macule:               sql.NullBool{Bool: c.FormValue("macule") == "Y", Valid: true},
		Papule:               sql.NullBool{Bool: c.FormValue("papule") == "Y", Valid: true},
		EarlyVesicle:         sql.NullBool{Bool: c.FormValue("early_vesicle") == "Y", Valid: true},
		SmallPustule:         sql.NullBool{Bool: c.FormValue("small_pustule") == "Y", Valid: true},
		UmbilicatedPustule:   sql.NullBool{Bool: c.FormValue("umbilicated_pustule") == "Y", Valid: true},
		UlceratedLesion:      sql.NullBool{Bool: c.FormValue("ulcerated_lesion") == "Y", Valid: true},
		Crusting:             sql.NullBool{Bool: c.FormValue("crusting") == "Y", Valid: true},
		PartiallyRemovedScab: sql.NullBool{Bool: c.FormValue("partially_removed_scab") == "Y", Valid: true},
		PainAtLesion:         sql.NullBool{Bool: c.FormValue("pain_at_lesion") == "Y", Valid: true},
		PainScore:            sql.NullInt64{Int64: utils.ParseInt64(c.FormValue("pain_score")), Valid: true},
	}

	err = rash.Insert(db)
	if err != nil {
		sl.Error("Error inserting rash evaluation", "error", err)
		return c.Status(http.StatusInternalServerError).SendString("Error saving rash evaluation")
	}

	// Create and insert laboratory investigations
	labs := models.MpoxLaboratoryInvestigations{
		DemographicsID:  demographics.ID,
		ALT:             sql.NullFloat64{Float64: utils.ParseFloat64(c.FormValue("alt")), Valid: true},
		AST:             sql.NullFloat64{Float64: utils.ParseFloat64(c.FormValue("ast")), Valid: true},
		Creatinine:      sql.NullFloat64{Float64: utils.ParseFloat64(c.FormValue("creatinine")), Valid: true},
		Potassium:       sql.NullFloat64{Float64: utils.ParseFloat64(c.FormValue("potassium")), Valid: true},
		Urea:            sql.NullFloat64{Float64: utils.ParseFloat64(c.FormValue("urea")), Valid: true},
		CreatineKinase:  sql.NullFloat64{Float64: utils.ParseFloat64(c.FormValue("creatine_kinase")), Valid: true},
		Calcium:         sql.NullFloat64{Float64: utils.ParseFloat64(c.FormValue("calcium")), Valid: true},
		Sodium:          sql.NullFloat64{Float64: utils.ParseFloat64(c.FormValue("sodium")), Valid: true},
		CRP:             sql.NullFloat64{Float64: utils.ParseFloat64(c.FormValue("crp")), Valid: true},
		Glucose:         sql.NullFloat64{Float64: utils.ParseFloat64(c.FormValue("glucose")), Valid: true},
		Lactate:         sql.NullFloat64{Float64: utils.ParseFloat64(c.FormValue("lactate")), Valid: true},
		Haemoglobin:     sql.NullFloat64{Float64: utils.ParseFloat64(c.FormValue("haemoglobin")), Valid: true},
		TotalBilirubin:  sql.NullFloat64{Float64: utils.ParseFloat64(c.FormValue("total_bilirubin")), Valid: true},
		WBCCount:        sql.NullFloat64{Float64: utils.ParseFloat64(c.FormValue("wbc_count")), Valid: true},
		Platelets:       sql.NullFloat64{Float64: utils.ParseFloat64(c.FormValue("platelets")), Valid: true},
		ProthrombinTime: sql.NullFloat64{Float64: utils.ParseFloat64(c.FormValue("prothrombin_time")), Valid: true},
		APTT:            sql.NullFloat64{Float64: utils.ParseFloat64(c.FormValue("aptt")), Valid: true},
		MalariaResult:   sql.NullString{String: c.FormValue("malaria_result"), Valid: true},
		SyphilisResult:  sql.NullString{String: c.FormValue("syphilis_result"), Valid: true},
		MpoxResult:      sql.NullString{String: c.FormValue("mpox_result"), Valid: true},
	}

	err = labs.Insert(db)
	if err != nil {
		sl.Error("Error inserting laboratory investigations", "error", err)
		return c.Status(http.StatusInternalServerError).SendString("Error saving laboratory investigations")
	}

	// Create and insert data entrant
	dataEntrant := models.MpoxDataEntrant{
		DemographicsID: demographics.ID,
		Name:           sql.NullString{String: c.FormValue("data_entrant_name"), Valid: true},
	}

	err = dataEntrant.Insert(db)
	if err != nil {
		sl.Error("Error inserting data entrant", "error", err)
		return c.Status(http.StatusInternalServerError).SendString("Error saving data entrant")
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		sl.Error("Error committing transaction", "error", err)
		return c.Status(http.StatusInternalServerError).SendString("Error saving admission data")
	}

	// Redirect to client page
	return c.Redirect("/cases/new/" + strconv.Itoa(clientID))
}

//Get all admission data by
