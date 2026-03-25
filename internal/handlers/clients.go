package handlers

import (
	"case/internal/config"
	"case/internal/models"
	"case/internal/security"
	"case/internal/services"
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	// "io"
	// "encoding/json"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
)

// Define a struct for the encounter form page data
type EncounterPageData struct {
	FormRef           models.Client
	Form              []models.ClientEncounter
	Date              string
	FormChild1        []models.Clinical
	FormChild2        []models.Vital
	FormChild3        []models.Lab
	FormChild4        []FullTreatmentData
	AllEncounters     []models.ClientEncounter
	Optionz           map[string]map[string]string
	OutbreakID        int
	HasMpoxAdmission  bool
	MpoxAdmissionID   int
	FilteredEmployees []struct {
		ID   int64  `json:"id"`
		Name string `json:"name"`
	}
}

// AdmissionData holds all admission-related information
type AdmissionData struct {
	Demographics  *models.MpoxDemographics
	Exposure      *models.MpoxExposureHistory
	Vitals        *models.MpoxOnsetVitals
	Comorbidities *models.MpoxComorbidities
	Rash          *models.MpoxRashEvaluation
	Labs          *models.MpoxLaboratoryInvestigations
}

// FullTreatmentData represents complete treatment data for templates
type FullTreatmentData struct {
	TreatmentID                 int             `json:"treatment_id"`
	EncounterID                 sql.NullInt64   `json:"encounter_id"`
	OutbreakID                  sql.NullInt64   `json:"outbreak_id"`
	Antibacterial               sql.NullInt64   `json:"antibacterial"`
	Amoxicillin                 sql.NullInt64   `json:"amoxicillin"`
	Ceftriaxone                 sql.NullInt64   `json:"ceftriaxone"`
	Cefixime                    sql.NullInt64   `json:"cefixime"`
	Ampicillin                  sql.NullInt64   `json:"ampicillin"`
	Chloramphenicol             sql.NullInt64   `json:"chloramphenicol"`
	Amoxiclav                   sql.NullInt64   `json:"amoxiclav"`
	Azithromycin                sql.NullInt64   `json:"azithromycin"`
	Cefotaxime                  sql.NullInt64   `json:"cefotaxime"`
	Ceftazidime                 sql.NullInt64   `json:"ceftazidime"`
	Ciprofloxacin               sql.NullInt64   `json:"ciprofloxacin"`
	Tetracycline                sql.NullInt64   `json:"tetracycline"`
	Imipenem                    sql.NullInt64   `json:"imipenem"`
	Cotrimoxazole               sql.NullInt64   `json:"cotrimoxazole"`
	Gentamicin                  sql.NullInt64   `json:"gentamicin"`
	Metronidazole               sql.NullInt64   `json:"metronidazole"`
	AntibacterialOther          sql.NullString  `json:"antibacterial_other"`
	AntibacterialDose           sql.NullString  `json:"antibacterial_dose"`
	AntibacterialRoute          sql.NullString  `json:"antibacterial_route"`
	AntibacterialFreq           sql.NullString  `json:"antibacterial_freq"`
	Antimalarial                sql.NullInt64   `json:"antimalarial"`
	AntimalarialArtesunate      sql.NullInt64   `json:"antimalarial_artesunate"`
	AntimalarialArthemeter      sql.NullInt64   `json:"antimalarial_arthemeter"`
	AntimalarialAl              sql.NullInt64   `json:"antimalarial_al"`
	AntimalarialAa              sql.NullInt64   `json:"antimalarial_aa"`
	AntimalarialDose            sql.NullString  `json:"antimalarial_dose"`
	AntimalarialRoute           sql.NullString  `json:"antimalarial_route"`
	AntimalarialFreq            sql.NullString  `json:"antimalarial_freq"`
	OtherMedsSpecify            sql.NullString  `json:"other_meds_specify"`
	OtherMedsDose               sql.NullString  `json:"other_meds_dose"`
	OtherMedsRoute              sql.NullString  `json:"other_meds_route"`
	OtherMedsFreq               sql.NullString  `json:"other_meds_freq"`
	EbolaExperimental           sql.NullInt64   `json:"ebola_experimental"`
	EbolaExperimentalIf         sql.NullString  `json:"ebola_experimental_if"`
	Oral                        sql.NullInt64   `json:"oral"`
	OralOrs                     sql.NullInt64   `json:"oral_ors"`
	OralOrsQty                  sql.NullFloat64 `json:"oral_ors_qty"`
	OralWater                   sql.NullInt64   `json:"oral_water"`
	OralWaterQty                sql.NullFloat64 `json:"oral_water_qty"`
	OralOther                   sql.NullInt64   `json:"oral_other"`
	OralOtherQty                sql.NullFloat64 `json:"oral_other_qty"`
	Iv                          sql.NullInt64   `json:"iv"`
	IvQty                       sql.NullString  `json:"iv_qty"`
	IvUsing                     sql.NullString  `json:"iv_using"`
	IvAza                       sql.NullString  `json:"iv_aza"`
	AccessType                  sql.NullInt64   `json:"access_type"`
	BloodTrans                  sql.NullInt64   `json:"blood_trans"`
	OxygenTherapy               sql.NullInt64   `json:"oxygen_therapy"`
	OxygenQty                   sql.NullFloat64 `json:"oxygen_qty"`
	OxygenWith                  sql.NullString  `json:"oxygen_with"`
	Vasopressors                sql.NullInt64   `json:"vasopressors"`
	Renal                       sql.NullInt64   `json:"renal"`
	Invasive                    sql.NullInt64   `json:"invasive"`
	EbolaRdtAza                 sql.NullInt64   `json:"ebola_rdt_aza"`
	EbolaExperimentalIfZmap     sql.NullInt64   `json:"ebola_experimental_if_zmap"`
	EbolaExperimentalIfRemd     sql.NullInt64   `json:"ebola_experimental_if_remd"`
	EbolaExperimentalIfRegn     sql.NullInt64   `json:"ebola_experimental_if_regn"`
	EbolaExperimentalIfFavi     sql.NullInt64   `json:"ebola_experimental_if_favi"`
	EbolaExperimentalIfMab      sql.NullInt64   `json:"ebola_experimental_if_mab"`
	OralOtherAza                sql.NullString  `json:"oral_other_aza"`
	AntibacterialAza            sql.NullInt64   `json:"antibacterial_aza"`
	AntimalarialArtesunateDose  sql.NullString  `json:"antimalarial_artesunate_dose"`
	AntimalarialArtesunateRoute sql.NullString  `json:"antimalarial_artesunate_route"`
	AntimalarialArtesunateFreq  sql.NullString  `json:"antimalarial_artesunate_freq"`
	AntimalarialArthemeterDose  sql.NullString  `json:"antimalarial_arthemeter_dose"`
	AntimalarialArthemeterRoute sql.NullString  `json:"antimalarial_arthemeter_route"`
	AntimalarialArthemeterFreq  sql.NullString  `json:"antimalarial_arthemeter_freq"`
	AntimalarialAlDose          sql.NullString  `json:"antimalarial_al_dose"`
	AntimalarialAlRoute         sql.NullString  `json:"antimalarial_al_route"`
	AntimalarialAlFreq          sql.NullString  `json:"antimalarial_al_freq"`
	AntimalarialAaDose          sql.NullString  `json:"antimalarial_aa_dose"`
	AntimalarialAaRoute         sql.NullString  `json:"antimalarial_aa_route"`
	AntimalarialAaFreq          sql.NullString  `json:"antimalarial_aa_freq"`
	AmoxicillinDose             sql.NullString  `json:"amoxicillin_dose"`
	AmoxicillinRoute            sql.NullString  `json:"amoxicillin_route"`
	AmoxicillinFreq             sql.NullString  `json:"amoxicillin_freq"`
	CeftriaxoneDose             sql.NullString  `json:"ceftriaxone_dose"`
	CeftriaxoneRoute            sql.NullString  `json:"ceftriaxone_route"`
	CeftriaxoneFreq             sql.NullString  `json:"ceftriaxone_freq"`
	CefiximeDose                sql.NullString  `json:"cefixime_dose"`
	CefiximeRoute               sql.NullString  `json:"cefixime_route"`
	CefiximeFreq                sql.NullString  `json:"cefixime_freq"`
	AmpicillinDose              sql.NullString  `json:"ampicillin_dose"`
	AmpicillinRoute             sql.NullString  `json:"ampicillin_route"`
	AmpicillinFreq              sql.NullString  `json:"ampicillin_freq"`
	ChloramphenicolDose         sql.NullString  `json:"chloramphenicol_dose"`
	ChloramphenicolRoute        sql.NullString  `json:"chloramphenicol_route"`
	ChloramphenicolFreq         sql.NullString  `json:"chloramphenicol_freq"`
	AmoxiclavDose               sql.NullString  `json:"amoxiclav_dose"`
	AmoxiclavRoute              sql.NullString  `json:"amoxiclav_route"`
	AmoxiclavFreq               sql.NullString  `json:"amoxiclav_freq"`
	AzithromycinDose            sql.NullString  `json:"azithromycin_dose"`
	AzithromycinRoute           sql.NullString  `json:"azithromycin_route"`
	AzithromycinFreq            sql.NullString  `json:"azithromycin_freq"`
	CefotaximeDose              sql.NullString  `json:"cefotaxime_dose"`
	CefotaximeRoute             sql.NullString  `json:"cefotaxime_route"`
	CefotaximeFreq              sql.NullString  `json:"cefotaxime_freq"`
	CeftazidimeDose             sql.NullString  `json:"ceftazidime_dose"`
	CeftazidimeRoute            sql.NullString  `json:"ceftazidime_route"`
	CeftazidimeFreq             sql.NullString  `json:"ceftazidime_freq"`
	CiprofloxacinDose           sql.NullString  `json:"ciprofloxacin_dose"`
	CiprofloxacinRoute          sql.NullString  `json:"ciprofloxacin_route"`
	CiprofloxacinFreq           sql.NullString  `json:"ciprofloxacin_freq"`
	TetracyclineDose            sql.NullString  `json:"tetracycline_dose"`
	TetracyclineRoute           sql.NullString  `json:"tetracycline_route"`
	TetracyclineFreq            sql.NullString  `json:"tetracycline_freq"`
	ImipenemDose                sql.NullString  `json:"imipenem_dose"`
	ImipenemRoute               sql.NullString  `json:"imipenem_route"`
	ImipenemFreq                sql.NullString  `json:"imipenem_freq"`
	CotrimoxazoleDose           sql.NullString  `json:"cotrimoxazole_dose"`
	CotrimoxazoleRoute          sql.NullString  `json:"cotrimoxazole_route"`
	CotrimoxazoleFreq           sql.NullString  `json:"cotrimoxazole_freq"`
	GentamicinDose              sql.NullString  `json:"gentamicin_dose"`
	GentamicinRoute             sql.NullString  `json:"gentamicin_route"`
	GentamicinFreq              sql.NullString  `json:"gentamicin_freq"`
	MetronidazoleDose           sql.NullString  `json:"metronidazole_dose"`
	MetronidazoleRoute          sql.NullString  `json:"metronidazole_route"`
	MetronidazoleFreq           sql.NullString  `json:"metronidazole_freq"`
}

func HandlerCasesForm(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config config.Config, smsService *services.SMSService, voiceService *services.VoiceService) error {
	DoZaLogging("INFO", "Starting Client form", nil)

	userID, userName := GetUser(c, sl, store)
	role := security.GetRoleID(db, userID, "admin")
	id, err := strconv.Atoi(c.Params("i"))
	data := NewTemplateData(c, store)

	var client models.Client

	if err != nil || id < 1 {
		client.ID = 0
		data.IsIDPos = false
	} else {
		// Get the client
		clientModel, err := models.ClientByID(c.Context(), db, id)
		if err == nil && clientModel != nil {
			client = *clientModel
			data.IsIDPos = true

			// Check facility-based access control
			userFacility := GetCurrentFacility(c, db, sl, store)
			if userFacility > 0 {
				// User has a facility assigned, check if they can access this case
				if client.Site.Int64 != int64(userFacility) {
					sl.Error("User attempted to access case from different facility",
						"user_id", userID, "user_facility", userFacility, "case_site", client.Site.Int64, "case_id", id)
					return c.Status(403).SendString("Access denied: You can only access cases from your assigned facility.")
				}
			}
		} else {
			// Client not found, treat as new
			client.ID = 0
			data.IsIDPos = false
			sl.Warn("Client not found", "client_id", id, "error", err)
		}
	}

	// Get outbreak ID from session
	sess, err := store.Get(c)
	if err != nil {
		return c.Status(400).SendString("Failed to get session")
	}
	outbreakID := sess.Get("outbreak_id")
	if outbreakID == nil {
		return c.Status(400).SendString("No outbreak selected")
	}

	// Set outbreak ID for new cases
	data.OutbreakID = outbreakID.(int)
	data.IsOutbreakID = data.OutbreakID > 0

	// Get outbreak name
	outbreak, err := models.OutbreakByID(c.Context(), db, outbreakID.(int))
	if err == nil && outbreak != nil {
		if outbreak.Name.Valid {
			data.OutbreakName = outbreak.Name.String
		}
	}

	if client.ID == 0 {
		client.OutbreakID = sql.NullInt64{Int64: int64(outbreakID.(int)), Valid: true}
	}

	// Initialize empty slices to avoid nil issues
	var cE []models.ClientEncounter = []models.ClientEncounter{}
	var st []models.Status = []models.Status{}
	var assessments []*models.MpoxAssessment = []*models.MpoxAssessment{}

	if id > 0 && client.ID > 0 {
		cE, err = models.ClientEncounterz(c.Context(), db, "client_id="+strconv.Itoa(id), outbreakID.(int))
		if err != nil {
			DoZaLogging("ERROR", "Failed to get encounters", err)
		}
		if cE == nil {
			cE = []models.ClientEncounter{}
		}

		st, err = models.Statuses(c.Context(), db, "client_id="+strconv.Itoa(id))
		if err != nil {
			DoZaLogging("ERROR", "Failed to get statuses", err)
		}
		if st == nil {
			st = []models.Status{}
		}

		// Check if there is an Mpox admission for this client
		hasAdmission := false
		var admissionID int
		err = db.QueryRow("SELECT id FROM mpox_demographics WHERE client_id = $1 LIMIT 1", client.ID).Scan(&admissionID)
		if err == nil {
			hasAdmission = true
		}
		data.HasMpoxAdmission = hasAdmission
		data.MpoxAdmissionID = admissionID // int

		assessments, err = models.GetMpoxAssessmentsByClientID(context.Background(), db, client.ID)
		if err != nil {
			DoZaLogging("ERROR", "Failed to get Mpox assessments", err)
		}
		if assessments == nil {
			assessments = []*models.MpoxAssessment{}
		}
	}

	data.User = userName
	data.Role = role
	data.Optionz = Get_Client_Optionz()
	data.Form = client
	data.FormChild1 = cE
	data.FormChild2 = st
	data.FormChild3 = assessments
	if client.AdmWard.String == strconv.Itoa(3) {
		data.IsHomeBasedCare = true
	}
	// fmt.Printf("assessments data: %s", assessments)

	// Set user's facility ID for auto-selection in dropdown
	userFacility := GetCurrentFacility(c, db, sl, store)
	if userFacility > 0 {
		data.UserFacilityID = strconv.Itoa(userFacility)
	} else {
		data.UserFacilityID = ""
	}

	DoZaLogging("INFO", "Load Client form", err)
	return GenerateHTML(c, db, data, "form_patients")
}

// HandlerPatientProfile handles the patient profile view with biodata and clinical notes
func HandlerPatientProfile(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config config.Config, smsService *services.SMSService, voiceService *services.VoiceService) error {
	DoZaLogging("INFO", "Starting Patient Profile", nil)

	userID, userName := GetUser(c, sl, store)
	role := security.GetRoleID(db, userID, "admin")
	id, err := strconv.Atoi(c.Params("i"))
	if err != nil || id < 1 {
		return c.Status(400).SendString("Invalid patient ID")
	}

	data := NewTemplateData(c, store)

	// Get the client
	clientModel, err := models.ClientByID(c.Context(), db, id)
	if err != nil {
		sl.Warn("Failed to get patient", "client_id", id, "error", err)
		// Try to check if it's a no rows error
		if err == sql.ErrNoRows {
			return c.Status(404).SendString("Patient not found")
		}
		return c.Status(500).SendString("Failed to retrieve patient: " + err.Error())
	}
	if clientModel == nil {
		sl.Warn("Patient model is nil", "client_id", id)
		return c.Status(404).SendString("Patient not found")
	}
	client := *clientModel
	sl.Info("Patient found", "client_id", id, "patient_id", client.ID)

	// Check facility-based access control
	userFacility := GetCurrentFacility(c, db, sl, store)
	if userFacility > 0 {
		if client.Site.Int64 != int64(userFacility) {
			sl.Error("User attempted to access case from different facility",
				"user_id", userID, "user_facility", userFacility, "case_site", client.Site.Int64, "case_id", id)
			return c.Status(403).SendString("Access denied: You can only access cases from your assigned facility.")
		}
	}

	// Get outbreak ID from session
	sess, err := store.Get(c)
	if err != nil {
		return c.Status(400).SendString("Failed to get session")
	}
	outbreakID := sess.Get("outbreak_id")
	if outbreakID == nil {
		return c.Status(400).SendString("No outbreak selected")
	}

	// Get outbreak name
	outbreak, err := models.OutbreakByID(c.Context(), db, outbreakID.(int))
	if err == nil && outbreak != nil {
		if outbreak.Name.Valid {
			data.OutbreakName = outbreak.Name.String
		}
	}

	// Get all related data
	cE, err := models.ClientEncounterz(c.Context(), db, "client_id="+strconv.Itoa(id), outbreakID.(int))
	if err != nil {
		DoZaLogging("ERROR", "Failed to get encounters", err)
		cE = []models.ClientEncounter{}
	}

	// Fetch clinical, vital, and lab data for each encounter
	type EncounterDetail struct {
		Encounter models.ClientEncounter
		Clinical  *models.Clinical
		Vital     *models.Vital
		Lab       *models.Lab
	}
	var encounterDetails []EncounterDetail
	for _, enc := range cE {
		detail := EncounterDetail{Encounter: enc}
		if enc.EncounterID > 0 {
			// Fetch clinical data
			clinical, err := models.ClinicalByEncounterID(c.Context(), db, enc.EncounterID)
			if err == nil && clinical != nil {
				detail.Clinical = clinical
			}
			// Fetch vital data
			vital, err := models.VitalByEncounterID(c.Context(), db, enc.EncounterID)
			if err == nil && vital != nil {
				detail.Vital = vital
			}
			// Fetch lab data
			lab, err := models.LabByEncounterID(c.Context(), db, enc.EncounterID)
			if err == nil && lab != nil {
				detail.Lab = lab
			}
		}
		encounterDetails = append(encounterDetails, detail)
	}

	st, err := models.Statuses(c.Context(), db, "client_id="+strconv.Itoa(id))
	if err != nil {
		DoZaLogging("ERROR", "Failed to get statuses", err)
		st = []models.Status{}
	}

	assessments, err := models.GetMpoxAssessmentsByClientID(context.Background(), db, client.ID)
	if err != nil {
		DoZaLogging("ERROR", "Failed to get Mpox assessments", err)
		assessments = []*models.MpoxAssessment{}
	}

	// Fetch Mpox admission data if exists
	hasAdmission := false
	var admissionID int
	err = db.QueryRow("SELECT id FROM mpox_demographics WHERE client_id = $1 LIMIT 1", client.ID).Scan(&admissionID)
	if err == nil {
		hasAdmission = true
		sl.Info("Found admission record", "client_id", client.ID, "admission_id", admissionID)
	} else {
		if err != sql.ErrNoRows {
			sl.Warn("Error checking for admission", "client_id", client.ID, "error", err)
		}
	}

	// Create struct to hold admission data
	var admissionData AdmissionData

	if hasAdmission {
		// Fetch demographics
		var demo models.MpoxDemographics
		err = db.QueryRowContext(c.Context(), `
			SELECT id, client_id, sex, date_of_birth, age_years, age_months, age_days,
				health_care_worker, laboratory_worker, ppe_status, tribe, pregnant,
				gestational_weeks, lmnp, recently_pregnant, pregnant_22_42,
				tetanus_vaccination, occupation, site_of_first_encounter,
				site_of_first_encounter_other, suspect_confirmed_case,
				lymph_painful, lymph_location, lymph_other_detail,
				lymph_pain_location, lymph_pain_other_detail, created_at, updated_at
			FROM mpox_demographics WHERE id = $1
		`, admissionID).Scan(
			&demo.ID, &demo.ClientID, &demo.Sex, &demo.DateOfBirth, &demo.AgeYears, &demo.AgeMonths,
			&demo.AgeDays, &demo.HealthCareWorker, &demo.LaboratoryWorker, &demo.PPEStatus, &demo.Tribe,
			&demo.Pregnant, &demo.GestationalWeeks, &demo.LMNP, &demo.RecentlyPregnant, &demo.Pregnant22_42,
			&demo.TetanusVaccination, &demo.Occupation, &demo.SiteOfFirstEncounter, &demo.SiteOfFirstEncounterOther,
			&demo.SuspectConfirmedCase, &demo.LymphPainful, &demo.LymphLocation, &demo.LymphOtherDetail,
			&demo.LymphPainLocation, &demo.LymphPainOtherDetail, &demo.CreatedAt, &demo.UpdatedAt)
		if err == nil {
			admissionData.Demographics = &demo
			sl.Info("Loaded admission demographics", "admission_id", admissionID)

			// Fetch exposure history
			var exp models.MpoxExposureHistory
			err = db.QueryRowContext(c.Context(), `
				SELECT id, demographics_id, known_link, sexually_active, sex_of_partners,
					recent_travel, travel_high_risk, travel_details, created_at, updated_at
				FROM mpox_exposure_history WHERE demographics_id = $1 LIMIT 1
			`, admissionID).Scan(
				&exp.ID, &exp.DemographicsID, &exp.KnownLink, &exp.SexuallyActive, &exp.SexOfPartners,
				&exp.RecentTravel, &exp.TravelHighRisk, &exp.TravelDetails, &exp.CreatedAt, &exp.UpdatedAt)
			if err == nil {
				admissionData.Exposure = &exp
				sl.Info("Loaded admission exposure history", "admission_id", admissionID)
			} else if err != sql.ErrNoRows {
				sl.Warn("Error fetching exposure history", "admission_id", admissionID, "error", err)
			}

			// Fetch onset vitals
			var vitals models.MpoxOnsetVitals
			err = db.QueryRowContext(c.Context(), `
				SELECT id, demographics_id, symptom_onset, fever, temperature, heart_rate, 
					respiratory_rate, bp_systolic, bp_diastolic, dehydration, avpu, height_cm, weight_kg
				FROM mpox_onset_vitals WHERE demographics_id = $1 LIMIT 1
			`, admissionID).Scan(
				&vitals.ID, &vitals.DemographicsID, &vitals.SymptomOnset, &vitals.Fever,
				&vitals.Temperature, &vitals.HeartRate, &vitals.RespiratoryRate, &vitals.BpSystolic,
				&vitals.BpDiastolic, &vitals.Dehydration, &vitals.AVPU, &vitals.HeightCm, &vitals.WeightKg)
			if err == nil {
				admissionData.Vitals = &vitals
				sl.Info("Loaded admission vitals", "admission_id", admissionID)
			} else if err != sql.ErrNoRows {
				sl.Warn("Error fetching onset vitals", "admission_id", admissionID, "error", err)
			}
		} else {
			sl.Error("Error fetching admission demographics", "admission_id", admissionID, "error", err)
		}
	}

	sl.Info("Admission data prepared", "has_admission", hasAdmission, "has_demo", admissionData.Demographics != nil)

	data.User = userName
	data.Role = role
	data.Optionz = Get_Client_Optionz()
	data.Form = client
	data.FormChild1 = encounterDetails
	data.FormChild2 = st
	data.FormChild3 = assessments
	if hasAdmission {
		data.FormChild4 = admissionData
	} else {
		data.FormChild4 = nil
	}
	data.HasMpoxAdmission = hasAdmission
	data.MpoxAdmissionID = admissionID

	// Debug logging
	if hasAdmission {
		sl.Info("Admission data status",
			"has_demo", admissionData.Demographics != nil,
			"has_exposure", admissionData.Exposure != nil,
			"has_vitals", admissionData.Vitals != nil,
			"admission_id", admissionID)
	}
	data.OutbreakID = outbreakID.(int)
	data.IsIDPos = true
	if client.AdmWard.String == strconv.Itoa(3) {
		data.IsHomeBasedCare = true
	}

	DoZaLogging("INFO", "Load Patient Profile", err)
	return GenerateHTML(c, db, data, "patient_profile")
}

func HandlerCasesSubmit(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config config.Config) error {

	id, er := strconv.Atoi(c.FormValue("id"))
	if er != nil {
		id = 0
	}

	// Get outbreak ID from session
	sess, err := store.Get(c)
	if err != nil {
		return c.Status(500).SendString("Session error")
	}

	var outbreakID sql.NullInt64
	outbreakIDFromSession := sess.Get("outbreak_id")
	if outbreakIDFromSession != nil {
		switch v := outbreakIDFromSession.(type) {
		case int:
			outbreakID.Int64 = int64(v)
			outbreakID.Valid = true
		case int64:
			outbreakID.Int64 = v
			outbreakID.Valid = true
		case float64:
			outbreakID.Int64 = int64(v)
			outbreakID.Valid = true
		case string:
			if id, err := strconv.Atoi(v); err == nil {
				outbreakID.Int64 = int64(id)
				outbreakID.Valid = true
			}
		}
	}

	client := models.Client{
		ID:               id,
		Firstname:        ParseNullString(c.FormValue("firstname")),
		Lastname:         ParseNullString(c.FormValue("lastname")),
		Othername:        ParseNullString(c.FormValue("othername")),
		Gender:           ParseNullInt(c.FormValue("gender")),
		DateOfBirth:      ParseNullString(c.FormValue("date_of_birth")),
		Age:              ParseNullFloat(c.FormValue("age")),
		Marital:          ParseNullInt(c.FormValue("marital")),
		Nin:              ParseNullString(c.FormValue("nin")),
		Nationality:      ParseNullInt(c.FormValue("nationality")),
		AdmDate:          ParseNullString(c.FormValue("adm_date")),
		AdmFrom:          ParseNullString(c.FormValue("adm_from")),
		LabNo:            ParseNullString(c.FormValue("lab_no")),
		CifNo:            ParseNullString(c.FormValue("cif_no")),
		EtuNo:            ParseNullString(c.FormValue("etu_no")),
		CaseNo:           ParseNullString(c.FormValue("case_no")),
		Occupation:       ParseNullInt(c.FormValue("occupation")),
		OccupationAza:    ParseNullString(c.FormValue("occupation_aza")),
		DateSymptomOnset: ParseNullString(c.FormValue("date_symptom_onset")),
		DateIsolation:    ParseNullString(c.FormValue("date_isolation")),
		Pregnant:         ParseNullInt(c.FormValue("pregnant")),
		AdmWard:          ParseNullString(c.FormValue("adm_ward")),
		Tb:               ParseNullInt(c.FormValue("tb")),
		Asplenia:         ParseNullInt(c.FormValue("asplenia")),
		Hep:              ParseNullInt(c.FormValue("hep")),
		Diabetes:         ParseNullInt(c.FormValue("diabetes")),
		Hiv:              ParseNullInt(c.FormValue("hiv")),
		Liver:            ParseNullInt(c.FormValue("liver")),
		Malignancy:       ParseNullInt(c.FormValue("malignancy")),
		Heart:            ParseNullInt(c.FormValue("heart")),
		Pulmonary:        ParseNullInt(c.FormValue("pulmonary")),
		Kidney:           ParseNullInt(c.FormValue("kidney")),
		Neurologic:       ParseNullInt(c.FormValue("neurologic")),
		Other:            ParseNullString(c.FormValue("other")),
		Transfer:         ParseNullInt(c.FormValue("transfer")),
		Site:             ParseNullInt(c.FormValue("site")),
		Status:           ParseNullString(c.FormValue("status")),
		OutbreakID:       outbreakID,
		HbcPhone:         ParseNullString(c.FormValue("hbc_phone")),
		HbcFollowup:      ParseNullString(c.FormValue("hbc_followup")),
		HbcLanguage:      ParseNullInt(c.FormValue("hbc_language")),
	}

	//visID, _ := utilities.GetSequentialVisitID()
	userID := GetCurrentUser(c, store)

	client.EditOn.Valid = true
	client.EditBy.Valid = true

	client.EditBy.Int64 = int64(userID)
	client.EditOn.Time = time.Now()

	if client.ID == 0 {

		client.EnterOn.Valid = true
		client.EnterBy.Valid = true

		client.EnterBy.Int64 = int64(userID)
		client.EnterOn.Time = time.Now()

		client.UUID.Valid = true
		client.UUID.String = models.CreateUUID()

		//appID := models.CreateUUID()
		//client.UUID.String = appID

	}

	if client.ID == 0 {
		err := client.Insert(c.Context(), db)
		if err != nil {
			sl.Error("Failed to insert client", "error", err)
			return c.Status(500).SendString("Failed to save patient: " + err.Error())
		}
		sl.Info("Client inserted successfully", "client_id", client.ID)
	} else {
		client.SetAsExists()
		err := client.Update(c.Context(), db)
		if err != nil {
			sl.Error("Failed to update client", "error", err, "client_id", client.ID)
			return c.Status(500).SendString("Failed to update patient: " + err.Error())
		}
		sl.Info("Client updated successfully", "client_id", client.ID)
	}

	urlx := "/cases/new/" + strconv.Itoa(client.ID)

	return c.Redirect(urlx)
}

func HandlerCasesList(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config config.Config) error {
	fmt.Println("starting case list")

	userID, userName := GetUser(c, sl, store)
	role := security.GetRoleID(db, userID, "admin")

	data := NewTemplateData(c, store)
	data.User = userName
	data.Role = role

	// Ensure Optionz is properly initialized
	if data.Optionz == nil {
		data.Optionz = make(map[string]map[string]string)
		data.Optionz["yn"] = map[string]string{"": " -- ", "1": "Yes", "2": "No"}
		data.Optionz["yn_extra"] = map[string]string{"": " -- ", "1": "Yes", "2": "No", "3": "Unknown"}
		data.Optionz["sex"] = map[string]string{"": " -- ", "1": "Male", "2": "Female"}
		data.Optionz["marital"] = map[string]string{"": " -- ", "1": "Married", "2": "Cohabiting", "3": "Widowed", "4": "Separated", "5": "Divorced", "6": "Single"}
		data.Optionz["nationality"] = map[string]string{"": " -- ", "1": "Ugandan", "2": "EAC", "3": "Other"}
		data.Optionz["mental"] = map[string]string{"": " -- ", "a": "A", "v": "V", "p": "P", "u": "U"}
		data.Optionz["preg"] = map[string]string{"": " -- ", "1": "Yes", "2": "No", "3": "ND"}
		data.Optionz["ward"] = map[string]string{"": " -- ", "1": "Ward", "2": "ICU"}
		data.Optionz["result1"] = map[string]string{"": " -- ", "1": "Pos", "2": "Neg", "3": "indeterminate"}
		data.Optionz["result2"] = map[string]string{"": " -- ", "1": "Pos", "2": "Neg", "3": "ND"}
		data.Optionz["language"] = map[string]string{"": " -- ", "1": "English", "2": "Luganda", "3": "Other"}
	}

	fmt.Println("loading case list page")

	// Get outbreak ID from session using helper function
	outbreakID := GetCurrentOutbreak(c, store)
	fmt.Printf("DEBUG: HandlerCasesList - User %d, outbreakID from session: %d\n", userID, outbreakID)

	if outbreakID == 0 {
		sl.Error("No outbreak selected for user", "user_id", userID)
		return c.Status(400).SendString("No outbreak selected. Please select an outbreak first.")
	}

	sl.Info("Loading cases for outbreak", "user_id", userID, "outbreak_id", outbreakID)

	// Get user's facility from database
	userFacility := GetCurrentFacility(c, db, sl, store)

	// Build filter based on outbreak and facility
	filter := fmt.Sprintf("outbreak_id = %d", outbreakID)

	// Admins (admin, super_admin) should see all facilities
	isAdmin := security.HasAnyRole(db, userID, []string{"admin", "super_admin"})
	if isAdmin {
		sl.Info("Admin user - bypassing facility filter", "user_id", userID)
	} else {
		// If user has a facility assigned, filter by that facility (include NULL site so unassigned cases in outbreak appear)
		if userFacility > 0 {
			filter += fmt.Sprintf(" AND (site = %d OR site IS NULL)", userFacility)
			sl.Info("Filtering cases by user facility", "user_id", userID, "facility_id", userFacility)
		} else {
			sl.Info("No facility assigned to user, showing all cases for outbreak", "user_id", userID)
		}
	}

	clients, err := models.Clients(c.Context(), db, filter)
	if err != nil {
		sl.Error("Error loading case list", "error", err, "outbreak_id", outbreakID, "user_facility", userFacility)
		clients = []models.Client{}
	}
	if clients == nil {
		clients = []models.Client{}
	}

	var phoneNumbers []string
	for _, client := range clients {
		fmt.Println("Client ID: ", client.ID, "Adm ward:", client.AdmWard.String, " HBC Followup: ", client.HbcFollowup.String, " HBC Phone: ", client.HbcPhone.String)
		if client.HbcFollowup.String == "ivr" && client.HbcPhone.String != "" {
			phoneNumbers = append(phoneNumbers, client.HbcPhone.String)
		}
	}
	phoneString := strings.Join(phoneNumbers, ", ")
	// fmt.Println("IVR Phone Numbers String: ", phoneString)
	// fmt.Println("IVR Phone Numbers original: ", phoneNumbers)
	// var ivrPhoneNumbers []string
	// ivrPhoneNumbers = append(ivrPhoneNumbers, phoneString)
	// fmt.Println("IVR Phone Numbers Slice: ", ivrPhoneNumbers)
	// fmt.Println("IVR Phone Numbers: ", ivrPhoneNumbers)

	data.Form = clients
	data.FormChild1 = phoneString

	return GenerateHTML(c, db, data, "list_patients")
}

func HandlerCaseEncounterForm(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config config.Config) error {
	// Get client ID from URL path parameter
	clientIDStr := c.Params("i")
	if clientIDStr == "" {
		sl.Error("Client ID is missing from URL path")
		return c.Status(400).SendString("Client ID is required")
	}

	// Convert client ID to int
	clientID, err := strconv.Atoi(clientIDStr)
	if err != nil {
		sl.Error("Invalid client ID", "error", err, "clientID", clientIDStr)
		return c.Status(400).SendString("Invalid client ID")
	}

	// Get outbreak ID from session or query parameter
	sess, err := store.Get(c)
	if err != nil {
		sl.Error("Failed to get session", "error", err)
		return c.Status(500).SendString("Failed to get session")
	}

	// Check if outbreak_id is provided in query parameter
	queryOutbreakID := c.Query("outbreak_id")
	var outbreakID interface{}

	if queryOutbreakID != "" {
		// Convert query parameter to int
		outbreakIDInt, err := strconv.Atoi(queryOutbreakID)
		if err != nil {
			sl.Error("Invalid outbreak_id in query parameter", "error", err, "outbreak_id", queryOutbreakID)
			return c.Status(400).SendString("Invalid outbreak_id")
		}
		outbreakID = outbreakIDInt

		// Update session with the new outbreak_id
		sess.Set("outbreak_id", outbreakIDInt)
		sess.Set("selected_outbreak", outbreakIDInt)
		if err := sess.Save(); err != nil {
			sl.Error("Failed to save outbreak_id to session", "error", err)
		}

		sl.Info("Updated outbreak_id from query parameter", "outbreak_id", outbreakIDInt)
	} else {
		// Get outbreak ID from session
		outbreakID = sess.Get("outbreak_id")
		if outbreakID == nil {
			sl.Error("No outbreak selected")
			return c.Status(400).SendString("No outbreak selected")
		}
	}

	// Get encounter date from query parameter
	encounterDate := c.Query("dte")

	// Validate date format
	if encounterDate == "" || encounterDate == "0000-00-00" {
		// If no date provided or invalid date, use current date
		encounterDate = time.Now().Format("2006-01-02")
	} else {
		// Try to parse the date to validate it
		if _, err := time.Parse("2006-01-02", encounterDate); err != nil {
			// If date is invalid, use current date
			encounterDate = time.Now().Format("2006-01-02")
		}
	}

	// Get client details first
	client, err := models.ClientByID(c.Context(), db, clientID)
	if err != nil {
		sl.Error("Failed to get client", "error", err, "clientID", clientID)
		return c.Status(500).SendString("Failed to get client details")
	}

	// Check facility-based access control
	userID := GetCurrentUser(c, store)
	userFacility := GetCurrentFacility(c, db, sl, store)
	// Allow admins to access any case regardless of facility
	isAdmin := security.HasAnyRole(db, userID, []string{"admin", "super_admin"})
	if !isAdmin {
		if userFacility > 0 {
			// User has a facility assigned, check if they can access this case
			if client.Site.Int64 != int64(userFacility) {
				sl.Error("User attempted to access case from different facility",
					"user_id", userID, "user_facility", userFacility, "case_site", client.Site.Int64, "case_id", clientID)
				return c.Status(403).SendString("Access denied: You can only access cases from your assigned facility.")
			}
		}
	}

	// Get all encounters for this client (not filtered by date)
	sl.Info("Fetching all encounters", "clientID", clientID, "outbreakID", outbreakID.(int))
	allEncounters, err := models.ClientEncounters(c.Context(), db, fmt.Sprintf("client_id = %d", clientID), outbreakID.(int))
	if err != nil {
		sl.Error("Failed to get all encounters", "error", err, "clientID", clientID)
		allEncounters = []models.ClientEncounter{}
	} else {
		sl.Info("Successfully fetched encounters", "count", len(allEncounters), "clientID", clientID)
	}

	// Get encounters for the specific date (for editing existing encounters)
	dateEncounters, err := models.ClientEncounters(c.Context(), db, fmt.Sprintf("client_id = %d AND encounter_date = '%s'", clientID, encounterDate), outbreakID.(int))
	if err != nil {
		sl.Error("Failed to get encounters for date", "error", err, "clientID", clientID, "date", encounterDate)
		dateEncounters = []models.ClientEncounter{}
	}

	// Use date-specific encounters if they exist, otherwise create empty encounter
	var encounters []models.ClientEncounter
	if len(dateEncounters) > 0 {
		encounters = dateEncounters
	} else {
		// Create an empty encounter with the current date
		emptyEncounter := models.ClientEncounter{
			EncounterID:   0,
			EncounterType: sql.NullInt64{Int64: 0, Valid: false},
			EmployeeFname: sql.NullString{String: "", Valid: true},
			EmployeeLname: sql.NullString{String: "", Valid: true},
			EncounterDate: sql.NullString{String: encounterDate, Valid: true},
			EncounterTime: sql.NullString{String: "", Valid: true},
			ClinicalTeam:  sql.NullString{String: "", Valid: true},
			ManagedBy:     sql.NullInt64{Int64: 0, Valid: false},
			ClientID:      clientID,
		}
		encounters = append(encounters, emptyEncounter)
	}

	// Ensure we have 3 encounters for morning, afternoon, evening
	for len(encounters) < 3 {
		emptyEncounter := models.ClientEncounter{
			EncounterID:   0,
			EncounterType: sql.NullInt64{Int64: 0, Valid: false},
			EmployeeFname: sql.NullString{String: "", Valid: true},
			EmployeeLname: sql.NullString{String: "", Valid: true},
			EncounterDate: sql.NullString{String: encounterDate, Valid: true},
			EncounterTime: sql.NullString{String: "", Valid: true},
			ClinicalTeam:  sql.NullString{String: "", Valid: true},
			ManagedBy:     sql.NullInt64{Int64: 0, Valid: false},
			ClientID:      clientID,
		}
		encounters = append(encounters, emptyEncounter)
	}

	// Get clinical data for the first encounter
	var clinical []models.Clinical
	if len(encounters) > 0 && encounters[0].EncounterID > 0 {
		clinicalData, err := models.ClinicalByEncounterID(c.Context(), db, encounters[0].EncounterID)
		if err == nil && clinicalData != nil {
			clinical = append(clinical, *clinicalData)
		}
	}
	if len(clinical) == 0 {
		// Add empty clinical data
		clinical = append(clinical, models.Clinical{
			ClinicalID:            0,
			PharyngealErythema:    sql.NullInt64{Int64: 0, Valid: true},
			PharyngealExudate:     sql.NullInt64{Int64: 0, Valid: true},
			ConjunctivalInjection: sql.NullInt64{Int64: 0, Valid: true},
			OedemaFace:            sql.NullInt64{Int64: 0, Valid: true},
			TenderAbdomen:         sql.NullInt64{Int64: 0, Valid: true},
			SunkenEyes:            sql.NullInt64{Int64: 0, Valid: true},
			TentingSkin:           sql.NullInt64{Int64: 0, Valid: true},
			PalpableLiver:         sql.NullInt64{Int64: 0, Valid: true},
			PalpableSpleen:        sql.NullInt64{Int64: 0, Valid: true},
			Jaundice:              sql.NullInt64{Int64: 0, Valid: true},
			EnlargedLymphNodes:    sql.NullInt64{Int64: 0, Valid: true},
			LowerExtremityOedema:  sql.NullInt64{Int64: 0, Valid: true},
			Bleeding:              sql.NullInt64{Int64: 0, Valid: true},
			BleedingNose:          sql.NullInt64{Int64: 0, Valid: true},
			BleedingMouth:         sql.NullInt64{Int64: 0, Valid: true},
			BleedingVagina:        sql.NullInt64{Int64: 0, Valid: true},
			BleedingRectum:        sql.NullInt64{Int64: 0, Valid: true},
			Shock:                 sql.NullInt64{Int64: 0, Valid: true},
			Meningitis:            sql.NullInt64{Int64: 0, Valid: true},
			Confusion:             sql.NullInt64{Int64: 0, Valid: true},
			Seizure:               sql.NullInt64{Int64: 0, Valid: true},
			Coma:                  sql.NullInt64{Int64: 0, Valid: true},
			Bacteraemia:           sql.NullInt64{Int64: 0, Valid: true},
			Hyperglycemia:         sql.NullInt64{Int64: 0, Valid: true},
			Hypoglycemia:          sql.NullInt64{Int64: 0, Valid: true},
		})
	}
	// Ensure we have 3 clinical records for morning, afternoon, evening
	for len(clinical) < 3 {
		clinical = append(clinical, models.Clinical{
			ClinicalID:            0,
			PharyngealErythema:    sql.NullInt64{Int64: 0, Valid: true},
			PharyngealExudate:     sql.NullInt64{Int64: 0, Valid: true},
			ConjunctivalInjection: sql.NullInt64{Int64: 0, Valid: true},
			OedemaFace:            sql.NullInt64{Int64: 0, Valid: true},
			TenderAbdomen:         sql.NullInt64{Int64: 0, Valid: true},
			SunkenEyes:            sql.NullInt64{Int64: 0, Valid: true},
			TentingSkin:           sql.NullInt64{Int64: 0, Valid: true},
			PalpableLiver:         sql.NullInt64{Int64: 0, Valid: true},
			PalpableSpleen:        sql.NullInt64{Int64: 0, Valid: true},
			Jaundice:              sql.NullInt64{Int64: 0, Valid: true},
			EnlargedLymphNodes:    sql.NullInt64{Int64: 0, Valid: true},
			LowerExtremityOedema:  sql.NullInt64{Int64: 0, Valid: true},
			Bleeding:              sql.NullInt64{Int64: 0, Valid: true},
			BleedingNose:          sql.NullInt64{Int64: 0, Valid: true},
			BleedingMouth:         sql.NullInt64{Int64: 0, Valid: true},
			BleedingVagina:        sql.NullInt64{Int64: 0, Valid: true},
			BleedingRectum:        sql.NullInt64{Int64: 0, Valid: true},
			Shock:                 sql.NullInt64{Int64: 0, Valid: true},
			Meningitis:            sql.NullInt64{Int64: 0, Valid: true},
			Confusion:             sql.NullInt64{Int64: 0, Valid: true},
			Seizure:               sql.NullInt64{Int64: 0, Valid: true},
			Coma:                  sql.NullInt64{Int64: 0, Valid: true},
			Bacteraemia:           sql.NullInt64{Int64: 0, Valid: true},
			Hyperglycemia:         sql.NullInt64{Int64: 0, Valid: true},
			Hypoglycemia:          sql.NullInt64{Int64: 0, Valid: true},
		})
	}

	// Get vitals data for the first encounter
	var vitals []models.Vital
	if len(encounters) > 0 && encounters[0].EncounterID > 0 {
		vitalsData, err := models.VitalByEncounterID(c.Context(), db, encounters[0].EncounterID)
		if err == nil && vitalsData != nil {
			vitals = append(vitals, *vitalsData)
		}
	}
	if len(vitals) == 0 {
		// Add empty vitals data
		vitals = append(vitals, models.Vital{
			VitalsID:        0,
			HeartRate:       sql.NullFloat64{Float64: 0, Valid: true},
			BpSystolic:      sql.NullFloat64{Float64: 0, Valid: true},
			BpDiastolic:     sql.NullFloat64{Float64: 0, Valid: true},
			CapillaryRefill: sql.NullInt64{Int64: 0, Valid: true},
			RespiratoryRate: sql.NullFloat64{Float64: 0, Valid: true},
			Saturation:      sql.NullFloat64{Float64: 0, Valid: true},
			Weight:          sql.NullFloat64{Float64: 0, Valid: true},
			Height:          sql.NullFloat64{Float64: 0, Valid: true},
			Temperature:     sql.NullFloat64{Float64: 0, Valid: true},
			Muac:            sql.NullFloat64{Float64: 0, Valid: true},
		})
	}
	// Ensure we have 3 vitals records for morning, afternoon, evening
	for len(vitals) < 3 {
		vitals = append(vitals, models.Vital{
			VitalsID:        0,
			HeartRate:       sql.NullFloat64{Float64: 0, Valid: true},
			BpSystolic:      sql.NullFloat64{Float64: 0, Valid: true},
			BpDiastolic:     sql.NullFloat64{Float64: 0, Valid: true},
			CapillaryRefill: sql.NullInt64{Int64: 0, Valid: true},
			RespiratoryRate: sql.NullFloat64{Float64: 0, Valid: true},
			Saturation:      sql.NullFloat64{Float64: 0, Valid: true},
			Weight:          sql.NullFloat64{Float64: 0, Valid: true},
			Height:          sql.NullFloat64{Float64: 0, Valid: true},
			Temperature:     sql.NullFloat64{Float64: 0, Valid: true},
			Muac:            sql.NullFloat64{Float64: 0, Valid: true},
		})
	}

	// Get lab data for the first encounter
	var labs []models.Lab
	if len(encounters) > 0 && encounters[0].EncounterID > 0 {
		labData, err := models.LabByEncounterID(c.Context(), db, encounters[0].EncounterID)
		if err == nil && labData != nil {
			labs = append(labs, *labData)
		}
	}
	if len(labs) == 0 {
		// Add empty lab data
		labs = append(labs, models.Lab{
			LabID: 0,
		})
	}
	// Ensure we have at least 1 lab record
	for len(labs) < 1 {
		labs = append(labs, models.Lab{
			LabID: 0,
		})
	}

	// Get treatment data for the first encounter
	var treatments []FullTreatmentData
	if len(encounters) > 0 && encounters[0].EncounterID > 0 {
		treatmentData, err := loadFullTreatmentData(c.Context(), db, encounters[0].EncounterID)
		if err == nil && treatmentData != nil {
			treatments = append(treatments, *treatmentData)
		}
	}
	if len(treatments) == 0 {
		// Add empty treatment data
		treatments = append(treatments, FullTreatmentData{
			TreatmentID: 0,
		})
	}
	// Ensure we have at least 1 treatment record
	for len(treatments) < 1 {
		treatments = append(treatments, FullTreatmentData{
			TreatmentID: 0,
		})
	}

	// Determine Mpox admission presence for this client
	hasAdmission := false
	admissionID := 0
	if client != nil {
		var admID int
		if err := db.QueryRowContext(c.Context(), "SELECT id FROM mpox_demographics WHERE client_id = $1 LIMIT 1", client.ID).Scan(&admID); err == nil {
			hasAdmission = true
			admissionID = admID
		}
	}

	// Prepare strongly typed data for the template
	data := EncounterPageData{
		FormRef:          *client,
		Form:             encounters,
		Date:             encounterDate,
		FormChild1:       clinical,
		FormChild2:       vitals,
		FormChild3:       labs,
		FormChild4:       treatments,
		AllEncounters:    allEncounters, // Add all encounters
		Optionz:          Get_Client_Optionz(),
		OutbreakID:       outbreakID.(int), // Add OutbreakID for templates
		HasMpoxAdmission: hasAdmission,
		MpoxAdmissionID:  admissionID,
	}

	// Get filtered employees based on user's facility
	var filteredEmployees []struct {
		ID   int64  `json:"id"`
		Name string `json:"name"`
	}

	if userFacility > 0 {
		// Query employees from the same facility
		rows, err := db.QueryContext(c.Context(), `
			SELECT employee_id, CONCAT(employee_fname, ' ', employee_lname) as full_name 
			FROM employee 
			WHERE facility = $1 AND employee_status = 'active'
			ORDER BY employee_fname, employee_lname
		`, userFacility)
		if err != nil {
			sl.Error("Failed to get employees for facility", "error", err, "facility", userFacility)
		} else {
			defer rows.Close()
			for rows.Next() {
				var emp struct {
					ID   int64  `json:"id"`
					Name string `json:"name"`
				}
				if err := rows.Scan(&emp.ID, &emp.Name); err == nil {
					filteredEmployees = append(filteredEmployees, emp)
				}
			}
		}
	} else {
		// If no facility assigned, get all active employees
		rows, err := db.QueryContext(c.Context(), `
			SELECT employee_id, CONCAT(employee_fname, ' ', employee_lname) as full_name 
			FROM employee 
			WHERE employee_status = 'active'
			ORDER BY employee_fname, employee_lname
		`)
		if err != nil {
			sl.Error("Failed to get all employees", "error", err)
		} else {
			defer rows.Close()
			for rows.Next() {
				var emp struct {
					ID   int64  `json:"id"`
					Name string `json:"name"`
				}
				if err := rows.Scan(&emp.ID, &emp.Name); err == nil {
					filteredEmployees = append(filteredEmployees, emp)
				}
			}
		}
	}

	// Add filtered employees to the data
	data.FilteredEmployees = filteredEmployees

	// Debug: Log the client data being passed to template
	sl.Info("Client data for template",
		"clientID", client.ID,
		"firstName", client.Firstname.String,
		"lastName", client.Lastname.String,
		"genderValid", client.Gender.Valid,
		"genderValue", client.Gender.Int64,
		"outbreakIDValid", client.OutbreakID.Valid,
		"outbreakIDValue", client.OutbreakID.Int64,
		"filteredEmployeesCount", len(filteredEmployees),
	)

	return GenerateHTML(c, db, data, "form_encounters")
}

func HandlerCaseEncounterList(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config config.Config) error {
	// Get client ID from query parameter
	ClientIDStr := c.Query("client_id")
	if ClientIDStr == "" {
		return c.Status(400).SendString("Client ID is required")
	}

	// Convert client ID to int
	ClientID, err := strconv.Atoi(ClientIDStr)
	if err != nil {
		return c.Status(400).SendString("Invalid client ID")
	}

	// Get outbreak ID from session
	sess, err := store.Get(c)
	if err != nil {
		return c.Status(400).SendString("Failed to get session")
	}
	outbreakID := sess.Get("outbreak_id")
	if outbreakID == nil {
		return c.Status(400).SendString("No outbreak selected")
	}

	// Get encounters for this client
	encounters, err := models.ClientEncounterz(c.Context(), db, fmt.Sprintf("client_id = %d", ClientID), outbreakID.(int))
	if err != nil {
		sl.Error("Failed to get encounters", "error", err)
		return c.Status(500).SendString("Failed to get encounters")
	}

	// Get client details
	client, err := models.ClientByID(c.Context(), db, ClientID)
	if err != nil {
		sl.Error("Failed to get client", "error", err)
		return c.Status(500).SendString("Failed to get client")
	}

	data := NewTemplateData(c, store)
	data.Form = fiber.Map{
		"Client":     client,
		"Encounters": encounters,
	}
	data.Optionz = Get_Client_Optionz()

	return GenerateHTML(c, db, data, "list_case_encounter")
}

func saveEncounter(c *fiber.Ctx, db *sql.DB, userID int, cid, dte string, outbreakID int) (int, int, int, error) {
	// Convert client ID to int
	clientID, err := strconv.Atoi(cid)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid client ID: %v", err)
	}

	// Create encounter
	ClinicalTeam := "0"
	encounter := models.Encounter{
		EncounterType: sql.NullInt64{Int64: 1, Valid: true}, // Assuming 1 is the default encounter type
		EncounterTime: sql.NullString{String: time.Now().Format("15:04:05"), Valid: true},
		ClientID:      sql.NullInt64{Int64: int64(clientID), Valid: true},
		EncounterDate: sql.NullString{String: dte, Valid: true},
		ManagedBy:     sql.NullInt64{Int64: int64(userID), Valid: true},
		EnterOn:       sql.NullTime{Time: time.Now(), Valid: true},
		EnterBy:       sql.NullInt64{Int64: int64(userID), Valid: true},
		OutbreakID:    sql.NullInt64{Int64: int64(outbreakID), Valid: true},
		ClinicalTeam:  sql.NullString{String: ClinicalTeam, Valid: true},
	}

	// Save encounter
	if err := encounter.Insert(c.Context(), db); err != nil {
		return 0, 0, 0, fmt.Errorf("failed to save encounter: %v", err)
	}

	return encounter.EncounterID, clientID, outbreakID, nil
}

func saveVitals(c *fiber.Ctx, db *sql.DB, id1, id2, id3 int) error {
	vital_id, err := strconv.Atoi(c.FormValue("vital_id"))
	if err != nil {
		vital_id = 0
	}

	vital := models.Vital{
		VitalsID:            vital_id,
		EncounterID:         sql.NullInt64{Int64: int64(id1), Valid: true},
		HeartRate:           ParseNullFloat(c.FormValue("heart_rate")),
		BpSystolic:          ParseNullFloat(c.FormValue("bp_systolic")),
		BpDiastolic:         ParseNullFloat(c.FormValue("bp_diastolic")),
		CapillaryRefill:     ParseNullInt(c.FormValue("capillary_refill")),
		RespiratoryRate:     ParseNullFloat(c.FormValue("respiratory_rate")),
		Saturation:          ParseNullFloat(c.FormValue("saturation")),
		Weight:              ParseNullFloat(c.FormValue("weight")),
		Height:              ParseNullFloat(c.FormValue("height")),
		Temperature:         ParseNullFloat(c.FormValue("temperature")),
		LowestConsciousness: ParseNullString(c.FormValue("lowest_consciousness")),
		MentalStatus:        ParseNullString(c.FormValue("mental_status")),
		Muac:                ParseNullFloat(c.FormValue("muac")),
		Bleeding:            ParseNullInt(c.FormValue("bleeding")),
		Shock:               ParseNullInt(c.FormValue("shock")),
		Meningitis:          ParseNullInt(c.FormValue("meningitis")),
		Confusion:           ParseNullInt(c.FormValue("confusion")),
		Seizure:             ParseNullInt(c.FormValue("seizure")),
		Coma:                ParseNullInt(c.FormValue("coma")),
		Bacteraemia:         ParseNullInt(c.FormValue("bacteraemia")),
		Hyperglycemia:       ParseNullInt(c.FormValue("hyperglycemia")),
		Hypoglycemia:        ParseNullInt(c.FormValue("hypoglycemia")),
		Other:               ParseNullString(c.FormValue("other")),
	}

	if vital_id == 0 {
		return vital.Insert(c.Context(), db)
	} else {
		vital.SetAsExists()
		return vital.Update(c.Context(), db)
	}
}

func getZaFormValue(c *fiber.Ctx, zname string, i int) string {
	return c.FormValue(fmt.Sprintf("%s%d", zname, i))
}

func saveClinical(c *fiber.Ctx, db *sql.DB, id1, id2, id3 int) error {
	clinical_id, err := strconv.Atoi(c.FormValue("clinical_id"))
	if err != nil {
		clinical_id = 0
	}

	clinical := models.Clinical{
		ClinicalID:              clinical_id,
		EncounterID:             sql.NullInt64{Int64: int64(id1), Valid: true},
		Fever:                   ParseNullInt(c.FormValue("fever")),
		Fatigue:                 ParseNullInt(c.FormValue("fatigue")),
		Weakness:                ParseNullInt(c.FormValue("weakness")),
		Malaise:                 ParseNullInt(c.FormValue("malaise")),
		Myalgia:                 ParseNullInt(c.FormValue("myalgia")),
		Anorexia:                ParseNullInt(c.FormValue("anorexia")),
		SoreThroat:              ParseNullInt(c.FormValue("sore_throat")),
		Headache:                ParseNullInt(c.FormValue("headache")),
		Nausea:                  ParseNullInt(c.FormValue("nausea")),
		ChestPain:               ParseNullInt(c.FormValue("chest_pain")),
		JointPain:               ParseNullInt(c.FormValue("joint_pain")),
		Hiccups:                 ParseNullInt(c.FormValue("hiccups")),
		Cough:                   ParseNullInt(c.FormValue("cough")),
		DifficultyBreathing:     ParseNullInt(c.FormValue("difficulty_breathing")),
		DifficultySwallowing:    ParseNullInt(c.FormValue("difficulty_swallowing")),
		AbdominalPain:           ParseNullInt(c.FormValue("abdominal_pain")),
		Diarrhoea:               ParseNullInt(c.FormValue("diarrhoea")),
		Vomiting:                ParseNullInt(c.FormValue("vomiting")),
		Irritability:            ParseNullInt(c.FormValue("irritability")),
		Dysphagia:               ParseNullInt(c.FormValue("dysphagia")),
		UnusualBleeding:         ParseNullInt(c.FormValue("unusual_bleeding")),
		Dehydration:             ParseNullInt(c.FormValue("dehydration")),
		Shock:                   ParseNullInt(c.FormValue("shock")),
		Anuria:                  ParseNullInt(c.FormValue("anuria")),
		Disorientation:          ParseNullInt(c.FormValue("disorientation")),
		Agitation:               ParseNullInt(c.FormValue("agitation")),
		Seizure:                 ParseNullInt(c.FormValue("seizure")),
		Meningitis:              ParseNullInt(c.FormValue("meningitis")),
		Confusion:               ParseNullInt(c.FormValue("confusion")),
		Coma:                    ParseNullInt(c.FormValue("coma")),
		Bacteraemia:             ParseNullInt(c.FormValue("bacteraemia")),
		Hyperglycemia:           ParseNullInt(c.FormValue("hyperglycemia")),
		Hypoglycemia:            ParseNullInt(c.FormValue("hypoglycemia")),
		OtherComplications:      ParseNullInt(c.FormValue("other_complications")),
		AzaComplicationsSpecif:  ParseNullString(c.FormValue("aza_complications_specif")),
		PharyngealErythema:      ParseNullInt(c.FormValue("pharyngeal_erythema")),
		PharyngealExudate:       ParseNullInt(c.FormValue("pharyngeal_exudate")),
		ConjunctivalInjection:   ParseNullInt(c.FormValue("conjunctival_injection")),
		OedemaFace:              ParseNullInt(c.FormValue("oedema_face")),
		TenderAbdomen:           ParseNullInt(c.FormValue("tender_abdomen")),
		SunkenEyes:              ParseNullInt(c.FormValue("sunken_eyes")),
		TentingSkin:             ParseNullInt(c.FormValue("tenting_skin")),
		PalpableLiver:           ParseNullInt(c.FormValue("palpable_liver")),
		PalpableSpleen:          ParseNullInt(c.FormValue("palpable_spleen")),
		Jaundice:                ParseNullInt(c.FormValue("jaundice")),
		EnlargedLymphNodes:      ParseNullInt(c.FormValue("enlarged_lymph_nodes")),
		LowerExtremityOedema:    ParseNullInt(c.FormValue("lower_extremity_oedema")),
		Bleeding:                ParseNullInt(c.FormValue("bleeding")),
		BleedingNose:            ParseNullInt(c.FormValue("bleeding_nose")),
		BleedingMouth:           ParseNullInt(c.FormValue("bleeding_mouth")),
		BleedingVagina:          ParseNullInt(c.FormValue("bleeding_vagina")),
		BleedingRectum:          ParseNullInt(c.FormValue("bleeding_rectum")),
		BleedingSputum:          ParseNullInt(c.FormValue("bleeding_sputum")),
		BleedingUrine:           ParseNullInt(c.FormValue("bleeding_urine")),
		BleedingIvSite:          ParseNullInt(c.FormValue("bleeding_iv_site")),
		BleedingOther:           ParseNullInt(c.FormValue("bleeding_other")),
		BleedingOtherSpecif:     ParseNullString(c.FormValue("bleeding_other_specif")),
		FinalDiagnosis:          ParseNullInt(c.FormValue("final_diagnosis")),
		FinalDiagnosisAza:       ParseNullString(c.FormValue("final_diagnosis_aza")),
		OutcomeDischarge:        ParseNullInt(c.FormValue("outcome_discharge")),
		OutcomeDischargeIfHear:  ParseNullInt(c.FormValue("outcome_discharge_if_hear")),
		OutcomeDischargeIfArth:  ParseNullInt(c.FormValue("outcome_discharge_if_arth")),
		OutcomeDischargeIfAbor:  ParseNullInt(c.FormValue("outcome_discharge_if_abor")),
		OutcomeDischargeIfNeur:  ParseNullInt(c.FormValue("outcome_discharge_if_neur")),
		OutcomeDischargeIfOcul:  ParseNullInt(c.FormValue("outcome_discharge_if_ocul")),
		OutcomeDischargeIfExtr:  ParseNullInt(c.FormValue("outcome_discharge_if_extr")),
		OutcomeDischargeIfOthe:  ParseNullInt(c.FormValue("outcome_discharge_if_othe")),
		OutcomeDischargeIfAza:   ParseNullString(c.FormValue("outcome_discharge_if_aza")),
		OutcomeReferredFacility: ParseNullString(c.FormValue("outcome_referred_facility")),
		DischargeDate:           ParseNullString(c.FormValue("discharge_date")),
		SurvivorCounselling:     ParseNullInt(c.FormValue("survivor_counselling")),
	}

	if clinical_id == 0 {
		return clinical.Insert(c.Context(), db)
	} else {
		clinical.SetAsExists()
		return clinical.Update(c.Context(), db)
	}
}

func HandlerCaseEncounterSubmit(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config config.Config) error {
	// Get user ID from session
	sess, err := store.Get(c)
	if err != nil {
		sl.Error("Failed to get session", "error", err)
		return c.Status(500).SendString("Internal server error")
	}
	userID := sess.Get("user_id")
	if userID == nil {
		sl.Error("User not authenticated")
		return c.Status(401).SendString("Not authenticated")
	}

	// Get client ID and validate
	cid := c.FormValue("client_id")
	if cid == "" {
		sl.Error("Missing client ID")
		return c.Status(400).SendString("Missing client ID")
	}

	// Convert client ID to int for facility check
	clientID, err := strconv.Atoi(cid)
	if err != nil {
		sl.Error("Invalid client ID format", "error", err, "clientID", cid)
		return c.Status(400).SendString("Invalid client ID format")
	}

	// Check facility-based access control
	userFacility := GetCurrentFacility(c, db, sl, store)
	if userFacility > 0 {
		// Get client details to check facility
		client, err := models.ClientByID(c.Context(), db, clientID)
		if err != nil {
			sl.Error("Failed to get client for facility check", "error", err, "clientID", clientID)
			return c.Status(500).SendString("Failed to get client details")
		}

		// User has a facility assigned, check if they can access this case
		if client.Site.Int64 != int64(userFacility) {
			sl.Error("User attempted to submit encounter for case from different facility",
				"user_id", userID, "user_facility", userFacility, "case_site", client.Site.Int64, "case_id", clientID)
			return c.Status(403).SendString("Access denied: You can only access cases from your assigned facility.")
		}
	}

	// Get encounter date and validate format
	dte := c.FormValue("encounter_date")
	if dte == "" {
		sl.Error("Missing encounter date")
		return c.Status(400).SendString("Missing encounter date")
	}
	if _, err := time.Parse("2006-01-02", dte); err != nil {
		sl.Error("Invalid encounter date format", "error", err)
		return c.Status(400).SendString("Invalid encounter date format")
	}

	// Get outbreak ID from session (fallback to form/query if missing)
	outbreakID := sess.Get("outbreak_id")
	if outbreakID == nil {
		formOutbreakID := c.FormValue("outbreak_id")
		if formOutbreakID != "" {
			if oid, err := strconv.Atoi(formOutbreakID); err == nil {
				sess.Set("outbreak_id", oid)
				sess.Set("selected_outbreak", oid)
				if err := sess.Save(); err != nil {
					sl.Error("Failed to persist outbreak_id to session", "error", err)
				}
				outbreakID = oid
			}
		}
	}
	if outbreakID == nil {
		qOID := c.Query("outbreak_id")
		if qOID != "" {
			if oid, err := strconv.Atoi(qOID); err == nil {
				sess.Set("outbreak_id", oid)
				sess.Set("selected_outbreak", oid)
				if err := sess.Save(); err != nil {
					sl.Error("Failed to persist outbreak_id to session", "error", err)
				}
				outbreakID = oid
			}
		}
	}
	if outbreakID == nil {
		sl.Error("No outbreak selected")
		return c.Status(400).SendString("No outbreak selected")
	}

	// Save encounter and get IDs
	id1, id2, id3, err := saveEncounter(c, db, userID.(int), cid, dte, outbreakID.(int))
	if err != nil {
		sl.Error("Failed to save encounter", "error", err)
		return c.Status(500).SendString("Failed to save encounter")
	}

	// Save vitals
	saveVitals(c, db, id1, id2, id3)

	// Save clinical data
	saveClinical(c, db, id1, id2, id3)

	// Save lab data
	if err := saveLab(c, db, id1); err != nil {
		sl.Error("Failed to save lab data", "error", err)
		return c.Status(500).SendString("Failed to save lab data")
	}

	// Save treatment data
	if err := saveTreatment(c, db, id1); err != nil {
		sl.Error("Failed to save treatment data", "error", err)
		return c.Status(500).SendString("Failed to save treatment data")
	}

	return c.Redirect(fmt.Sprintf("/cases/encounters/new/%d?dte=%s&outbreak_id=%d", id2, dte, outbreakID.(int)))
}

// Helper function to save lab data
func saveLab(c *fiber.Ctx, db *sql.DB, encounterID int) error {
	lab_id, err := strconv.Atoi(c.FormValue("lab_id"))
	if err != nil {
		lab_id = 0
	}

	lab := models.Lab{
		LabID:                 lab_id,
		EncounterID:           sql.NullInt64{Int64: int64(encounterID), Valid: true},
		Specimen:              ParseNullInt(c.FormValue("specimen")),
		SampleBlood:           ParseNullInt(c.FormValue("sample_blood")),
		SampleUrine:           ParseNullInt(c.FormValue("sample_urine")),
		SampleSwab:            ParseNullInt(c.FormValue("sample_swab")),
		SampleAza:             ParseNullString(c.FormValue("sample_aza")),
		EbolaRdt:              ParseNullInt(c.FormValue("ebola_rdt")),
		EbolaRdtDate:          ParseNullString(c.FormValue("ebola_rdt_date")),
		EbolaRdtResults:       ParseNullInt(c.FormValue("ebola_rdt_results")),
		EbolaPcr:              ParseNullInt(c.FormValue("ebola_pcr")),
		EbolaPcrAza:           ParseNullString(c.FormValue("ebola_pcr_aza")),
		EbolaPcrDate:          ParseNullString(c.FormValue("ebola_pcr_date")),
		EbolaPcrGp:            ParseNullInt(c.FormValue("ebola_pcr_gp")),
		EbolaPcrGpCt:          ParseNullFloat(c.FormValue("ebola_pcr_gp_ct")),
		EbolaPcrNp:            ParseNullInt(c.FormValue("ebola_pcr_np")),
		EbolaPcrNpCt:          ParseNullFloat(c.FormValue("ebola_pcr_np_ct")),
		EbolaPcrIndeterminate: ParseNullInt(c.FormValue("ebola_pcr_indeterminate")),
		MalariaRdtDate:        ParseNullString(c.FormValue("malaria_rdt_date")),
		MalariaRdtResult:      ParseNullInt(c.FormValue("malaria_rdt_result")),
		BloodCultureDate:      ParseNullString(c.FormValue("blood_culture_date")),
		BloodCultureResult:    ParseNullInt(c.FormValue("blood_culture_result")),
		TestPosInfection:      ParseNullInt(c.FormValue("test_pos_infection")),
		TestPosInfectionAza:   ParseNullString(c.FormValue("test_pos_infection_aza")),
		Haemoglobinuria:       ParseNullInt(c.FormValue("haemoglobinuria")),
		Proteinuria:           ParseNullInt(c.FormValue("proteinuria")),
		Hematuria:             ParseNullInt(c.FormValue("hematuria")),
		BloodGas:              ParseNullInt(c.FormValue("blood_gas")),
		Ph:                    ParseNullFloat(c.FormValue("ph")),
		Pco2:                  ParseNullFloat(c.FormValue("pco2")),
		Pao2:                  ParseNullFloat(c.FormValue("pao2")),
		Hco3:                  ParseNullFloat(c.FormValue("hco3")),
		OxygenTherapy:         ParseNullFloat(c.FormValue("oxygen_therapy")),
		AltSgpt:               ParseNullFloat(c.FormValue("alt_sgpt")),
		AstSgo:                ParseNullFloat(c.FormValue("ast_sgo")),
		Creatinine:            ParseNullFloat(c.FormValue("creatinine")),
		Potassium:             ParseNullFloat(c.FormValue("potassium")),
		Urea:                  ParseNullFloat(c.FormValue("urea")),
		CreatinineKinase:      ParseNullFloat(c.FormValue("creatinine_kinase")),
		Calcium:               ParseNullFloat(c.FormValue("calcium")),
		Sodium:                ParseNullFloat(c.FormValue("sodium")),
		AltSgptNd:             ParseNullInt(c.FormValue("alt_sgpt_nd")),
		AstSgoNd:              ParseNullInt(c.FormValue("ast_sgo_nd")),
		CreatinineNd:          ParseNullInt(c.FormValue("creatinine_nd")),
		PotassiumNd:           ParseNullInt(c.FormValue("potassium_nd")),
		UreaNd:                ParseNullInt(c.FormValue("urea_nd")),
		CreatinineKinaseNd:    ParseNullInt(c.FormValue("creatinine_kinase_nd")),
		CalciumNd:             ParseNullInt(c.FormValue("calcium_nd")),
		SodiumNd:              ParseNullInt(c.FormValue("sodium_nd")),
		Glucose:               ParseNullFloat(c.FormValue("glucose")),
		Lactate:               ParseNullFloat(c.FormValue("lactate")),
		Haemoglobin:           ParseNullFloat(c.FormValue("haemoglobin")),
		TotalBilirubin:        ParseNullFloat(c.FormValue("total_bilirubin")),
		WbcCount:              ParseNullFloat(c.FormValue("wbc_count")),
		Platelets:             ParseNullFloat(c.FormValue("platelets")),
		Pt:                    ParseNullFloat(c.FormValue("pt")),
		Aptt:                  ParseNullFloat(c.FormValue("aptt")),
		GlucoseNd:             ParseNullInt(c.FormValue("glucose_nd")),
		LactateNd:             ParseNullInt(c.FormValue("lactate_nd")),
		HaemoglobinNd:         ParseNullInt(c.FormValue("haemoglobin_nd")),
		TotalBilirubinNd:      ParseNullInt(c.FormValue("total_bilirubin_nd")),
		WbcCountNd:            ParseNullInt(c.FormValue("wbc_count_nd")),
		PlateletsNd:           ParseNullInt(c.FormValue("platelets_nd")),
		PtNd:                  ParseNullInt(c.FormValue("pt_nd")),
		ApttNd:                ParseNullInt(c.FormValue("aptt_nd")),
		// New laboratory investigation fields
		TotalProtein: ParseNullFloat(c.FormValue("total_protein")),
		Albumin:      ParseNullFloat(c.FormValue("albumin")),
		BilirubinD:   ParseNullFloat(c.FormValue("bilirubin_d")),
		Lymphocytes:  ParseNullFloat(c.FormValue("lymphocytes")),
		Monocytes:    ParseNullFloat(c.FormValue("monocytes")),
		Eosinophils:  ParseNullFloat(c.FormValue("eosinophils")),
		Basophils:    ParseNullFloat(c.FormValue("basophils")),
		Neutrophils:  ParseNullFloat(c.FormValue("neutrophils")),
		Hgb:          ParseNullFloat(c.FormValue("hgb")),
		Hct:          ParseNullFloat(c.FormValue("hct")),
		Mcv:          ParseNullFloat(c.FormValue("mcv")),
		Mch:          ParseNullFloat(c.FormValue("mch")),
		Mchc:         ParseNullFloat(c.FormValue("mchc")),
		Rdw:          ParseNullFloat(c.FormValue("rdw")),
		RdwSd:        ParseNullFloat(c.FormValue("rdw_sd")),
		RdwCv:        ParseNullFloat(c.FormValue("rdw_cv")),
		Mpv:          ParseNullFloat(c.FormValue("mpv")),
		Pdw:          ParseNullString(c.FormValue("pdw")),
		Pct:          ParseNullFloat(c.FormValue("pct")),
		LabOther:     ParseNullString(c.FormValue("lab_other")),
		// "Not done" fields for new lab tests
		TotalProteinNd: sql.NullBool{Bool: c.FormValue("total_protein_nd") == "on", Valid: true},
		AlbuminNd:      sql.NullBool{Bool: c.FormValue("albumin_nd") == "on", Valid: true},
		BilirubinDNd:   sql.NullBool{Bool: c.FormValue("bilirubin_d_nd") == "on", Valid: true},
		LymphocytesNd:  sql.NullBool{Bool: c.FormValue("lymphocytes_nd") == "on", Valid: true},
		MonocytesNd:    sql.NullBool{Bool: c.FormValue("monocytes_nd") == "on", Valid: true},
		EosinophilsNd:  sql.NullBool{Bool: c.FormValue("eosinophils_nd") == "on", Valid: true},
		BasophilsNd:    sql.NullBool{Bool: c.FormValue("basophils_nd") == "on", Valid: true},
		NeutrophilsNd:  sql.NullBool{Bool: c.FormValue("neutrophils_nd") == "on", Valid: true},
		HgbNd:          sql.NullBool{Bool: c.FormValue("hgb_nd") == "on", Valid: true},
		HctNd:          sql.NullBool{Bool: c.FormValue("hct_nd") == "on", Valid: true},
		McvNd:          sql.NullBool{Bool: c.FormValue("mcv_nd") == "on", Valid: true},
		MchNd:          sql.NullBool{Bool: c.FormValue("mch_nd") == "on", Valid: true},
		MchcNd:         sql.NullBool{Bool: c.FormValue("mchc_nd") == "on", Valid: true},
		RdwNd:          sql.NullBool{Bool: c.FormValue("rdw_nd") == "on", Valid: true},
		RdwSdNd:        sql.NullBool{Bool: c.FormValue("rdw_sd_nd") == "on", Valid: true},
		RdwCvNd:        sql.NullBool{Bool: c.FormValue("rdw_cv_nd") == "on", Valid: true},
		MpvNd:          sql.NullBool{Bool: c.FormValue("mpv_nd") == "on", Valid: true},
		PdwNd:          sql.NullBool{Bool: c.FormValue("pdw_nd") == "on", Valid: true},
		PctNd:          sql.NullBool{Bool: c.FormValue("pct_nd") == "on", Valid: true},
		LabOtherNd:     sql.NullBool{Bool: c.FormValue("lab_other_nd") == "on", Valid: true},
		// Other test fields
		OtherMalaria:    ParseNullString(c.FormValue("other_malaria")),
		OtherHIV:        ParseNullString(c.FormValue("other_hiv")),
		OtherSyphilis:   ParseNullString(c.FormValue("other_syphilis")),
		OtherMpox:       ParseNullString(c.FormValue("other_mpox")),
		HepatitisB:      ParseNullString(c.FormValue("hepatitis_b")),
		HepatitisC:      ParseNullString(c.FormValue("hepatitis_c")),
		DataEntrantName: ParseNullString(c.FormValue("data_entrant_name")),
	}

	if lab_id == 0 {
		// Create minimal row first to avoid INSERT column-mismatch issues, then update all fields
		var newID int
		if err := db.QueryRowContext(c.Context(), "INSERT INTO public.lab (encounter_id) VALUES ($1) RETURNING lab_id", encounterID).Scan(&newID); err != nil {
			return err
		}
		lab.LabID = newID
		lab.SetAsExists()
		return lab.Update(c.Context(), db)
	} else {
		lab.SetAsExists()
		return lab.Update(c.Context(), db)
	}
}

// Helper function to save treatment data
func saveTreatment(c *fiber.Ctx, db *sql.DB, encounterID int) error {
	treat_id, err := strconv.Atoi(c.FormValue("treat_id"))
	if err != nil {
		treat_id = 0
	}

	// Use a simpler approach: create minimal row first, then update
	if treat_id == 0 {
		// Insert new treatment with just encounter_id and outbreak_id
		var newID int
		if err := db.QueryRowContext(c.Context(),
			"INSERT INTO public.treatment (encounter_id, outbreak_id) VALUES ($1, $2) RETURNING treatment_id",
			encounterID, 6).Scan(&newID); err != nil {
			return fmt.Errorf("failed to insert treatment: %v", err)
		}
		treat_id = newID
	}

	// Now update with all the form data
	query := `UPDATE public.treatment SET 
		antibacterial = $1, amoxicillin = $2, ceftriaxone = $3, cefixime = $4, 
		ampicillin = $5, chloramphenicol = $6, amoxiclav = $7, azithromycin = $8,
		cefotaxime = $9, ceftazidime = $10, ciprofloxacin = $11, tetracycline = $12,
		imipenem = $13, cotrimoxazole = $14, gentamicin = $15, metronidazole = $16,
		antibacterial_other = $17, antibacterial_dose = $18, antibacterial_route = $19, 
		antibacterial_freq = $20, antimalarial = $21, antimalarial_artesunate = $22,
		antimalarial_arthemeter = $23, antimalarial_al = $24, antimalarial_aa = $25,
		antimalarial_dose = $26, antimalarial_route = $27, antimalarial_freq = $28,
		other_meds_specify = $29, other_meds_dose = $30, other_meds_route = $31,
		other_meds_freq = $32, ebola_experimental = $33, ebola_experimental_if = $34,
		oral = $35, oral_ors = $36, oral_ors_qty = $37, oral_water = $38,
		oral_water_qty = $39, oral_other = $40, oral_other_qty = $41, iv = $42,
		iv_qty = $43, iv_using = $44, iv_aza = $45, access_type = $46,
		blood_trans = $47, oxygen_therapy = $48, oxygen_qty = $49, oxygen_with = $50,
		vasopressors = $51, renal = $52, invasive = $53, ebola_rdt_aza = $54,
		ebola_experimental_if_zmap = $55, ebola_experimental_if_remd = $56,
		ebola_experimental_if_regn = $57, ebola_experimental_if_favi = $58,
		ebola_experimental_if_mab = $59, oral_other_aza = $60, antibacterial_aza = $61,
		antimalarial_artesunate_dose = $62, antimalarial_artesunate_route = $63,
		antimalarial_artesunate_freq = $64, antimalarial_arthemeter_dose = $65,
		antimalarial_arthemeter_route = $66, antimalarial_arthemeter_freq = $67,
		antimalarial_al_dose = $68, antimalarial_al_route = $69, antimalarial_al_freq = $70,
		antimalarial_aa_dose = $71, antimalarial_aa_route = $72, antimalarial_aa_freq = $73,
		amoxicillin_dose = $74, amoxicillin_route = $75, amoxicillin_freq = $76,
		ceftriaxone_dose = $77, ceftriaxone_route = $78, ceftriaxone_freq = $79,
		cefixime_dose = $80, cefixime_route = $81, cefixime_freq = $82,
		ampicillin_dose = $83, ampicillin_route = $84, ampicillin_freq = $85,
		chloramphenicol_dose = $86, chloramphenicol_route = $87, chloramphenicol_freq = $88,
		amoxiclav_dose = $89, amoxiclav_route = $90, amoxiclav_freq = $91,
		azithromycin_dose = $92, azithromycin_route = $93, azithromycin_freq = $94,
		cefotaxime_dose = $95, cefotaxime_route = $96, cefotaxime_freq = $97,
		ceftazidime_dose = $98, ceftazidime_route = $99, ceftazidime_freq = $100,
		ciprofloxacin_dose = $101, ciprofloxacin_route = $102, ciprofloxacin_freq = $103,
		tetracycline_dose = $104, tetracycline_route = $105, tetracycline_freq = $106,
		imipenem_dose = $107, imipenem_route = $108, imipenem_freq = $109,
		cotrimoxazole_dose = $110, cotrimoxazole_route = $111, cotrimoxazole_freq = $112,
		gentamicin_dose = $113, gentamicin_route = $114, gentamicin_freq = $115,
		metronidazole_dose = $116, metronidazole_route = $117, metronidazole_freq = $118
		WHERE treatment_id = $119`

	args := []interface{}{
		ParseNullInt2(c.FormValue("antibacterial")),
		ParseNullInt2(c.FormValue("amoxicillin")),
		ParseNullInt2(c.FormValue("ceftriaxone")),
		ParseNullInt2(c.FormValue("cefixime")),
		ParseNullInt2(c.FormValue("ampicillin")),
		ParseNullInt2(c.FormValue("chloramphenicol")),
		ParseNullInt2(c.FormValue("amoxiclav")),
		ParseNullInt2(c.FormValue("azithromycin")),
		ParseNullInt2(c.FormValue("cefotaxime")),
		ParseNullInt2(c.FormValue("ceftazidime")),
		ParseNullInt2(c.FormValue("ciprofloxacin")),
		ParseNullInt2(c.FormValue("tetracycline")),
		ParseNullInt2(c.FormValue("imipenem")),
		ParseNullInt2(c.FormValue("cotrimoxazole")),
		ParseNullInt2(c.FormValue("gentamicin")),
		ParseNullInt2(c.FormValue("metronidazole")),
		ParseNullString2(c.FormValue("antibacterial_other")),
		ParseNullString2(c.FormValue("antibacterial_dose")),
		ParseNullString2(c.FormValue("antibacterial_route")),
		ParseNullString2(c.FormValue("antibacterial_freq")),
		ParseNullInt2(c.FormValue("antimalarial")),
		ParseNullInt2(c.FormValue("antimalarial_artesunate")),
		ParseNullInt2(c.FormValue("antimalarial_arthemeter")),
		ParseNullInt2(c.FormValue("antimalarial_al")),
		ParseNullInt2(c.FormValue("antimalarial_aa")),
		ParseNullString2(c.FormValue("antimalarial_dose")),
		ParseNullString2(c.FormValue("antimalarial_route")),
		ParseNullString2(c.FormValue("antimalarial_freq")),
		ParseNullString2(c.FormValue("other_meds_specify")),
		ParseNullString2(c.FormValue("other_meds_dose")),
		ParseNullString2(c.FormValue("other_meds_route")),
		ParseNullString2(c.FormValue("other_meds_freq")),
		ParseNullInt2(c.FormValue("ebola_experimental")),
		ParseNullString2(c.FormValue("ebola_experimental_if")),
		ParseNullInt2(c.FormValue("oral")),
		ParseNullInt2(c.FormValue("oral_ors")),
		ParseNullFloat(c.FormValue("oral_ors_qty")),
		ParseNullInt2(c.FormValue("oral_water")),
		ParseNullFloat(c.FormValue("oral_water_qty")),
		ParseNullInt2(c.FormValue("oral_other")),
		ParseNullFloat(c.FormValue("oral_other_qty")),
		ParseNullInt2(c.FormValue("iv")),
		ParseNullString2(c.FormValue("iv_qty")),
		ParseNullString2(c.FormValue("iv_using")),
		ParseNullString2(c.FormValue("iv_aza")),
		ParseNullInt2(c.FormValue("access_type")),
		ParseNullInt2(c.FormValue("blood_trans")),
		ParseNullInt2(c.FormValue("oxygen_therapy")),
		ParseNullFloat(c.FormValue("oxygen_qty")),
		ParseNullString2(c.FormValue("oxygen_with")),
		ParseNullInt2(c.FormValue("vasopressors")),
		ParseNullInt2(c.FormValue("renal")),
		ParseNullInt2(c.FormValue("invasive")),
		ParseNullInt2(c.FormValue("ebola_rdt_aza")),
		ParseNullInt2(c.FormValue("ebola_experimental_if_zmap")),
		ParseNullInt2(c.FormValue("ebola_experimental_if_remd")),
		ParseNullInt2(c.FormValue("ebola_experimental_if_regn")),
		ParseNullInt2(c.FormValue("ebola_experimental_if_favi")),
		ParseNullInt2(c.FormValue("ebola_experimental_if_mab")),
		ParseNullString2(c.FormValue("oral_other_aza")),
		ParseNullInt2(c.FormValue("antibacterial_aza")),
		ParseNullString2(c.FormValue("antimalarial_artesunate_dose")),
		ParseNullString2(c.FormValue("antimalarial_artesunate_route")),
		ParseNullString2(c.FormValue("antimalarial_artesunate_freq")),
		ParseNullString2(c.FormValue("antimalarial_arthemeter_dose")),
		ParseNullString2(c.FormValue("antimalarial_arthemeter_route")),
		ParseNullString2(c.FormValue("antimalarial_arthemeter_freq")),
		ParseNullString2(c.FormValue("antimalarial_al_dose")),
		ParseNullString2(c.FormValue("antimalarial_al_route")),
		ParseNullString2(c.FormValue("antimalarial_al_freq")),
		ParseNullString2(c.FormValue("antimalarial_aa_dose")),
		ParseNullString2(c.FormValue("antimalarial_aa_route")),
		ParseNullString2(c.FormValue("antimalarial_aa_freq")),
		ParseNullString2(c.FormValue("amoxicillin_dose")),
		ParseNullString2(c.FormValue("amoxicillin_route")),
		ParseNullString2(c.FormValue("amoxicillin_freq")),
		ParseNullString2(c.FormValue("ceftriaxone_dose")),
		ParseNullString2(c.FormValue("ceftriaxone_route")),
		ParseNullString2(c.FormValue("ceftriaxone_freq")),
		ParseNullString2(c.FormValue("cefixime_dose")),
		ParseNullString2(c.FormValue("cefixime_route")),
		ParseNullString2(c.FormValue("cefixime_freq")),
		ParseNullString2(c.FormValue("ampicillin_dose")),
		ParseNullString2(c.FormValue("ampicillin_route")),
		ParseNullString2(c.FormValue("ampicillin_freq")),
		ParseNullString2(c.FormValue("chloramphenicol_dose")),
		ParseNullString2(c.FormValue("chloramphenicol_route")),
		ParseNullString2(c.FormValue("chloramphenicol_freq")),
		ParseNullString2(c.FormValue("amoxiclav_dose")),
		ParseNullString2(c.FormValue("amoxiclav_route")),
		ParseNullString2(c.FormValue("amoxiclav_freq")),
		ParseNullString2(c.FormValue("azithromycin_dose")),
		ParseNullString2(c.FormValue("azithromycin_route")),
		ParseNullString2(c.FormValue("azithromycin_freq")),
		ParseNullString2(c.FormValue("cefotaxime_dose")),
		ParseNullString2(c.FormValue("cefotaxime_route")),
		ParseNullString2(c.FormValue("cefotaxime_freq")),
		ParseNullString2(c.FormValue("ceftazidime_dose")),
		ParseNullString2(c.FormValue("ceftazidime_route")),
		ParseNullString2(c.FormValue("ceftazidime_freq")),
		ParseNullString2(c.FormValue("ciprofloxacin_dose")),
		ParseNullString2(c.FormValue("ciprofloxacin_route")),
		ParseNullString2(c.FormValue("ciprofloxacin_freq")),
		ParseNullString2(c.FormValue("tetracycline_dose")),
		ParseNullString2(c.FormValue("tetracycline_route")),
		ParseNullString2(c.FormValue("tetracycline_freq")),
		ParseNullString2(c.FormValue("imipenem_dose")),
		ParseNullString2(c.FormValue("imipenem_route")),
		ParseNullString2(c.FormValue("imipenem_freq")),
		ParseNullString2(c.FormValue("cotrimoxazole_dose")),
		ParseNullString2(c.FormValue("cotrimoxazole_route")),
		ParseNullString2(c.FormValue("cotrimoxazole_freq")),
		ParseNullString2(c.FormValue("gentamicin_dose")),
		ParseNullString2(c.FormValue("gentamicin_route")),
		ParseNullString2(c.FormValue("gentamicin_freq")),
		ParseNullString2(c.FormValue("metronidazole_dose")),
		ParseNullString2(c.FormValue("metronidazole_route")),
		ParseNullString2(c.FormValue("metronidazole_freq")),
		treat_id,
	}

	if _, err := db.ExecContext(c.Context(), query, args...); err != nil {
		return fmt.Errorf("failed to update treatment: %v", err)
	}
	return nil
}

func HandlerAPIGetEncounter(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config config.Config) error {
	// Get ID from the query parameter

	id := c.Query("id")

	if id == "" {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message": "",
		})
	}

	encounter_id, err := strconv.Atoi(id)
	if err != nil {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message": "",
		})
	}

	var clinical = &models.Clinical{}
	var vital = &models.Vital{}

	clinical, _ = models.ClinicalByEncounterID(c.Context(), db, encounter_id)
	vital, _ = models.VitalByEncounterID(c.Context(), db, encounter_id)

	rtnStr := ` Vitals<br />
				<table class="full-width" border="1">
					<tr>
						<td>Weight: ` + fmt.Sprintf("%.2f", vital.Weight.Float64) + `</td>
						<td>Height: ` + fmt.Sprintf("%.2f", vital.Height.Float64) + `</td>
					</tr>
				</table>
				Symptomms<br/>
				<table class="full-width" border="1">
					<tr>
						<td valign="top">
							Fever: ` + strconv.Itoa(int(clinical.Fever.Int64)) + `<br/>
							Fatigue:<br/>
							Weakness:<br/>
							Malaise:<br/>
							Myalgia:<br/>
							Anorexia:<br/>
							Sore throat
						</td>
						<td valign="top">
							Headache:<br/> 
							Nausea:<br/> 
							Chest pain:<br/> 
							Joint Pain:<br/> 
							Hiccups:<br/>
							Cough:<br/>
						</td>
						<td valign="top">
							Chest pain:<br/>
							Difficulty breathing:<br/>
							Difficulty swallowing:<br/> 
							Abdominal pain:<br/> 
							Diarrhoea:<br/>
							Vomiting:<br/>
							Irritability / Confusion:<br/> 
						</td>
					</tr>
				</table>

				<br/>
				Signs<br/>
				<table class="full-width" border="1">
					<tr>
						<td valign="top">
							Pharyngeal erythema:<br/>  
							Pharyngeal exudate:<br/>  
							Conjunctival injection/bleeding:<br/>  
							Oedema of face/neck:<br/> 
							Tender abdomen:<br/> 
							Sunken eyes or fontanelle:<br/>  
							Tenting on skin pinch:<br/>  
							Palpable liver:<br/> 
							Palpable spleen Rash:<br/> 
							Jaundice:<br/> 

						</td>
						<td valign="top">
							Enlarged lymph nodes:<br/>
							Lower extremity oedema :<br/> 
							Bleeding:<br/> 
						</td>
					</tr>
				</table>
				<br/>
				Specimen <br/>
				<table class="full-width" border="1">
					<tr>
						<td valign="top">
						</td>
					</tr>
				</table>
				<br/>
				Lab Results <br/>
				<table class="full-width" border="1">
					<tr>
						<td valign="top">
						</td>
					</tr>
				</table>
				<br/>
				Medications <br/>
				<table class="full-width" border="1">
					<tr>
						<td valign="top">
						</td>
					</tr>
				</table>`

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": rtnStr,
	})

}

func HandlerAPIGetStatuses(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config config.Config) error {
	userID := GetCurrentUser(c, store)

	// Check if user is logged in
	if userID == 0 {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	ClientID := c.Query("client_id")
	if ClientID == "" {
		ClientID = "0"
	}

	statuses, er := models.Statusez(c.Context(), db, " client_id = "+ClientID)
	if er != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Error fetching statuses",
		})
	}

	return c.JSON(statuses)

}

func HandlerAPIPostStatus(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config config.Config) error {

	//=================

	userID := GetCurrentUser(c, store)
	// Check if user is logged in
	if userID == 0 {
		fmt.Println("unauthorized")
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	//=============================

	var formData map[string]interface{}

	if err := c.BodyParser(&formData); err != nil {
		fmt.Println("JSON parsing failed:", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	var s models.Status

	s.StatusID = int(ParseNullInt2(formData["status_id"]).Int64)
	s.ClientID = ParseNullInt2(formData["client_id"])
	s.StatusDate = ParseNullString2(formData["status_date"])
	s.Status = ParseNullString2(formData["status"])

	// Build status notes - include transfer info if status is Transfer
	statusNotes := ParseNullString2(formData["status_notes"])
	if s.Status.String == "Transfer" {
		transferSite := ""
		transferStatus := ""
		treatmentProgram := ""

		if val, ok := formData["transfer_site"].(string); ok && val != "" {
			transferSite = val
		}
		if val, ok := formData["transfer_status"].(string); ok && val != "" {
			transferStatus = val
		}
		if val, ok := formData["treatment_program"].(string); ok && val != "" {
			treatmentProgram = val
		}

		transferInfo := ""
		if transferSite != "" || transferStatus != "" || treatmentProgram != "" {
			transferInfo = " [Transfer Details: "
			if transferSite != "" {
				transferInfo += "Site: " + transferSite + "; "
			}
			if transferStatus != "" {
				transferInfo += "Status: " + transferStatus + "; "
			}
			if treatmentProgram != "" {
				transferInfo += "Treatment Program: " + treatmentProgram + "; "
			}
			transferInfo = strings.TrimSuffix(transferInfo, "; ") + "]"
		}

		if statusNotes.String != "" {
			s.StatusNotes = sql.NullString{String: statusNotes.String + transferInfo, Valid: true}
		} else {
			s.StatusNotes = sql.NullString{String: strings.TrimPrefix(transferInfo, " "), Valid: true}
		}
	} else {
		s.StatusNotes = statusNotes
	}

	s.UpdatedBy.Valid = true
	s.UpdatedBy.Int64 = int64(userID)

	s.UpdatedOn.Valid = true
	currentTime := time.Now()
	formattedTime := currentTime.Format("2006-01-02")
	s.UpdatedOn.String = formattedTime

	// Check if it's a new record (StatusID == 0)
	if s.StatusID > 0 {
		s.SetAsExists()
		err := s.Update(c.Context(), db)
		if err != nil {
			fmt.Println("update fail:", err.Error())
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
	} else {

		err := s.Insert(c.Context(), db)
		if err != nil {
			fmt.Println("insert fail:", err.Error())
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "success",
	})
}

// Helper function to load full treatment data from database
func loadFullTreatmentData(ctx context.Context, db *sql.DB, encounterID int) (*FullTreatmentData, error) {
	query := `SELECT 
		treatment_id, encounter_id, outbreak_id, antibacterial, amoxicillin, ceftriaxone, cefixime, ampicillin, chloramphenicol, amoxiclav, azithromycin, cefotaxime, ceftazidime, ciprofloxacin, tetracycline, imipenem, cotrimoxazole, gentamicin, metronidazole, antibacterial_other, antibacterial_dose, antibacterial_route, antibacterial_freq, antimalarial, antimalarial_artesunate, antimalarial_arthemeter, antimalarial_al, antimalarial_aa, antimalarial_dose, antimalarial_route, antimalarial_freq, other_meds_specify, other_meds_dose, other_meds_route, other_meds_freq, ebola_experimental, ebola_experimental_if, oral, oral_ors, oral_ors_qty, oral_water, oral_water_qty, oral_other, oral_other_qty, iv, iv_qty, iv_using, iv_aza, access_type, blood_trans, oxygen_therapy, oxygen_qty, oxygen_with, vasopressors, renal, invasive, ebola_rdt_aza, ebola_experimental_if_zmap, ebola_experimental_if_remd, ebola_experimental_if_regn, ebola_experimental_if_favi, ebola_experimental_if_mab, oral_other_aza, antibacterial_aza, antimalarial_artesunate_dose, antimalarial_artesunate_route, antimalarial_artesunate_freq, antimalarial_arthemeter_dose, antimalarial_arthemeter_route, antimalarial_arthemeter_freq, antimalarial_al_dose, antimalarial_al_route, antimalarial_al_freq, antimalarial_aa_dose, antimalarial_aa_route, antimalarial_aa_freq, amoxicillin_dose, amoxicillin_route, amoxicillin_freq, ceftriaxone_dose, ceftriaxone_route, ceftriaxone_freq, cefixime_dose, cefixime_route, cefixime_freq, ampicillin_dose, ampicillin_route, ampicillin_freq, chloramphenicol_dose, chloramphenicol_route, chloramphenicol_freq, amoxiclav_dose, amoxiclav_route, amoxiclav_freq, azithromycin_dose, azithromycin_route, azithromycin_freq, cefotaxime_dose, cefotaxime_route, cefotaxime_freq, ceftazidime_dose, ceftazidime_route, ceftazidime_freq, ciprofloxacin_dose, ciprofloxacin_route, ciprofloxacin_freq, tetracycline_dose, tetracycline_route, tetracycline_freq, imipenem_dose, imipenem_route, imipenem_freq, cotrimoxazole_dose, cotrimoxazole_route, cotrimoxazole_freq, gentamicin_dose, gentamicin_route, gentamicin_freq, metronidazole_dose, metronidazole_route, metronidazole_freq
		FROM public.treatment 
		WHERE encounter_id = $1`

	var treatment FullTreatmentData

	err := db.QueryRowContext(ctx, query, encounterID).Scan(
		&treatment.TreatmentID, &treatment.EncounterID, &treatment.OutbreakID, &treatment.Antibacterial, &treatment.Amoxicillin, &treatment.Ceftriaxone, &treatment.Cefixime, &treatment.Ampicillin, &treatment.Chloramphenicol, &treatment.Amoxiclav, &treatment.Azithromycin, &treatment.Cefotaxime, &treatment.Ceftazidime, &treatment.Ciprofloxacin, &treatment.Tetracycline, &treatment.Imipenem, &treatment.Cotrimoxazole, &treatment.Gentamicin, &treatment.Metronidazole, &treatment.AntibacterialOther, &treatment.AntibacterialDose, &treatment.AntibacterialRoute, &treatment.AntibacterialFreq, &treatment.Antimalarial, &treatment.AntimalarialArtesunate, &treatment.AntimalarialArthemeter, &treatment.AntimalarialAl, &treatment.AntimalarialAa, &treatment.AntimalarialDose, &treatment.AntimalarialRoute, &treatment.AntimalarialFreq, &treatment.OtherMedsSpecify, &treatment.OtherMedsDose, &treatment.OtherMedsRoute, &treatment.OtherMedsFreq, &treatment.EbolaExperimental, &treatment.EbolaExperimentalIf, &treatment.Oral, &treatment.OralOrs, &treatment.OralOrsQty, &treatment.OralWater, &treatment.OralWaterQty, &treatment.OralOther, &treatment.OralOtherQty, &treatment.Iv, &treatment.IvQty, &treatment.IvUsing, &treatment.IvAza, &treatment.AccessType, &treatment.BloodTrans, &treatment.OxygenTherapy, &treatment.OxygenQty, &treatment.OxygenWith, &treatment.Vasopressors, &treatment.Renal, &treatment.Invasive, &treatment.EbolaRdtAza, &treatment.EbolaExperimentalIfZmap, &treatment.EbolaExperimentalIfRemd, &treatment.EbolaExperimentalIfRegn, &treatment.EbolaExperimentalIfFavi, &treatment.EbolaExperimentalIfMab, &treatment.OralOtherAza, &treatment.AntibacterialAza, &treatment.AntimalarialArtesunateDose, &treatment.AntimalarialArtesunateRoute, &treatment.AntimalarialArtesunateFreq, &treatment.AntimalarialArthemeterDose, &treatment.AntimalarialArthemeterRoute, &treatment.AntimalarialArthemeterFreq, &treatment.AntimalarialAlDose, &treatment.AntimalarialAlRoute, &treatment.AntimalarialAlFreq, &treatment.AntimalarialAaDose, &treatment.AntimalarialAaRoute, &treatment.AntimalarialAaFreq, &treatment.AmoxicillinDose, &treatment.AmoxicillinRoute, &treatment.AmoxicillinFreq, &treatment.CeftriaxoneDose, &treatment.CeftriaxoneRoute, &treatment.CeftriaxoneFreq, &treatment.CefiximeDose, &treatment.CefiximeRoute, &treatment.CefiximeFreq, &treatment.AmpicillinDose, &treatment.AmpicillinRoute, &treatment.AmpicillinFreq, &treatment.ChloramphenicolDose, &treatment.ChloramphenicolRoute, &treatment.ChloramphenicolFreq, &treatment.AmoxiclavDose, &treatment.AmoxiclavRoute, &treatment.AmoxiclavFreq, &treatment.AzithromycinDose, &treatment.AzithromycinRoute, &treatment.AzithromycinFreq, &treatment.CefotaximeDose, &treatment.CefotaximeRoute, &treatment.CefotaximeFreq, &treatment.CeftazidimeDose, &treatment.CeftazidimeRoute, &treatment.CeftazidimeFreq, &treatment.CiprofloxacinDose, &treatment.CiprofloxacinRoute, &treatment.CiprofloxacinFreq, &treatment.TetracyclineDose, &treatment.TetracyclineRoute, &treatment.TetracyclineFreq, &treatment.ImipenemDose, &treatment.ImipenemRoute, &treatment.ImipenemFreq, &treatment.CotrimoxazoleDose, &treatment.CotrimoxazoleRoute, &treatment.CotrimoxazoleFreq, &treatment.GentamicinDose, &treatment.GentamicinRoute, &treatment.GentamicinFreq, &treatment.MetronidazoleDose, &treatment.MetronidazoleRoute, &treatment.MetronidazoleFreq,
	)

	if err != nil {
		return nil, err
	}

	return &treatment, nil
}
