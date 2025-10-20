package handlers

import (
	"context"
	"database/sql"
	"log/slog"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"

	"case/internal/models"
)

// User Management APIs
func HandlerUserListAPI(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	userID, userRole := GetUser(c, sl, store)
	if userRole == "" || userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	users := []fiber.Map{
		{"id": 1, "username": "admin", "email": "admin@example.com"},
		{"id": 2, "username": "user1", "email": "user1@example.com"},
	}

	return c.JSON(fiber.Map{"users": users})
}

func HandlerGetUserAPI(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	userID, userRole := GetUser(c, sl, store)
	if userRole == "" || userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	return c.JSON(fiber.Map{"user": fiber.Map{"id": c.Params("id"), "username": "username"}})
}

func HandlerUserSubmitAPI(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	userID, userRole := GetUser(c, sl, store)
	if userRole == "" || userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	var data fiber.Map
	c.BodyParser(&data)
	return c.Status(201).JSON(fiber.Map{"message": "User created successfully"})
}

func HandlerUserUpdateAPI(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	userID, userRole := GetUser(c, sl, store)
	if userRole == "" || userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	_ = c.Params("id")
	var data fiber.Map
	c.BodyParser(&data)
	return c.JSON(fiber.Map{"message": "User updated successfully"})
}

func HandlerUserDeleteAPI(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	userID, userRole := GetUser(c, sl, store)
	if userRole == "" || userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	_ = c.Params("id")
	return c.JSON(fiber.Map{"message": "User deleted successfully"})
}

// Facility Management APIs
func HandlerFacilityListAPI(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	userID, userRole := GetUser(c, sl, store)
	if userRole == "" || userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	facilities := []fiber.Map{
		{"id": 1, "name": "Hospital A", "district": "Kampala"},
		{"id": 2, "name": "Hospital B", "district": "Wakiso"},
	}

	return c.JSON(fiber.Map{"facilities": facilities})
}

func HandlerGetFacilityAPI(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	userID, userRole := GetUser(c, sl, store)
	if userRole == "" || userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	return c.JSON(fiber.Map{"facility": fiber.Map{"id": c.Params("id"), "name": "Facility Name"}})
}

func HandlerFacilitySubmitAPI(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	userID, userRole := GetUser(c, sl, store)
	if userRole == "" || userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	var data fiber.Map
	c.BodyParser(&data)
	return c.Status(201).JSON(fiber.Map{"message": "Facility created successfully"})
}

func HandlerFacilityUpdateAPI(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	userID, userRole := GetUser(c, sl, store)
	if userRole == "" || userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	_ = c.Params("id")
	var data fiber.Map
	c.BodyParser(&data)
	return c.JSON(fiber.Map{"message": "Facility updated successfully"})
}

func HandlerFacilityDeleteAPI(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	userID, userRole := GetUser(c, sl, store)
	if userRole == "" || userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	_ = c.Params("id")
	return c.JSON(fiber.Map{"message": "Facility deleted successfully"})
}

// Outbreak Management APIs
func HandlerOutbreakListAPI(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	userID, userRole := GetUser(c, sl, store)
	if userRole == "" || userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	outbreaks := []fiber.Map{
		{"id": 1, "name": "Ebola Outbreak", "status": "active"},
		{"id": 2, "name": "Measles Outbreak", "status": "closed"},
	}

	return c.JSON(fiber.Map{"outbreaks": outbreaks})
}

func HandlerGetOutbreakAPI(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	userID, userRole := GetUser(c, sl, store)
	if userRole == "" || userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	return c.JSON(fiber.Map{"outbreak": fiber.Map{"id": c.Params("id"), "name": "Outbreak Name"}})
}

func HandlerOutbreakSubmitAPI(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	userID, userRole := GetUser(c, sl, store)
	if userRole == "" || userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	var data fiber.Map
	c.BodyParser(&data)
	return c.Status(201).JSON(fiber.Map{"message": "Outbreak created successfully"})
}

func HandlerOutbreakUpdateAPI(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	userID, userRole := GetUser(c, sl, store)
	if userRole == "" || userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	_ = c.Params("id")
	var data fiber.Map
	c.BodyParser(&data)
	return c.JSON(fiber.Map{"message": "Outbreak updated successfully"})
}

func HandlerOutbreakDeleteAPI(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	userID, userRole := GetUser(c, sl, store)
	if userRole == "" || userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	_ = c.Params("id")
	return c.JSON(fiber.Map{"message": "Outbreak deleted successfully"})
}

func HandlerOutbreakCloseAPI(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	userID, userRole := GetUser(c, sl, store)
	if userRole == "" || userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	_ = c.Params("id")
	return c.JSON(fiber.Map{"message": "Outbreak closed successfully"})
}

func HandlerOutbreakSelectAPI(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	userID, userRole := GetUser(c, sl, store)
	if userRole == "" || userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	_ = c.Params("id")
	return c.JSON(fiber.Map{"message": "Outbreak selected successfully"})
}

// Case Management APIs
func HandlerCasesListAPI(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	userID, userRole := GetUser(c, sl, store)
	if userRole == "" || userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	cases := []fiber.Map{
		{"id": 1, "patient_name": "John Doe", "outbreak": "Ebola"},
		{"id": 2, "patient_name": "Jane Smith", "outbreak": "Measles"},
	}

	return c.JSON(fiber.Map{"cases": cases})
}

func HandlerGetCaseAPI(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	userID, userRole := GetUser(c, sl, store)
	if userRole == "" || userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	return c.JSON(fiber.Map{"case": fiber.Map{"id": c.Params("id"), "patient_name": "Patient Name"}})
}

func HandlerCasesSubmitAPI(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	userID, userRole := GetUser(c, sl, store)
	if userRole == "" || userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	var data struct {
		OutbreakID int64       `json:"outbreak_id"`
		CaseID     interface{} `json:"case_id"`
		Patient    interface{} `json:"patient"`
	}
	if err := c.BodyParser(&data); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Create a minimal admission row linked to a new encounter (if required by schema)
	// For now, insert into admission with current time and current user
	adm := &models.Admission{}
	adm.Admitted.Valid = true
	adm.Admitted.Int64 = 1
	adm.AdmissionDate.Valid = true
	adm.AdmissionDate.Time = time.Now()
	adm.EnterBy.Valid = true
	adm.EnterBy.Int64 = int64(userID)
	adm.EnterOn.Valid = true
	adm.EnterOn.Time = time.Now()

	if err := adm.Insert(context.Background(), db); err != nil {
		sl.Error("Failed to create admission", "error", err)
		return c.Status(500).JSON(fiber.Map{"error": "Failed to create admission"})
	}

	// Persist outbreak_id in session for subsequent pages (e.g., /cases/new/:id)
	if data.OutbreakID > 0 {
		if sess, err := store.Get(c); err == nil {
			sess.Set("outbreak_id", int(data.OutbreakID))
			sess.Set("selected_outbreak", int(data.OutbreakID))
			if err := sess.Save(); err != nil {
				sl.Error("Failed to save outbreak_id to session", "error", err)
			}
		} else {
			sl.Error("Failed to get session to persist outbreak_id", "error", err)
		}
	}

	// Return redirect to patient registration form (front-end will navigate)
	return c.Status(201).JSON(fiber.Map{
		"message":      "Admission created",
		"admission_id": adm.ID,
		"redirect":     "/cases/new/0",
	})
}

func HandlerCaseUpdateAPI(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	userID, userRole := GetUser(c, sl, store)
	if userRole == "" || userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	_ = c.Params("id")
	var data fiber.Map
	c.BodyParser(&data)
	return c.JSON(fiber.Map{"message": "Case updated successfully"})
}

func HandlerCaseDeleteAPI(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	userID, userRole := GetUser(c, sl, store)
	if userRole == "" || userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	_ = c.Params("id")
	return c.JSON(fiber.Map{"message": "Case deleted successfully"})
}

// Case Encounter APIs
func HandlerCaseEncounterListAPI(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	userID, userRole := GetUser(c, sl, store)
	if userRole == "" || userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	caseID := c.Params("id")
	encounters := []fiber.Map{
		{"id": 1, "case_id": caseID, "encounter_type": "admission"},
		{"id": 2, "case_id": caseID, "encounter_type": "follow_up"},
	}

	return c.JSON(fiber.Map{"encounters": encounters})
}

func HandlerGetCaseEncounterAPI(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	userID, userRole := GetUser(c, sl, store)
	if userRole == "" || userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	return c.JSON(fiber.Map{"encounter": fiber.Map{"id": c.Params("encounter_id"), "case_id": c.Params("id")}})
}

func HandlerCaseEncounterSubmitAPI(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	userID, userRole := GetUser(c, sl, store)
	if userRole == "" || userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	var data fiber.Map
	c.BodyParser(&data)
	return c.Status(201).JSON(fiber.Map{"message": "Encounter created successfully"})
}

func HandlerCaseEncounterUpdateAPI(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	userID, userRole := GetUser(c, sl, store)
	if userRole == "" || userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	_ = c.Params("id")
	_ = c.Params("encounter_id")
	var data fiber.Map
	c.BodyParser(&data)
	return c.JSON(fiber.Map{"message": "Encounter updated successfully"})
}

func HandlerCaseEncounterDeleteAPI(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	userID, userRole := GetUser(c, sl, store)
	if userRole == "" || userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	_ = c.Params("id")
	_ = c.Params("encounter_id")
	return c.JSON(fiber.Map{"message": "Encounter deleted successfully"})
}

// HandlerGetCIFAPI returns CIF data for a VHF case by ID
func HandlerGetCIFAPI(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	userID, userRole := GetUser(c, sl, store)
	if userRole == "" || userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	idStr := c.Params("id")
	if idStr == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Case ID required"})
	}

	// Convert to int64
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid case ID"})
	}

	// Fetch basic patient data
	patient, err := models.GetVHFPatient(db, id)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to load CIF"})
	}
	if patient == nil {
		return c.Status(404).JSON(fiber.Map{"error": "CIF not found"})
	}

	// Optionally fetch additional CIF sections
	signs, _ := models.GetVHFClinicalSigns(db, id)
	hosp, _ := models.GetVHFHospitalization(db, id)
	risk, _ := models.GetVHFRiskFactors(db, id)
	lab, _ := models.GetVHFLaboratory(db, id)
	inv, _ := models.GetVHFInvestigator(db, id)

	return c.JSON(fiber.Map{
		"patient":         patient,
		"clinical_signs":  signs,
		"hospitalization": hosp,
		"risk_factors":    risk,
		"laboratory":      lab,
		"investigator":    inv,
	})
}

// HandlerGetCIFByCaseCodeAPI returns CIF data using case_code query param
func HandlerGetCIFByCaseCodeAPI(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	userID, userRole := GetUser(c, sl, store)
	if userRole == "" || userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	caseCode := c.Query("case_code")
	if caseCode == "" {
		return c.Status(400).JSON(fiber.Map{"error": "case_code is required"})
	}

	patient, err := models.GetVHFPatientByCaseCode(db, caseCode)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to load CIF"})
	}
	if patient == nil {
		return c.Status(404).JSON(fiber.Map{"error": "CIF not found"})
	}

	id := patient.ID
	signs, _ := models.GetVHFClinicalSigns(db, id)
	hosp, _ := models.GetVHFHospitalization(db, id)
	risk, _ := models.GetVHFRiskFactors(db, id)
	lab, _ := models.GetVHFLaboratory(db, id)
	inv, _ := models.GetVHFInvestigator(db, id)

	return c.JSON(fiber.Map{
		"patient":         patient,
		"clinical_signs":  signs,
		"hospitalization": hosp,
		"risk_factors":    risk,
		"laboratory":      lab,
		"investigator":    inv,
	})
}

// Disease-specific CIF endpoints
// VHF
func HandlerVhfCIFByID(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	userID, userRole := GetUser(c, sl, store)
	if userRole == "" || userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}
	idStr := c.Params("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid case ID"})
	}
	patient, err := models.GetVHFPatient(db, id)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to load CIF"})
	}
	if patient == nil {
		return c.Status(404).JSON(fiber.Map{"error": "CIF not found"})
	}
	signs, _ := models.GetVHFClinicalSigns(db, id)
	hosp, _ := models.GetVHFHospitalization(db, id)
	risk, _ := models.GetVHFRiskFactors(db, id)
	lab, _ := models.GetVHFLaboratory(db, id)
	inv, _ := models.GetVHFInvestigator(db, id)
	return c.JSON(fiber.Map{
		"patient": patient, "clinical_signs": signs, "hospitalization": hosp,
		"risk_factors": risk, "laboratory": lab, "investigator": inv,
	})
}

func HandlerVhfCIFByCaseCode(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	userID, userRole := GetUser(c, sl, store)
	if userRole == "" || userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}
	caseCode := c.Query("case_code")
	if caseCode == "" {
		return c.Status(400).JSON(fiber.Map{"error": "case_code is required"})
	}
	patient, err := models.GetVHFPatientByCaseCode(db, caseCode)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to load CIF"})
	}
	if patient == nil {
		return c.Status(404).JSON(fiber.Map{"error": "CIF not found"})
	}
	id := patient.ID
	signs, _ := models.GetVHFClinicalSigns(db, id)
	hosp, _ := models.GetVHFHospitalization(db, id)
	risk, _ := models.GetVHFRiskFactors(db, id)
	lab, _ := models.GetVHFLaboratory(db, id)
	inv, _ := models.GetVHFInvestigator(db, id)
	return c.JSON(fiber.Map{
		"patient": patient, "clinical_signs": signs, "hospitalization": hosp,
		"risk_factors": risk, "laboratory": lab, "investigator": inv,
	})
}

// Measles (placeholder)
func HandlerMeaslesCIFByID(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	userID, userRole := GetUser(c, sl, store)
	if userRole == "" || userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}
	pid := c.Params("id")
	if pid == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Patient ID required"})
	}
	var patient models.MeaslesPatient
	err := db.QueryRow(`SELECT patient_id, measles_code, patient_name, sex, dob, created_at FROM measles_patients WHERE patient_id = $1`, pid).
		Scan(&patient.PatientID, &patient.MeaslesCode, &patient.PatientName, &patient.Sex, &patient.DOB, &patient.CreatedAt)
	if err == sql.ErrNoRows {
		return c.Status(404).JSON(fiber.Map{"error": "CIF not found"})
	}
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to load patient"})
	}
	var demo *models.MeaslesDemographics
	var inv *models.MeaslesInvestigators
	var hist *models.MeaslesClinicalHistory
	var res *models.MeaslesResults
	var spec *models.MeaslesSpecimens
	demo = &models.MeaslesDemographics{}
	if err := db.QueryRow(`SELECT id, patient_id, onset_district, reporting_unit, age_months, head_of_household, guardian_occupation, home_district, subcounty, parish, lc1_zone, lc1_chairman, lc1_tel FROM measles_demographics WHERE patient_id = $1`, patient.PatientID).
		Scan(&demo.ID, &demo.PatientID, &demo.OnsetDistrict, &demo.ReportingUnit, &demo.AgeMonths, &demo.HeadOfHousehold, &demo.GuardianOccupation, &demo.HomeDistrict, &demo.Subcounty, &demo.Parish, &demo.LC1Zone, &demo.LC1Chairman, &demo.LC1Tel); err != nil {
		demo = nil
	}
	inv = &models.MeaslesInvestigators{}
	if err := db.QueryRow(`SELECT id, patient_id, investigator_name, investigator_title, investigator_date FROM measles_investigators WHERE patient_id = $1`, patient.PatientID).
		Scan(&inv.ID, &inv.PatientID, &inv.InvestigatorName, &inv.InvestigatorTitle, &inv.InvestigatorDate); err != nil {
		inv = nil
	}
	hist = &models.MeaslesClinicalHistory{}
	if err := db.QueryRow(`SELECT id, patient_id, fever, fever_onset, temperature, rash, rash_onset, cough, red_eyes, running_nose, other_complications, complications_specify, outcome, vitamin_a, vitamin_a_doses, immunisation_card_seen, measles_doses, last_measles_vaccination, vaccination_reason, diagnosis FROM measles_clinical_history WHERE patient_id = $1`, patient.PatientID).
		Scan(&hist.ID, &hist.PatientID, &hist.Fever, &hist.FeverOnset, &hist.Temperature, &hist.Rash, &hist.RashOnset, &hist.Cough, &hist.RedEyes, &hist.RunningNose, &hist.OtherComplications, &hist.ComplicationsSpecify, &hist.Outcome, &hist.VitaminA, &hist.VitaminADoses, &hist.ImmunisationCardSeen, &hist.MeaslesDoses, &hist.LastMeaslesVaccination, &hist.VaccinationReason, &hist.Diagnosis); err != nil {
		hist = nil
	}
	res = &models.MeaslesResults{}
	if err := db.QueryRow(`SELECT id, patient_id, serology_igm, serology_date, serology_epi_sent_date, virus_isolation_urine, virus_isolation_date, final_classification, results_sent_date FROM measles_results WHERE patient_id = $1`, patient.PatientID).
		Scan(&res.ID, &res.PatientID, &res.SerologyIgM, &res.SerologyDate, &res.SerologyEpiSentDate, &res.VirusIsolationUrine, &res.VirusIsolationDate, &res.FinalClassification, &res.ResultsSentDate); err != nil {
		res = nil
	}
	spec = &models.MeaslesSpecimens{}
	if err := db.QueryRow(`SELECT id, patient_id, blood_collection_date, blood_sent_date, blood_received_date, blood_condition, urine_collection_date, urine_sent_date, urine_received_date, urine_condition, form_sent_date, form_received_date FROM measles_specimens WHERE patient_id = $1`, patient.PatientID).
		Scan(&spec.ID, &spec.PatientID, &spec.BloodCollectionDate, &spec.BloodSentDate, &spec.BloodReceivedDate, &spec.BloodCondition, &spec.UrineCollectionDate, &spec.UrineSentDate, &spec.UrineReceivedDate, &spec.UrineCondition, &spec.FormSentDate, &spec.FormReceivedDate); err != nil {
		spec = nil
	}
	return c.JSON(fiber.Map{
		"patient":          patient,
		"demographics":     demo,
		"investigators":    inv,
		"clinical_history": hist,
		"results":          res,
		"specimens":        spec,
	})
}
func HandlerMeaslesCIFByCaseCode(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	userID, userRole := GetUser(c, sl, store)
	if userRole == "" || userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}
	code := c.Query("case_code")
	if code == "" {
		return c.Status(400).JSON(fiber.Map{"error": "case_code is required"})
	}
	var pid string
	if err := db.QueryRow(`SELECT patient_id FROM measles_patients WHERE measles_code = $1`, code).Scan(&pid); err != nil {
		if err == sql.ErrNoRows {
			return c.Status(404).JSON(fiber.Map{"error": "CIF not found"})
		}
		return c.Status(500).JSON(fiber.Map{"error": "Lookup failed"})
	}
	// Build same payload as ID variant
	var patient models.MeaslesPatient
	if err := db.QueryRow(`SELECT patient_id, measles_code, patient_name, sex, dob, created_at FROM measles_patients WHERE patient_id = $1`, pid).
		Scan(&patient.PatientID, &patient.MeaslesCode, &patient.PatientName, &patient.Sex, &patient.DOB, &patient.CreatedAt); err != nil {
		if err == sql.ErrNoRows {
			return c.Status(404).JSON(fiber.Map{"error": "CIF not found"})
		}
		return c.Status(500).JSON(fiber.Map{"error": "Failed to load patient"})
	}
	var demo *models.MeaslesDemographics
	var inv *models.MeaslesInvestigators
	var hist *models.MeaslesClinicalHistory
	var res *models.MeaslesResults
	var spec *models.MeaslesSpecimens
	demo = &models.MeaslesDemographics{}
	if err := db.QueryRow(`SELECT id, patient_id, onset_district, reporting_unit, age_months, head_of_household, guardian_occupation, home_district, subcounty, parish, lc1_zone, lc1_chairman, lc1_tel FROM measles_demographics WHERE patient_id = $1`, patient.PatientID).
		Scan(&demo.ID, &demo.PatientID, &demo.OnsetDistrict, &demo.ReportingUnit, &demo.AgeMonths, &demo.HeadOfHousehold, &demo.GuardianOccupation, &demo.HomeDistrict, &demo.Subcounty, &demo.Parish, &demo.LC1Zone, &demo.LC1Chairman, &demo.LC1Tel); err != nil {
		demo = nil
	}
	inv = &models.MeaslesInvestigators{}
	if err := db.QueryRow(`SELECT id, patient_id, investigator_name, investigator_title, investigator_date FROM measles_investigators WHERE patient_id = $1`, patient.PatientID).
		Scan(&inv.ID, &inv.PatientID, &inv.InvestigatorName, &inv.InvestigatorTitle, &inv.InvestigatorDate); err != nil {
		inv = nil
	}
	hist = &models.MeaslesClinicalHistory{}
	if err := db.QueryRow(`SELECT id, patient_id, fever, fever_onset, temperature, rash, rash_onset, cough, red_eyes, running_nose, other_complications, complications_specify, outcome, vitamin_a, vitamin_a_doses, immunisation_card_seen, measles_doses, last_measles_vaccination, vaccination_reason, diagnosis FROM measles_clinical_history WHERE patient_id = $1`, patient.PatientID).
		Scan(&hist.ID, &hist.PatientID, &hist.Fever, &hist.FeverOnset, &hist.Temperature, &hist.Rash, &hist.RashOnset, &hist.Cough, &hist.RedEyes, &hist.RunningNose, &hist.OtherComplications, &hist.ComplicationsSpecify, &hist.Outcome, &hist.VitaminA, &hist.VitaminADoses, &hist.ImmunisationCardSeen, &hist.MeaslesDoses, &hist.LastMeaslesVaccination, &hist.VaccinationReason, &hist.Diagnosis); err != nil {
		hist = nil
	}
	res = &models.MeaslesResults{}
	if err := db.QueryRow(`SELECT id, patient_id, serology_igm, serology_date, serology_epi_sent_date, virus_isolation_urine, virus_isolation_date, final_classification, results_sent_date FROM measles_results WHERE patient_id = $1`, patient.PatientID).
		Scan(&res.ID, &res.PatientID, &res.SerologyIgM, &res.SerologyDate, &res.SerologyEpiSentDate, &res.VirusIsolationUrine, &res.VirusIsolationDate, &res.FinalClassification, &res.ResultsSentDate); err != nil {
		res = nil
	}
	spec = &models.MeaslesSpecimens{}
	if err := db.QueryRow(`SELECT id, patient_id, blood_collection_date, blood_sent_date, blood_received_date, blood_condition, urine_collection_date, urine_sent_date, urine_received_date, urine_condition, form_sent_date, form_received_date FROM measles_specimens WHERE patient_id = $1`, patient.PatientID).
		Scan(&spec.ID, &spec.PatientID, &spec.BloodCollectionDate, &spec.BloodSentDate, &spec.BloodReceivedDate, &spec.BloodCondition, &spec.UrineCollectionDate, &spec.UrineSentDate, &spec.UrineReceivedDate, &spec.UrineCondition, &spec.FormSentDate, &spec.FormReceivedDate); err != nil {
		spec = nil
	}
	return c.JSON(fiber.Map{
		"patient":          patient,
		"demographics":     demo,
		"investigators":    inv,
		"clinical_history": hist,
		"results":          res,
		"specimens":        spec,
	})
}

// Polio (placeholder)
func HandlerPolioCIFByID(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	userID, userRole := GetUser(c, sl, store)
	if userRole == "" || userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}
	caseID := c.Params("id")
	if caseID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Case ID required"})
	}
	var main models.PolioCaseInvestigation
	if err := db.QueryRow(`SELECT id, case_id, epid_number, country, region_province, district, year_onset, case_number, received_date, created_at FROM polio_case_investigation WHERE case_id = $1`, caseID).
		Scan(&main.ID, &main.CaseID, &main.EpidNumber, &main.Country, &main.RegionProvince, &main.District, &main.YearOnset, &main.CaseNumber, &main.ReceivedDate, &main.CreatedAt); err != nil {
		if err == sql.ErrNoRows {
			return c.Status(404).JSON(fiber.Map{"error": "CIF not found"})
		}
		return c.Status(500).JSON(fiber.Map{"error": "Failed to load case"})
	}
	var ident *models.PolioIdentification
	var notif *models.PolioNotificationInvestigation
	var hosp *models.PolioHospitalization
	var clin *models.PolioClinicalHistory
	var imm *models.PolioImmunizationHistory
	var stool *models.PolioStoolSpecimenCollection
	var stoolRes *models.PolioStoolSpecimenResults
	var follow *models.PolioFollowUpExamination
	var history *models.PolioPatientHistory
	var investigator *models.PolioInvestigator
	ident = &models.PolioIdentification{}
	if err := db.QueryRow(`SELECT id, case_id, district, region_province, address, village, city, nearest_health_facility, longitude, latitude, patient_name, father_mother, phone_number, date_of_birth, age_years, age_months, sex FROM polio_identification WHERE case_id = $1`, caseID).
		Scan(&ident.ID, &ident.CaseID, &ident.District, &ident.RegionProvince, &ident.Address, &ident.Village, &ident.City, &ident.NearestHealthFacility, &ident.Longitude, &ident.Latitude, &ident.PatientName, &ident.FatherMother, &ident.PhoneNumber, &ident.DateOfBirth, &ident.AgeYears, &ident.AgeMonths, &ident.Sex); err != nil {
		ident = nil
	}
	notif = &models.PolioNotificationInvestigation{}
	if err := db.QueryRow(`SELECT id, case_id, notified_by, date_of_notification, date_of_investigation FROM polio_notification_investigation WHERE case_id = $1`, caseID).
		Scan(&notif.ID, &notif.CaseID, &notif.NotifiedBy, &notif.DateOfNotification, &notif.DateOfInvestigation); err != nil {
		notif = nil
	}
	hosp = &models.PolioHospitalization{}
	if err := db.QueryRow(`SELECT id, case_id, hospitalized, date_of_admission, hospital_record_number, hospital_name_address FROM polio_hospitalization WHERE case_id = $1`, caseID).
		Scan(&hosp.ID, &hosp.CaseID, &hosp.Hospitalized, &hosp.DateOfAdmission, &hosp.HospitalRecordNumber, &hosp.HospitalNameAddress); err != nil {
		hosp = nil
	}
	clin = &models.PolioClinicalHistory{}
	if err := db.QueryRow(`SELECT id, case_id, fever_at_onset, date_onset_of_fever, progressive_paralysis, date_onset_of_paralysis, flaccid_acute_paralysis, sensation_loss, sudden_onset, "asymmetric", left_arm_paralysis, right_arm_paralysis, left_leg_paralysis, right_leg_paralysis, diminished_reflexes, diminished_muscle_tone, muscle_wasting, muscle_weakness, respiratory_muscles, face, stiff_neck, convulsions, headache, vomiting, diarrhoea, other_sites, recent_injection, total_injections, injection_type, paralyzed_limb_sensitive, injection_facility_name, provisional_diagnosis, true_afp_case FROM polio_clinical_history WHERE case_id = $1`, caseID).
		Scan(&clin.ID, &clin.CaseID, &clin.FeverAtOnset, &clin.DateOnsetOfFever, &clin.ProgressiveParalysis, &clin.DateOnsetOfParalysis, &clin.FlaccidAcuteParalysis, &clin.SensationLoss, &clin.SuddenOnset, &clin.Asymmetric, &clin.LeftArmParalysis, &clin.RightArmParalysis, &clin.LeftLegParalysis, &clin.RightLegParalysis, &clin.DiminishedReflexes, &clin.DiminishedMuscleTone, &clin.MuscleWasting, &clin.MuscleWeakness, &clin.RespiratoryMuscles, &clin.Face, &clin.StiffNeck, &clin.Convulsions, &clin.Headache, &clin.Vomiting, &clin.Diarrhoea, &clin.OtherSites, &clin.RecentInjection, &clin.TotalInjections, &clin.InjectionType, &clin.ParalyzedLimbSensitive, &clin.InjectionFacilityName, &clin.ProvisionalDiagnosis, &clin.TrueAFPCase); err != nil {
		clin = nil
	}
	imm = &models.PolioImmunizationHistory{}
	if err := db.QueryRow(`SELECT id, case_id, total_polio_doses, exclude_dose_at_birth, opv_dose_at_birth, opv_dose1, opv_dose2, opv_dose3, opv_dose4, opv_dose_more_than4, last_opv_dose, total_opv_sia, last_opv_sia, total_opv_ri, total_ipv_sia, total_ipv_ri, last_ipv_sia, source_of_ri_vaccination, unknown_zero_dose_reasons FROM polio_immunization_history WHERE case_id = $1`, caseID).
		Scan(&imm.ID, &imm.CaseID, &imm.TotalPolioDoses, &imm.ExcludeDoseAtBirth, &imm.OPVDoseAtBirth, &imm.OPVDose1, &imm.OPVDose2, &imm.OPVDose3, &imm.OPVDose4, &imm.OPVDoseMoreThan4, &imm.LastOPVDose, &imm.TotalOPVSIA, &imm.LastOPVSIA, &imm.TotalOPVRI, &imm.TotalIPVSIA, &imm.TotalIPVRI, &imm.LastIPVSIA, &imm.SourceOfRIVaccination, &imm.UnknownZeroDoseReasons); err != nil {
		imm = nil
	}
	stool = &models.PolioStoolSpecimenCollection{}
	if err := db.QueryRow(`SELECT id, case_id, date_first_specimen, date_second_specimen, date_specimen_sent_national, date_specimen_received_national, date_specimen_sent_lab FROM polio_stool_specimen_collection WHERE case_id = $1`, caseID).
		Scan(&stool.ID, &stool.CaseID, &stool.DateFirstSpecimen, &stool.DateSecondSpecimen, &stool.DateSpecimenSentNational, &stool.DateSpecimenReceivedNational, &stool.DateSpecimenSentLab); err != nil {
		stool = nil
	}
	stoolRes = &models.PolioStoolSpecimenResults{}
	if err := db.QueryRow(`SELECT id, case_id, date_received_at_lab, specimen_status_at_reception, date_combined_cell_culture, date_results_sent_to_epi, date_results_received_at_epi, final_cell_culture_results, w1, w2, w3, discordant_sabin, sl1, sl2, sl3, r_npent, nev, date_sent_to_regional_lab, date_it_differentiation_sent, date_it_differentiation_received, date_isolate_sent_sequencing, date_seq_results_sent_program FROM polio_stool_specimen_results WHERE case_id = $1`, caseID).
		Scan(&stoolRes.ID, &stoolRes.CaseID, &stoolRes.DateReceivedAtLab, &stoolRes.SpecimenStatusAtReception, &stoolRes.DateCombinedCellCulture, &stoolRes.DateResultsSentToEPI, &stoolRes.DateResultsReceivedAtEPI, &stoolRes.FinalCellCultureResults, &stoolRes.W1, &stoolRes.W2, &stoolRes.W3, &stoolRes.DiscordantSabin, &stoolRes.SL1, &stoolRes.SL2, &stoolRes.SL3, &stoolRes.RNPENT, &stoolRes.NEV, &stoolRes.DateSentToRegionalLab, &stoolRes.DateITDifferentiationSent, &stoolRes.DateITDifferentiationReceived, &stoolRes.DateIsolateSentSequencing, &stoolRes.DateSeqResultsSentProgram); err != nil {
		stoolRes = nil
	}
	follow = &models.PolioFollowUpExamination{}
	if err := db.QueryRow(`SELECT id, case_id, date_of_follow_up, residual_paralysis_la, residual_paralysis_ra, residual_paralysis_ll, residual_paralysis_rl, results_of_exam, immunocompromised_status, final_classification, cvdpv, avdpv, ivdpv, serotype FROM polio_follow_up_examination WHERE case_id = $1`, caseID).
		Scan(&follow.ID, &follow.CaseID, &follow.DateOfFollowUp, &follow.ResidualParalysisLA, &follow.ResidualParalysisRA, &follow.ResidualParalysisLL, &follow.ResidualParalysisRL, &follow.ResultsOfExam, &follow.ImmunocompromisedStatus, &follow.FinalClassification, &follow.CVDPV, &follow.AVDPV, &follow.IVDPV, &follow.Serotype); err != nil {
		follow = nil
	}
	history = &models.PolioPatientHistory{}
	if err := db.QueryRow(`SELECT id, case_id, place1, duration1_months, duration1_days, place2, duration2_months, duration2_days, place3, duration3_months, duration3_days, place4, duration4_months, duration4_days FROM polio_patient_history WHERE case_id = $1`, caseID).
		Scan(&history.ID, &history.CaseID, &history.Place1, &history.Duration1Months, &history.Duration1Days, &history.Place2, &history.Duration2Months, &history.Duration2Days, &history.Place3, &history.Duration3Months, &history.Duration3Days, &history.Place4, &history.Duration4Months, &history.Duration4Days); err != nil {
		history = nil
	}
	investigator = &models.PolioInvestigator{}
	if err := db.QueryRow(`SELECT id, case_id, investigator_name, investigator_title, unit, address, telephone FROM polio_investigator WHERE case_id = $1`, caseID).
		Scan(&investigator.ID, &investigator.CaseID, &investigator.InvestigatorName, &investigator.InvestigatorTitle, &investigator.Unit, &investigator.Address, &investigator.Telephone); err != nil {
		investigator = nil
	}
	return c.JSON(fiber.Map{
		"case":             main,
		"identification":   ident,
		"notification":     notif,
		"hospitalization":  hosp,
		"clinical_history": clin,
		"immunization":     imm,
		"stool_collection": stool,
		"stool_results":    stoolRes,
		"follow_up":        follow,
		"patient_history":  history,
		"investigator":     investigator,
	})
}
func HandlerPolioCIFByCaseCode(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	userID, userRole := GetUser(c, sl, store)
	if userRole == "" || userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}
	code := c.Query("case_code")
	if code == "" {
		return c.Status(400).JSON(fiber.Map{"error": "case_code is required"})
	}
	caseID := code
	var main models.PolioCaseInvestigation
	if err := db.QueryRow(`SELECT id, case_id, epid_number, country, region_province, district, year_onset, case_number, received_date, created_at FROM polio_case_investigation WHERE case_id = $1`, caseID).
		Scan(&main.ID, &main.CaseID, &main.EpidNumber, &main.Country, &main.RegionProvince, &main.District, &main.YearOnset, &main.CaseNumber, &main.ReceivedDate, &main.CreatedAt); err != nil {
		if err == sql.ErrNoRows {
			return c.Status(404).JSON(fiber.Map{"error": "CIF not found"})
		}
		return c.Status(500).JSON(fiber.Map{"error": "Failed to load case"})
	}
	var ident *models.PolioIdentification
	var notif *models.PolioNotificationInvestigation
	var hosp *models.PolioHospitalization
	var clin *models.PolioClinicalHistory
	var imm *models.PolioImmunizationHistory
	var stool *models.PolioStoolSpecimenCollection
	var stoolRes *models.PolioStoolSpecimenResults
	var follow *models.PolioFollowUpExamination
	var history *models.PolioPatientHistory
	var investigator *models.PolioInvestigator
	ident = &models.PolioIdentification{}
	if err := db.QueryRow(`SELECT id, case_id, district, region_province, address, village, city, nearest_health_facility, longitude, latitude, patient_name, father_mother, phone_number, date_of_birth, age_years, age_months, sex FROM polio_identification WHERE case_id = $1`, caseID).
		Scan(&ident.ID, &ident.CaseID, &ident.District, &ident.RegionProvince, &ident.Address, &ident.Village, &ident.City, &ident.NearestHealthFacility, &ident.Longitude, &ident.Latitude, &ident.PatientName, &ident.FatherMother, &ident.PhoneNumber, &ident.DateOfBirth, &ident.AgeYears, &ident.AgeMonths, &ident.Sex); err != nil {
		ident = nil
	}
	notif = &models.PolioNotificationInvestigation{}
	if err := db.QueryRow(`SELECT id, case_id, notified_by, date_of_notification, date_of_investigation FROM polio_notification_investigation WHERE case_id = $1`, caseID).
		Scan(&notif.ID, &notif.CaseID, &notif.NotifiedBy, &notif.DateOfNotification, &notif.DateOfInvestigation); err != nil {
		notif = nil
	}
	hosp = &models.PolioHospitalization{}
	if err := db.QueryRow(`SELECT id, case_id, hospitalized, date_of_admission, hospital_record_number, hospital_name_address FROM polio_hospitalization WHERE case_id = $1`, caseID).
		Scan(&hosp.ID, &hosp.CaseID, &hosp.Hospitalized, &hosp.DateOfAdmission, &hosp.HospitalRecordNumber, &hosp.HospitalNameAddress); err != nil {
		hosp = nil
	}
	clin = &models.PolioClinicalHistory{}
	if err := db.QueryRow(`SELECT id, case_id, fever_at_onset, date_onset_of_fever, progressive_paralysis, date_onset_of_paralysis, flaccid_acute_paralysis, sensation_loss, sudden_onset, "asymmetric", left_arm_paralysis, right_arm_paralysis, left_leg_paralysis, right_leg_paralysis, diminished_reflexes, diminished_muscle_tone, muscle_wasting, muscle_weakness, respiratory_muscles, face, stiff_neck, convulsions, headache, vomiting, diarrhoea, other_sites, recent_injection, total_injections, injection_type, paralyzed_limb_sensitive, injection_facility_name, provisional_diagnosis, true_afp_case FROM polio_clinical_history WHERE case_id = $1`, caseID).
		Scan(&clin.ID, &clin.CaseID, &clin.FeverAtOnset, &clin.DateOnsetOfFever, &clin.ProgressiveParalysis, &clin.DateOnsetOfParalysis, &clin.FlaccidAcuteParalysis, &clin.SensationLoss, &clin.SuddenOnset, &clin.Asymmetric, &clin.LeftArmParalysis, &clin.RightArmParalysis, &clin.LeftLegParalysis, &clin.RightLegParalysis, &clin.DiminishedReflexes, &clin.DiminishedMuscleTone, &clin.MuscleWasting, &clin.MuscleWeakness, &clin.RespiratoryMuscles, &clin.Face, &clin.StiffNeck, &clin.Convulsions, &clin.Headache, &clin.Vomiting, &clin.Diarrhoea, &clin.OtherSites, &clin.RecentInjection, &clin.TotalInjections, &clin.InjectionType, &clin.ParalyzedLimbSensitive, &clin.InjectionFacilityName, &clin.ProvisionalDiagnosis, &clin.TrueAFPCase); err != nil {
		clin = nil
	}
	imm = &models.PolioImmunizationHistory{}
	if err := db.QueryRow(`SELECT id, case_id, total_polio_doses, exclude_dose_at_birth, opv_dose_at_birth, opv_dose1, opv_dose2, opv_dose3, opv_dose4, opv_dose_more_than4, last_opv_dose, total_opv_sia, last_opv_sia, total_opv_ri, total_ipv_sia, total_ipv_ri, last_ipv_sia, source_of_ri_vaccination, unknown_zero_dose_reasons FROM polio_immunization_history WHERE case_id = $1`, caseID).
		Scan(&imm.ID, &imm.CaseID, &imm.TotalPolioDoses, &imm.ExcludeDoseAtBirth, &imm.OPVDoseAtBirth, &imm.OPVDose1, &imm.OPVDose2, &imm.OPVDose3, &imm.OPVDose4, &imm.OPVDoseMoreThan4, &imm.LastOPVDose, &imm.TotalOPVSIA, &imm.LastOPVSIA, &imm.TotalOPVRI, &imm.TotalIPVSIA, &imm.TotalIPVRI, &imm.LastIPVSIA, &imm.SourceOfRIVaccination, &imm.UnknownZeroDoseReasons); err != nil {
		imm = nil
	}
	stool = &models.PolioStoolSpecimenCollection{}
	if err := db.QueryRow(`SELECT id, case_id, date_first_specimen, date_second_specimen, date_specimen_sent_national, date_specimen_received_national, date_specimen_sent_lab FROM polio_stool_specimen_collection WHERE case_id = $1`, caseID).
		Scan(&stool.ID, &stool.CaseID, &stool.DateFirstSpecimen, &stool.DateSecondSpecimen, &stool.DateSpecimenSentNational, &stool.DateSpecimenReceivedNational, &stool.DateSpecimenSentLab); err != nil {
		stool = nil
	}
	stoolRes = &models.PolioStoolSpecimenResults{}
	if err := db.QueryRow(`SELECT id, case_id, date_received_at_lab, specimen_status_at_reception, date_combined_cell_culture, date_results_sent_to_epi, date_results_received_at_epi, final_cell_culture_results, w1, w2, w3, discordant_sabin, sl1, sl2, sl3, r_npent, nev, date_sent_to_regional_lab, date_it_differentiation_sent, date_it_differentiation_received, date_isolate_sent_sequencing, date_seq_results_sent_program FROM polio_stool_specimen_results WHERE case_id = $1`, caseID).
		Scan(&stoolRes.ID, &stoolRes.CaseID, &stoolRes.DateReceivedAtLab, &stoolRes.SpecimenStatusAtReception, &stoolRes.DateCombinedCellCulture, &stoolRes.DateResultsSentToEPI, &stoolRes.DateResultsReceivedAtEPI, &stoolRes.FinalCellCultureResults, &stoolRes.W1, &stoolRes.W2, &stoolRes.W3, &stoolRes.DiscordantSabin, &stoolRes.SL1, &stoolRes.SL2, &stoolRes.SL3, &stoolRes.RNPENT, &stoolRes.NEV, &stoolRes.DateSentToRegionalLab, &stoolRes.DateITDifferentiationSent, &stoolRes.DateITDifferentiationReceived, &stoolRes.DateIsolateSentSequencing, &stoolRes.DateSeqResultsSentProgram); err != nil {
		stoolRes = nil
	}
	follow = &models.PolioFollowUpExamination{}
	if err := db.QueryRow(`SELECT id, case_id, date_of_follow_up, residual_paralysis_la, residual_paralysis_ra, residual_paralysis_ll, residual_paralysis_rl, results_of_exam, immunocompromised_status, final_classification, cvdpv, avdpv, ivdpv, serotype FROM polio_follow_up_examination WHERE case_id = $1`, caseID).
		Scan(&follow.ID, &follow.CaseID, &follow.DateOfFollowUp, &follow.ResidualParalysisLA, &follow.ResidualParalysisRA, &follow.ResidualParalysisLL, &follow.ResidualParalysisRL, &follow.ResultsOfExam, &follow.ImmunocompromisedStatus, &follow.FinalClassification, &follow.CVDPV, &follow.AVDPV, &follow.IVDPV, &follow.Serotype); err != nil {
		follow = nil
	}
	history = &models.PolioPatientHistory{}
	if err := db.QueryRow(`SELECT id, case_id, place1, duration1_months, duration1_days, place2, duration2_months, duration2_days, place3, duration3_months, duration3_days, place4, duration4_months, duration4_days FROM polio_patient_history WHERE case_id = $1`, caseID).
		Scan(&history.ID, &history.CaseID, &history.Place1, &history.Duration1Months, &history.Duration1Days, &history.Place2, &history.Duration2Months, &history.Duration2Days, &history.Place3, &history.Duration3Months, &history.Duration3Days, &history.Place4, &history.Duration4Months, &history.Duration4Days); err != nil {
		history = nil
	}
	investigator = &models.PolioInvestigator{}
	if err := db.QueryRow(`SELECT id, case_id, investigator_name, investigator_title, unit, address, telephone FROM polio_investigator WHERE case_id = $1`, caseID).
		Scan(&investigator.ID, &investigator.CaseID, &investigator.InvestigatorName, &investigator.InvestigatorTitle, &investigator.Unit, &investigator.Address, &investigator.Telephone); err != nil {
		investigator = nil
	}
	return c.JSON(fiber.Map{
		"case":             main,
		"identification":   ident,
		"notification":     notif,
		"hospitalization":  hosp,
		"clinical_history": clin,
		"immunization":     imm,
		"stool_collection": stool,
		"stool_results":    stoolRes,
		"follow_up":        follow,
		"patient_history":  history,
		"investigator":     investigator,
	})
}

// Mpox (placeholder)
func HandlerMpoxCIFByID(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	userID, userRole := GetUser(c, sl, store)
	if userRole == "" || userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}
	caseID := c.Params("id")
	if caseID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Case ID required"})
	}
	var ci *models.MpoxCaseInvestigation
	var demo *models.MpoxPatientDemographics
	var clin *models.MpoxClinicianInfo
	var exp *models.MpoxCaseExposureHistory
	var man *models.MpoxClinicalManifestations
	var travel *models.MpoxTravelHistory
	var lab *models.MpoxLabInvestigation
	ci = &models.MpoxCaseInvestigation{}
	if err := db.QueryRow(`SELECT id, case_id, date, case_status, case_classification FROM mpox_case_investigation WHERE case_id = $1`, caseID).
		Scan(&ci.ID, &ci.CaseID, &ci.Date, &ci.CaseStatus, &ci.CaseClassification); err != nil {
		ci = nil
	}
	demo = &models.MpoxPatientDemographics{}
	if err := db.QueryRow(`SELECT id, case_id, health_facility_case_id, surname, other_names, sex, date_of_birth, age, parish, sub_county, physical_address, contact_telephone, occupation, nationality, vaccination_status, date_of_vaccination, next_of_kin, next_of_kin_contact, marital_status, if_dead_date_of_death, admission_date, onset_date, rash_onset_date FROM mpox_patient_demographics WHERE case_id = $1`, caseID).
		Scan(&demo.ID, &demo.CaseID, &demo.HealthFacilityCaseID, &demo.Surname, &demo.OtherNames, &demo.Sex, &demo.DateOfBirth, &demo.Age, &demo.Parish, &demo.SubCounty, &demo.PhysicalAddress, &demo.ContactTelephone, &demo.Occupation, &demo.Nationality, &demo.VaccinationStatus, &demo.DateOfVaccination, &demo.NextOfKin, &demo.NextOfKinContact, &demo.MaritalStatus, &demo.IfDeadDateOfDeath, &demo.AdmissionDate, &demo.OnsetDate, &demo.RashOnsetDate); err != nil {
		demo = nil
	}
	clin = &models.MpoxClinicianInfo{}
	if err := db.QueryRow(`SELECT id, case_id, clinician_name, clinician_contact, facility_name, clinician_email, facility_district, pdpid_number, admission_date, ward FROM mpox_clinician_info WHERE case_id = $1`, caseID).
		Scan(&clin.ID, &clin.CaseID, &clin.ClinicianName, &clin.ClinicianContact, &clin.FacilityName, &clin.ClinicianEmail, &clin.FacilityDistrict, &clin.PDPIDNumber, &clin.AdmissionDate, &clin.Ward); err != nil {
		clin = nil
	}
	exp = &models.MpoxCaseExposureHistory{}
	if err := db.QueryRow(`SELECT id, case_id, traveled_country_reported_mpox, close_contact_mpox, intl_travel, contact_animals, domestic_wild_animals, sexual_exposure FROM mpox_case_exposure_history WHERE case_id = $1`, caseID).
		Scan(&exp.ID, &exp.CaseID, &exp.TraveledCountryReportedMpox, &exp.CloseContactMpox, &exp.IntlTravel, &exp.ContactAnimals, &exp.DomesticWildAnimals, &exp.SexualExposure); err != nil {
		exp = nil
	}
	man = &models.MpoxClinicalManifestations{}
	if err := db.QueryRow(`SELECT id, case_id, onset_date, fever, fever_temperature, lymphadenopathy, symptoms, symptom_other_specify, nausea_vomiting, pregnant, pregnant_trimester, vaccinated, vaccination_date, rash, rash_onset_date, rash_distribution, rash_type, underlying_illness, underlying_illness_details FROM mpox_clinical_manifestations WHERE case_id = $1`, caseID).
		Scan(&man.ID, &man.CaseID, &man.OnsetDate, &man.Fever, &man.FeverTemperature, &man.Lymphadenopathy, &man.Symptoms, &man.SymptomOtherSpecify, &man.NauseaVomiting, &man.Pregnant, &man.PregnantTrimester, &man.Vaccinated, &man.VaccinationDate, &man.Rash, &man.RashOnsetDate, &man.RashDistribution, &man.RashType, &man.UnderlyingIllness, &man.UnderlyingIllnessDetails); err != nil {
		man = nil
	}
	travel = &models.MpoxTravelHistory{}
	if err := db.QueryRow(`SELECT id, case_id, travel_outside_uganda, country_visited, location_visited, date_arrival, date_departure, activities_location FROM mpox_travel_history WHERE case_id = $1`, caseID).
		Scan(&travel.ID, &travel.CaseID, &travel.TravelOutsideUganda, &travel.CountryVisited, &travel.LocationVisited, &travel.DateArrival, &travel.DateDeparture, &travel.ActivitiesLocation); err != nil {
		travel = nil
	}
	lab = &models.MpoxLabInvestigation{}
	if err := db.QueryRow(`SELECT id, case_id, lab_id, sample_collected, sample_other_specify, test_requested, test_other_specify, date_sample_collection, time_sample_collection, date_sample_dispatch, sample_collector_name, sample_collector_phone, date_sample_reception, time_sample_reception, sample_recipient_name, sample_recipient_phone, genomic_characterization, clade, accession_number FROM mpox_lab_investigation WHERE case_id = $1`, caseID).
		Scan(&lab.ID, &lab.CaseID, &lab.LabID, &lab.SampleCollected, &lab.SampleOtherSpecify, &lab.TestRequested, &lab.TestOtherSpecify, &lab.DateSampleCollection, &lab.TimeSampleCollection, &lab.DateSampleDispatch, &lab.SampleCollectorName, &lab.SampleCollectorPhone, &lab.DateSampleReception, &lab.TimeSampleReception, &lab.SampleRecipientName, &lab.SampleRecipientPhone, &lab.GenomicCharacterization, &lab.Clade, &lab.AccessionNumber); err != nil {
		lab = nil
	}
	return c.JSON(fiber.Map{
		"case":           ci,
		"demographics":   demo,
		"clinician":      clin,
		"exposure":       exp,
		"manifestations": man,
		"travel_history": travel,
		"laboratory":     lab,
	})
}
func HandlerMpoxCIFByCaseCode(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	userID, userRole := GetUser(c, sl, store)
	if userRole == "" || userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}
	code := c.Query("case_code")
	if code == "" {
		return c.Status(400).JSON(fiber.Map{"error": "case_code is required"})
	}
	caseID := code
	var ci *models.MpoxCaseInvestigation
	var demo *models.MpoxPatientDemographics
	var clin *models.MpoxClinicianInfo
	var exp *models.MpoxCaseExposureHistory
	var man *models.MpoxClinicalManifestations
	var travel *models.MpoxTravelHistory
	var lab *models.MpoxLabInvestigation
	ci = &models.MpoxCaseInvestigation{}
	if err := db.QueryRow(`SELECT id, case_id, date, case_status, case_classification FROM mpox_case_investigation WHERE case_id = $1`, caseID).
		Scan(&ci.ID, &ci.CaseID, &ci.Date, &ci.CaseStatus, &ci.CaseClassification); err != nil {
		ci = nil
	}
	demo = &models.MpoxPatientDemographics{}
	if err := db.QueryRow(`SELECT id, case_id, health_facility_case_id, surname, other_names, sex, date_of_birth, age, parish, sub_county, physical_address, contact_telephone, occupation, nationality, vaccination_status, date_of_vaccination, next_of_kin, next_of_kin_contact, marital_status, if_dead_date_of_death, admission_date, onset_date, rash_onset_date FROM mpox_patient_demographics WHERE case_id = $1`, caseID).
		Scan(&demo.ID, &demo.CaseID, &demo.HealthFacilityCaseID, &demo.Surname, &demo.OtherNames, &demo.Sex, &demo.DateOfBirth, &demo.Age, &demo.Parish, &demo.SubCounty, &demo.PhysicalAddress, &demo.ContactTelephone, &demo.Occupation, &demo.Nationality, &demo.VaccinationStatus, &demo.DateOfVaccination, &demo.NextOfKin, &demo.NextOfKinContact, &demo.MaritalStatus, &demo.IfDeadDateOfDeath, &demo.AdmissionDate, &demo.OnsetDate, &demo.RashOnsetDate); err != nil {
		demo = nil
	}
	clin = &models.MpoxClinicianInfo{}
	if err := db.QueryRow(`SELECT id, case_id, clinician_name, clinician_contact, facility_name, clinician_email, facility_district, pdpid_number, admission_date, ward FROM mpox_clinician_info WHERE case_id = $1`, caseID).
		Scan(&clin.ID, &clin.CaseID, &clin.ClinicianName, &clin.ClinicianContact, &clin.FacilityName, &clin.ClinicianEmail, &clin.FacilityDistrict, &clin.PDPIDNumber, &clin.AdmissionDate, &clin.Ward); err != nil {
		clin = nil
	}
	exp = &models.MpoxCaseExposureHistory{}
	if err := db.QueryRow(`SELECT id, case_id, traveled_country_reported_mpox, close_contact_mpox, intl_travel, contact_animals, domestic_wild_animals, sexual_exposure FROM mpox_case_exposure_history WHERE case_id = $1`, caseID).
		Scan(&exp.ID, &exp.CaseID, &exp.TraveledCountryReportedMpox, &exp.CloseContactMpox, &exp.IntlTravel, &exp.ContactAnimals, &exp.DomesticWildAnimals, &exp.SexualExposure); err != nil {
		exp = nil
	}
	man = &models.MpoxClinicalManifestations{}
	if err := db.QueryRow(`SELECT id, case_id, onset_date, fever, fever_temperature, lymphadenopathy, symptoms, symptom_other_specify, nausea_vomiting, pregnant, pregnant_trimester, vaccinated, vaccination_date, rash, rash_onset_date, rash_distribution, rash_type, underlying_illness, underlying_illness_details FROM mpox_clinical_manifestations WHERE case_id = $1`, caseID).
		Scan(&man.ID, &man.CaseID, &man.OnsetDate, &man.Fever, &man.FeverTemperature, &man.Lymphadenopathy, &man.Symptoms, &man.SymptomOtherSpecify, &man.NauseaVomiting, &man.Pregnant, &man.PregnantTrimester, &man.Vaccinated, &man.VaccinationDate, &man.Rash, &man.RashOnsetDate, &man.RashDistribution, &man.RashType, &man.UnderlyingIllness, &man.UnderlyingIllnessDetails); err != nil {
		man = nil
	}
	travel = &models.MpoxTravelHistory{}
	if err := db.QueryRow(`SELECT id, case_id, travel_outside_uganda, country_visited, location_visited, date_arrival, date_departure, activities_location FROM mpox_travel_history WHERE case_id = $1`, caseID).
		Scan(&travel.ID, &travel.CaseID, &travel.TravelOutsideUganda, &travel.CountryVisited, &travel.LocationVisited, &travel.DateArrival, &travel.DateDeparture, &travel.ActivitiesLocation); err != nil {
		travel = nil
	}
	lab = &models.MpoxLabInvestigation{}
	if err := db.QueryRow(`SELECT id, case_id, lab_id, sample_collected, sample_other_specify, test_requested, test_other_specify, date_sample_collection, time_sample_collection, date_sample_dispatch, sample_collector_name, sample_collector_phone, date_sample_reception, time_sample_reception, sample_recipient_name, sample_recipient_phone, genomic_characterization, clade, accession_number FROM mpox_lab_investigation WHERE case_id = $1`, caseID).
		Scan(&lab.ID, &lab.CaseID, &lab.LabID, &lab.SampleCollected, &lab.SampleOtherSpecify, &lab.TestRequested, &lab.TestOtherSpecify, &lab.DateSampleCollection, &lab.TimeSampleCollection, &lab.DateSampleDispatch, &lab.SampleCollectorName, &lab.SampleCollectorPhone, &lab.DateSampleReception, &lab.TimeSampleReception, &lab.SampleRecipientName, &lab.SampleRecipientPhone, &lab.GenomicCharacterization, &lab.Clade, &lab.AccessionNumber); err != nil {
		lab = nil
	}
	return c.JSON(fiber.Map{
		"case":           ci,
		"demographics":   demo,
		"clinician":      clin,
		"exposure":       exp,
		"manifestations": man,
		"travel_history": travel,
		"laboratory":     lab,
	})
}
