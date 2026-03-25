package handlers

import (
	"case/internal/config"
	"case/internal/models"
	"case/internal/utils"
	"database/sql"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
)

// Helper function to get multiple checkbox values
func getCheckboxValues(c *fiber.Ctx, fieldName string) []string {
	var values []string
	form, err := c.MultipartForm()
	if err == nil && form != nil {
		if formValues, exists := form.Value[fieldName]; exists {
			values = formValues
		}
	}
	// Fallback to single value if multipart form fails
	if len(values) == 0 {
		if singleValue := c.FormValue(fieldName); singleValue != "" {
			values = []string{singleValue}
		}
	}
	return values
}

// Helper function to join checkbox values into a comma-separated string
func joinCheckboxValues(values []string) string {
	if len(values) == 0 {
		return ""
	}
	return strings.Join(values, ", ")
}

// HandlerMpoxAdmissionForm renders the mpox admission form
func HandlerMpoxAdmissionForm(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config config.Config) error {
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
	err = db.QueryRow("SELECT id, firstname, lastname, site FROM clients WHERE id = $1", clientID).Scan(&client.ID, &client.Firstname, &client.Lastname, &client.Site)
	if err != nil {
		sl.Error("Error fetching client", "error", err)
		return c.Status(http.StatusInternalServerError).SendString("Error fetching client details")
	}

	// Check facility-based access control
	userID := GetCurrentUser(c, store)
	userFacility := GetCurrentFacility(c, db, sl, store)
	if userFacility > 0 {
		// User has a facility assigned, check if they can access this case
		if client.Site.Int64 != int64(userFacility) {
			sl.Error("User attempted to access mpox admission for case from different facility",
				"user_id", userID, "user_facility", userFacility, "case_site", client.Site.Int64, "case_id", clientID)
			return c.Status(403).SendString("Access denied: You can only access cases from your assigned facility.")
		}
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

	data := NewTemplateData(c, store)
	data.Form = fiber.Map{
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
func HandlerMpoxAdmissionSubmit(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config config.Config) error {
	// Start a transaction
	tx, err := db.Begin()
	if err != nil {
		sl.Error("Error starting transaction", "error", err)
		return c.Status(http.StatusInternalServerError).SendString("Error saving admission data")
	}
	defer tx.Rollback()

	// Parse form data
	clientID, _ := strconv.Atoi(c.FormValue("client_id"))

	// Check facility-based access control
	userID := GetCurrentUser(c, store)
	userFacility := GetCurrentFacility(c, db, sl, store)
	if userFacility > 0 {
		// Get client details to check facility
		var clientSite sql.NullInt64
		err := db.QueryRow("SELECT site FROM clients WHERE id = $1", clientID).Scan(&clientSite)
		if err != nil {
			sl.Error("Failed to get client for facility check", "error", err, "clientID", clientID)
			return c.Status(http.StatusInternalServerError).SendString("Failed to get client details")
		}

		// User has a facility assigned, check if they can access this case
		if clientSite.Int64 != int64(userFacility) {
			sl.Error("User attempted to submit mpox admission for case from different facility",
				"user_id", userID, "user_facility", userFacility, "case_site", clientSite.Int64, "case_id", clientID)
			return c.Status(403).SendString("Access denied: You can only access cases from your assigned facility.")
		}
	}

	// Create and insert demographics
	demographics := models.MpoxDemographics{
		ClientID:                  sql.NullInt64{Int64: int64(clientID), Valid: true},
		Sex:                       sql.NullString{String: c.FormValue("sex"), Valid: true},
		DateOfBirth:               ParseNullTime(c.FormValue("date_of_birth")),
		AgeYears:                  sql.NullInt64{Int64: utils.ParseInt64(c.FormValue("age_years")), Valid: true},
		AgeMonths:                 sql.NullInt64{Int64: utils.ParseInt64(c.FormValue("age_months")), Valid: true},
		AgeDays:                   sql.NullInt64{Int64: utils.ParseInt64(c.FormValue("age_days")), Valid: true},
		HealthCareWorker:          sql.NullString{String: c.FormValue("health_care_worker"), Valid: true},
		LaboratoryWorker:          sql.NullString{String: c.FormValue("laboratory_worker"), Valid: true},
		PPEStatus:                 sql.NullString{String: c.FormValue("ppe_status"), Valid: true},
		Tribe:                     sql.NullString{String: c.FormValue("tribe"), Valid: true},
		Pregnant:                  sql.NullBool{Bool: c.FormValue("pregnant") == "Y", Valid: true},
		GestationalWeeks:          sql.NullInt64{Int64: utils.ParseInt64(c.FormValue("gestational_weeks")), Valid: true},
		LMNP:                      ParseNullTime(c.FormValue("lmnp")),
		RecentlyPregnant:          sql.NullBool{Bool: c.FormValue("recently_pregnant") == "Y", Valid: true},
		Pregnant22_42:             sql.NullBool{Bool: c.FormValue("pregnant_22_42") == "Y", Valid: true},
		TetanusVaccination:        sql.NullBool{Bool: c.FormValue("tetanus_vaccination") == "Y", Valid: true},
		Occupation:                sql.NullString{String: c.FormValue("occupation"), Valid: true},
		SiteOfFirstEncounter:      sql.NullString{String: joinCheckboxValues(getCheckboxValues(c, "site_of_first_encounter")), Valid: true},
		SiteOfFirstEncounterOther: sql.NullString{String: c.FormValue("site_of_first_encounter_other"), Valid: true},
		SuspectConfirmedCase:      sql.NullString{String: c.FormValue("suspect_confirmed_case"), Valid: true},
		LymphPainful:              sql.NullString{String: c.FormValue("lymph_painful"), Valid: true},
		LymphLocation:             sql.NullString{String: joinCheckboxValues(getCheckboxValues(c, "lymph_location")), Valid: true},
		LymphOtherDetail:          sql.NullString{String: c.FormValue("lymph_other_detail"), Valid: true},
		LymphPainLocation:         sql.NullString{String: joinCheckboxValues(getCheckboxValues(c, "lymph_pain_location")), Valid: true},
		LymphPainOtherDetail:      sql.NullString{String: c.FormValue("lymph_pain_other_detail"), Valid: true},
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
		SexOfPartners:  sql.NullString{String: joinCheckboxValues(getCheckboxValues(c, "sex_of_partners")), Valid: true},
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
		DemographicsID:                    demographics.ID,
		SymptomOnset:                      ParseNullTime(c.FormValue("symptom_onset")),
		Fever:                             sql.NullBool{Bool: c.FormValue("fever") == "Y", Valid: true},
		FeverOnsetDate:                    ParseNullTime(c.FormValue("fever_onset_date")),
		SoreThroat:                        sql.NullBool{Bool: c.FormValue("sore_throat") == "Y", Valid: true},
		SoreThroatOnsetDate:               ParseNullTime(c.FormValue("sore_throat_onset_date")),
		Headache:                          sql.NullBool{Bool: c.FormValue("headache") == "Y", Valid: true},
		HeadacheOnsetDate:                 ParseNullTime(c.FormValue("headache_onset_date")),
		MuscleAches:                       sql.NullBool{Bool: c.FormValue("muscle_aches") == "Y", Valid: true},
		MuscleAchesOnsetDate:              ParseNullTime(c.FormValue("muscle_aches_onset_date")),
		Cough:                             sql.NullBool{Bool: c.FormValue("cough") == "Y", Valid: true},
		CoughOnsetDate:                    ParseNullTime(c.FormValue("cough_onset_date")),
		Fatigue:                           sql.NullBool{Bool: c.FormValue("fatigue") == "Y", Valid: true},
		FatigueOnsetDate:                  ParseNullTime(c.FormValue("fatigue_onset_date")),
		OralPain:                          sql.NullBool{Bool: c.FormValue("oral_pain") == "Y", Valid: true},
		OralPainOnsetDate:                 ParseNullTime(c.FormValue("oral_pain_onset_date")),
		Nausea:                            sql.NullBool{Bool: c.FormValue("nausea") == "Y", Valid: true},
		NauseaOnsetDate:                   ParseNullTime(c.FormValue("nausea_onset_date")),
		Vomiting:                          sql.NullBool{Bool: c.FormValue("vomiting") == "Y", Valid: true},
		VomitingOnsetDate:                 ParseNullTime(c.FormValue("vomiting_onset_date")),
		Diarrhea:                          sql.NullBool{Bool: c.FormValue("diarrhea") == "Y", Valid: true},
		DiarrheaOnsetDate:                 ParseNullTime(c.FormValue("diarrhea_onset_date")),
		RectalPain:                        sql.NullBool{Bool: c.FormValue("rectal_pain") == "Y", Valid: true},
		RectalPainOnsetDate:               ParseNullTime(c.FormValue("rectal_pain_onset_date")),
		Lesions:                           sql.NullBool{Bool: c.FormValue("lesions") == "Y", Valid: true},
		LesionsOnsetDate:                  ParseNullTime(c.FormValue("lesions_onset_date")),
		Lymphadenopathy:                   sql.NullBool{Bool: c.FormValue("lymphadenopathy") == "Y", Valid: true},
		LymphadenopathyOnsetDate:          ParseNullTime(c.FormValue("lymphadenopathy_onset_date")),
		Pruritis:                          sql.NullBool{Bool: c.FormValue("pruritis") == "Y", Valid: true},
		PruritisOnsetDate:                 ParseNullTime(c.FormValue("pruritis_onset_date")),
		PainSwallowing:                    sql.NullBool{Bool: c.FormValue("pain_swallowing") == "Y", Valid: true},
		PainSwallowingOnsetDate:           ParseNullTime(c.FormValue("pain_swallowing_onset_date")),
		DifficultySwallowing:              sql.NullBool{Bool: c.FormValue("difficulty_swallowing") == "Y", Valid: true},
		DifficultySwallowingOnsetDate:     ParseNullTime(c.FormValue("difficulty_swallowing_onset_date")),
		Urethritis:                        sql.NullBool{Bool: c.FormValue("urethritis") == "Y", Valid: true},
		UrethritisOnsetDate:               ParseNullTime(c.FormValue("urethritis_onset_date")),
		ChestPain:                         sql.NullBool{Bool: c.FormValue("chest_pain") == "Y", Valid: true},
		ChestPainOnsetDate:                ParseNullTime(c.FormValue("chest_pain_onset_date")),
		DecreasedUrine:                    sql.NullBool{Bool: c.FormValue("decreased_urine") == "Y", Valid: true},
		DecreasedUrineOnsetDate:           ParseNullTime(c.FormValue("decreased_urine_onset_date")),
		Dizziness:                         sql.NullBool{Bool: c.FormValue("dizziness") == "Y", Valid: true},
		DizzinessOnsetDate:                ParseNullTime(c.FormValue("dizziness_onset_date")),
		JointPain:                         sql.NullBool{Bool: c.FormValue("joint_pain") == "Y", Valid: true},
		JointPainOnsetDate:                ParseNullTime(c.FormValue("joint_pain_onset_date")),
		PsychologicalDisturbance:          sql.NullBool{Bool: c.FormValue("psychological_disturbance") == "Y", Valid: true},
		PsychologicalDisturbanceOnsetDate: ParseNullTime(c.FormValue("psychological_disturbance_onset_date")),
		Temperature:                       sql.NullFloat64{Float64: utils.ParseFloat64(c.FormValue("temperature")), Valid: true},
		HeartRate:                         sql.NullInt64{Int64: utils.ParseInt64(c.FormValue("heart_rate")), Valid: true},
		RespiratoryRate:                   sql.NullInt64{Int64: utils.ParseInt64(c.FormValue("respiratory_rate")), Valid: true},
		BpSystolic:                        sql.NullInt64{Int64: utils.ParseInt64(c.FormValue("bp_systolic")), Valid: true},
		BpDiastolic:                       sql.NullInt64{Int64: utils.ParseInt64(c.FormValue("bp_diastolic")), Valid: true},
		Dehydration:                       sql.NullBool{Bool: c.FormValue("dehydration") == "Y", Valid: true},
		AVPU:                              sql.NullString{String: c.FormValue("avpu"), Valid: true},
		HeightCm:                          sql.NullFloat64{Float64: utils.ParseFloat64(c.FormValue("height_cm")), Valid: true},
		WeightKg:                          sql.NullFloat64{Float64: utils.ParseFloat64(c.FormValue("weight_kg")), Valid: true},
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
		DemographicsID:        demographics.ID,
		ALT:                   sql.NullFloat64{Float64: utils.ParseFloat64(c.FormValue("lab_alt")), Valid: true},
		AST:                   sql.NullFloat64{Float64: utils.ParseFloat64(c.FormValue("lab_ast")), Valid: true},
		Creatinine:            sql.NullFloat64{Float64: utils.ParseFloat64(c.FormValue("lab_creatinine")), Valid: true},
		Potassium:             sql.NullFloat64{Float64: utils.ParseFloat64(c.FormValue("lab_potassium")), Valid: true},
		Urea:                  sql.NullFloat64{Float64: utils.ParseFloat64(c.FormValue("lab_urea")), Valid: true},
		CreatineKinase:        sql.NullFloat64{Float64: utils.ParseFloat64(c.FormValue("lab_ck")), Valid: true},
		Calcium:               sql.NullFloat64{Float64: utils.ParseFloat64(c.FormValue("lab_calcium")), Valid: true},
		Sodium:                sql.NullFloat64{Float64: utils.ParseFloat64(c.FormValue("lab_sodium")), Valid: true},
		CRP:                   sql.NullFloat64{Float64: utils.ParseFloat64(c.FormValue("lab_crp")), Valid: true},
		Glucose:               sql.NullFloat64{Float64: utils.ParseFloat64(c.FormValue("lab_glucose")), Valid: true},
		Lactate:               sql.NullFloat64{Float64: utils.ParseFloat64(c.FormValue("lab_lactate")), Valid: true},
		Haemoglobin:           sql.NullFloat64{Float64: utils.ParseFloat64(c.FormValue("lab_haemoglobin")), Valid: true},
		TotalBilirubin:        sql.NullFloat64{Float64: utils.ParseFloat64(c.FormValue("lab_bilirubin")), Valid: true},
		WBCCount:              sql.NullFloat64{Float64: utils.ParseFloat64(c.FormValue("lab_wbc")), Valid: true},
		Platelets:             sql.NullFloat64{Float64: utils.ParseFloat64(c.FormValue("lab_platelets")), Valid: true},
		ProthrombinTime:       sql.NullFloat64{Float64: utils.ParseFloat64(c.FormValue("lab_prothrombin")), Valid: true},
		APTT:                  sql.NullFloat64{Float64: utils.ParseFloat64(c.FormValue("lab_aptt")), Valid: true},
		TotalProtein:          sql.NullFloat64{Float64: utils.ParseFloat64(c.FormValue("total_protein")), Valid: true},
		Albumin:               sql.NullFloat64{Float64: utils.ParseFloat64(c.FormValue("albumin")), Valid: true},
		BilirubinD:            sql.NullFloat64{Float64: utils.ParseFloat64(c.FormValue("lab_bilirubin_d")), Valid: true},
		Lymphocytes:           sql.NullFloat64{Float64: utils.ParseFloat64(c.FormValue("lab_lymphocytes")), Valid: true},
		Monocytes:             sql.NullFloat64{Float64: utils.ParseFloat64(c.FormValue("lab_monocytes")), Valid: true},
		Eosinophils:           sql.NullFloat64{Float64: utils.ParseFloat64(c.FormValue("lab_eosinophils")), Valid: true},
		Basophils:             sql.NullFloat64{Float64: utils.ParseFloat64(c.FormValue("lab_basophils")), Valid: true},
		Neutrophils:           sql.NullFloat64{Float64: utils.ParseFloat64(c.FormValue("lab_neutrophils")), Valid: true},
		HGB:                   sql.NullFloat64{Float64: utils.ParseFloat64(c.FormValue("lab_hgb")), Valid: true},
		HCT:                   sql.NullFloat64{Float64: utils.ParseFloat64(c.FormValue("lab_hct")), Valid: true},
		MCV:                   sql.NullFloat64{Float64: utils.ParseFloat64(c.FormValue("lab_mcv")), Valid: true},
		MCH:                   sql.NullFloat64{Float64: utils.ParseFloat64(c.FormValue("lab_mch")), Valid: true},
		MCHC:                  sql.NullFloat64{Float64: utils.ParseFloat64(c.FormValue("lab_mchc")), Valid: true},
		RDW:                   sql.NullFloat64{Float64: utils.ParseFloat64(c.FormValue("lab_rdw")), Valid: true},
		RDWSD:                 sql.NullFloat64{Float64: utils.ParseFloat64(c.FormValue("lab_rdw_sd")), Valid: true},
		RDWCV:                 sql.NullFloat64{Float64: utils.ParseFloat64(c.FormValue("lab_rdw_cv")), Valid: true},
		MPV:                   sql.NullFloat64{Float64: utils.ParseFloat64(c.FormValue("lab_mpv")), Valid: true},
		PDW:                   sql.NullString{String: c.FormValue("pdw"), Valid: true},
		PCT:                   sql.NullFloat64{Float64: utils.ParseFloat64(c.FormValue("lab_pct")), Valid: true},
		LabOther:              sql.NullString{String: c.FormValue("lab_other"), Valid: true},
		LabALTNotDone:         sql.NullBool{Bool: c.FormValue("lab_alt_notdone") == "on", Valid: true},
		LabASTNotDone:         sql.NullBool{Bool: c.FormValue("lab_ast_notdone") == "on", Valid: true},
		LabCreatinineNotDone:  sql.NullBool{Bool: c.FormValue("lab_creatinine_notdone") == "on", Valid: true},
		LabPotassiumNotDone:   sql.NullBool{Bool: c.FormValue("lab_potassium_notdone") == "on", Valid: true},
		TotalProteinNotDone:   sql.NullBool{Bool: c.FormValue("total_protein_notdone") == "on", Valid: true},
		AlbuminNotDone:        sql.NullBool{Bool: c.FormValue("albumin_notdone") == "on", Valid: true},
		LabUreaNotDone:        sql.NullBool{Bool: c.FormValue("lab_urea_notdone") == "on", Valid: true},
		LabCKNotDone:          sql.NullBool{Bool: c.FormValue("lab_ck_notdone") == "on", Valid: true},
		LabCalciumNotDone:     sql.NullBool{Bool: c.FormValue("lab_calcium_notdone") == "on", Valid: true},
		LabSodiumNotDone:      sql.NullBool{Bool: c.FormValue("lab_sodium_notdone") == "on", Valid: true},
		LabLymphocytesNotDone: sql.NullBool{Bool: c.FormValue("lab_lymphocytes_notdone") == "on", Valid: true},
		LabMonocytesNotDone:   sql.NullBool{Bool: c.FormValue("lab_monocytes_notdone") == "on", Valid: true},
		LabEosinophilsNotDone: sql.NullBool{Bool: c.FormValue("lab_eosinophils_notdone") == "on", Valid: true},
		LabBasophilsNotDone:   sql.NullBool{Bool: c.FormValue("lab_basophils_notdone") == "on", Valid: true},
		LabCRPNotDone:         sql.NullBool{Bool: c.FormValue("lab_crp_notdone") == "on", Valid: true},
		LabNeutrophilsNotDone: sql.NullBool{Bool: c.FormValue("lab_neutrophils_notdone") == "on", Valid: true},
		LabHGBNotDone:         sql.NullBool{Bool: c.FormValue("lab_hgb_notdone") == "on", Valid: true},
		LabHCTNotDone:         sql.NullBool{Bool: c.FormValue("lab_hct_notdone") == "on", Valid: true},
		LabMCVNotDone:         sql.NullBool{Bool: c.FormValue("lab_mcv_notdone") == "on", Valid: true},
		LabMCHNotDone:         sql.NullBool{Bool: c.FormValue("lab_mch_notdone") == "on", Valid: true},
		LabMCHCNotDone:        sql.NullBool{Bool: c.FormValue("lab_mchc_notdone") == "on", Valid: true},
		LabRDWNotDone:         sql.NullBool{Bool: c.FormValue("lab_rdw_notdone") == "on", Valid: true},
		LabRDWSDNotDone:       sql.NullBool{Bool: c.FormValue("lab_rdw_sd_notdone") == "on", Valid: true},
		LabRDWCVNotDone:       sql.NullBool{Bool: c.FormValue("lab_rdw_cv_notdone") == "on", Valid: true},
		LabMPVNotDone:         sql.NullBool{Bool: c.FormValue("lab_mpv_notdone") == "on", Valid: true},
		LabPDWNotDone:         sql.NullBool{Bool: c.FormValue("lab_pdw_notdone") == "on", Valid: true},
		LabPCTNotDone:         sql.NullBool{Bool: c.FormValue("lab_pct_notdone") == "on", Valid: true},
		LabOtherNotDone:       sql.NullBool{Bool: c.FormValue("lab_other_notdone") == "on", Valid: true},
		LabGlucoseNotDone:     sql.NullBool{Bool: c.FormValue("lab_glucose_notdone") == "on", Valid: true},
		LabLactateNotDone:     sql.NullBool{Bool: c.FormValue("lab_lactate_notdone") == "on", Valid: true},
		LabHaemoglobinNotDone: sql.NullBool{Bool: c.FormValue("lab_haemoglobin_notdone") == "on", Valid: true},
		LabBilirubinNotDone:   sql.NullBool{Bool: c.FormValue("lab_bilirubin_notdone") == "on", Valid: true},
		LabBilirubinDNotDone:  sql.NullBool{Bool: c.FormValue("lab_bilirubin_d_notdone") == "on", Valid: true},
		LabWBCNotDone:         sql.NullBool{Bool: c.FormValue("lab_wbc_notdone") == "on", Valid: true},
		LabPlateletsNotDone:   sql.NullBool{Bool: c.FormValue("lab_platelets_notdone") == "on", Valid: true},
		LabProthrombinNotDone: sql.NullBool{Bool: c.FormValue("lab_prothrombin_notdone") == "on", Valid: true},
		LabAPTTNotDone:        sql.NullBool{Bool: c.FormValue("lab_aptt_notdone") == "on", Valid: true},
		OtherMalaria:          sql.NullString{String: c.FormValue("other_malaria"), Valid: true},
		OtherHIV:              sql.NullString{String: c.FormValue("other_hiv"), Valid: true},
		OtherSyphilis:         sql.NullString{String: c.FormValue("other_syphilis"), Valid: true},
		OtherMpox:             sql.NullString{String: c.FormValue("other_mpox"), Valid: true},
		HepatitisB:            sql.NullString{String: c.FormValue("hepatitis_b"), Valid: true},
		HepatitisC:            sql.NullString{String: c.FormValue("hepatitis_c"), Valid: true},
		DataEntrantName:       sql.NullString{String: c.FormValue("data_entrant_name"), Valid: true},
	}

	// Robust two-step save to avoid INSERT column/value mismatches
	var mpoxLabID int
	if err := db.QueryRowContext(c.Context(), "INSERT INTO mpox_laboratory_investigations (demographics_id) VALUES ($1) RETURNING id", demographics.ID).Scan(&mpoxLabID); err != nil {
		sl.Error("Error creating mpox lab row", "error", err)
		return c.Status(http.StatusInternalServerError).SendString("Error saving laboratory investigations")
	}

	updateSQL := `UPDATE mpox_laboratory_investigations SET
		alt=$1, ast=$2, creatinine=$3, potassium=$4, urea=$5,
		creatine_kinase=$6, calcium=$7, sodium=$8, crp=$9, glucose=$10, lactate=$11,
		haemoglobin=$12, total_bilirubin=$13, wbc_count=$14, platelets=$15,
		prothrombin_time=$16, aptt=$17, total_protein=$18, albumin=$19, bilirubin_d=$20,
		lymphocytes=$21, monocytes=$22, eosinophils=$23, basophils=$24, neutrophils=$25,
		hgb=$26, hct=$27, mcv=$28, mch=$29, mchc=$30, rdw=$31, rdw_sd=$32, rdw_cv=$33,
		mpv=$34, pdw=$35, pct=$36, lab_other=$37,
		lab_alt_notdone=$38, lab_ast_notdone=$39, lab_creatinine_notdone=$40,
		lab_potassium_notdone=$41, total_protein_notdone=$42, albumin_notdone=$43,
		lab_urea_notdone=$44, lab_ck_notdone=$45, lab_calcium_notdone=$46, lab_sodium_notdone=$47,
		lab_lymphocytes_notdone=$48, lab_monocytes_notdone=$49, lab_eosinophils_notdone=$50,
		lab_basophils_notdone=$51, lab_crp_notdone=$52, lab_neutrophils_notdone=$53,
		lab_hgb_notdone=$54, lab_hct_notdone=$55, lab_mcv_notdone=$56, lab_mch_notdone=$57,
		lab_mchc_notdone=$58, lab_rdw_notdone=$59, lab_rdw_sd_notdone=$60, lab_rdw_cv_notdone=$61,
		lab_mpv_notdone=$62, lab_pdw_notdone=$63, lab_pct_notdone=$64, lab_other_notdone=$65,
		lab_glucose_notdone=$66, lab_lactate_notdone=$67, lab_haemoglobin_notdone=$68,
		lab_bilirubin_notdone=$69, lab_bilirubin_d_notdone=$70, lab_wbc_notdone=$71,
		lab_platelets_notdone=$72, lab_prothrombin_notdone=$73, lab_aptt_notdone=$74,
		other_malaria=$75, other_hiv=$76, other_syphilis=$77, other_mpox=$78, hepatitis_b=$79, hepatitis_c=$80,
		data_entrant_name=$81, updated_at=CURRENT_TIMESTAMP
		WHERE id=$82`

	if _, err := db.ExecContext(c.Context(), updateSQL,
		labs.ALT, labs.AST, labs.Creatinine, labs.Potassium, labs.Urea,
		labs.CreatineKinase, labs.Calcium, labs.Sodium, labs.CRP, labs.Glucose, labs.Lactate,
		labs.Haemoglobin, labs.TotalBilirubin, labs.WBCCount, labs.Platelets,
		labs.ProthrombinTime, labs.APTT, labs.TotalProtein, labs.Albumin, labs.BilirubinD,
		labs.Lymphocytes, labs.Monocytes, labs.Eosinophils, labs.Basophils, labs.Neutrophils,
		labs.HGB, labs.HCT, labs.MCV, labs.MCH, labs.MCHC, labs.RDW, labs.RDWSD, labs.RDWCV,
		labs.MPV, labs.PDW, labs.PCT, labs.LabOther,
		labs.LabALTNotDone, labs.LabASTNotDone, labs.LabCreatinineNotDone,
		labs.LabPotassiumNotDone, labs.TotalProteinNotDone, labs.AlbuminNotDone,
		labs.LabUreaNotDone, labs.LabCKNotDone, labs.LabCalciumNotDone, labs.LabSodiumNotDone,
		labs.LabLymphocytesNotDone, labs.LabMonocytesNotDone, labs.LabEosinophilsNotDone,
		labs.LabBasophilsNotDone, labs.LabCRPNotDone, labs.LabNeutrophilsNotDone,
		labs.LabHGBNotDone, labs.LabHCTNotDone, labs.LabMCVNotDone, labs.LabMCHNotDone,
		labs.LabMCHCNotDone, labs.LabRDWNotDone, labs.LabRDWSDNotDone, labs.LabRDWCVNotDone,
		labs.LabMPVNotDone, labs.LabPDWNotDone, labs.LabPCTNotDone, labs.LabOtherNotDone,
		labs.LabGlucoseNotDone, labs.LabLactateNotDone, labs.LabHaemoglobinNotDone,
		labs.LabBilirubinNotDone, labs.LabBilirubinDNotDone, labs.LabWBCNotDone,
		labs.LabPlateletsNotDone, labs.LabProthrombinNotDone, labs.LabAPTTNotDone,
		labs.OtherMalaria, labs.OtherHIV, labs.OtherSyphilis, labs.OtherMpox, labs.HepatitisB, labs.HepatitisC,
		labs.DataEntrantName, mpoxLabID); err != nil {
		sl.Error("Error updating mpox lab row", "error", err)
		return c.Status(http.StatusInternalServerError).SendString("Error saving laboratory investigations")
	}

	// Success for lab section

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
