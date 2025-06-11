package handlers

import (
	"case/internal/models"
	"case/internal/services"
	"database/sql"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
)

// HandlerVHFPatientSubmit handles the submission of patient information
func HandlerVHFPatientSubmit(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config, smsService *services.SMSService) error {
	// Parse form data
	patient := &models.VHFPatient{
		Surname:                     c.FormValue("surname"),
		OtherNames:                  c.FormValue("other_names"),
		DateOfBirth:                 sql.NullTime{Time: parseDate(c.FormValue("dob")), Valid: true},
		AgeYears:                    sql.NullInt32{Int32: parseInt32(c.FormValue("age_years")), Valid: true},
		AgeMonths:                   sql.NullInt32{Int32: parseInt32(c.FormValue("age_months")), Valid: true},
		Gender:                      c.FormValue("gender"),
		PatientPhone:                c.FormValue("patient_phone"),
		PhoneOwner:                  c.FormValue("phone_owner"),
		NextOfKin:                   c.FormValue("next_of_kin"),
		NextOfKinPhone:              c.FormValue("next_of_kin_phone"),
		DataCapturerName:            sql.NullString{String: c.FormValue("data_capturer_name"), Valid: true},
		DataCapturerPhone:           c.FormValue("data_capturer_phone"),
		ReportingHealthFacilityName: c.FormValue("reporting_health_facility_name"),
		CaseCode:                    c.FormValue("case_code"),
		Status:                      c.FormValue("status"),
		HeadOfHousehold:             c.FormValue("head_of_household"),
		VillageTown:                 c.FormValue("village_town"),
		Parish:                      c.FormValue("parish"),
		Subcounty:                   c.FormValue("subcounty"),
		District:                    c.FormValue("district"),
		CountryOfResidence:          c.FormValue("country_of_residence"),
		Occupation:                  c.FormValue("occupation"),
		IllVillageTown:              c.FormValue("ill_village_town"),
		IllSubcounty:                c.FormValue("ill_subcounty"),
		IllDistrict:                 c.FormValue("ill_district"),
		CreatedAt:                   time.Now(),
	}

	// Save patient data
	if err := models.SaveVHFPatient(db, patient); err != nil {
		sl.Error("Failed to save patient data", "error", err)
		return c.Status(500).SendString("Failed to save patient data")
	}

	// Parse and save clinical signs
	clinicalSigns := &models.VHFClinicalSigns{
		PatientID:        patient.ID,
		DateInitialOnset: sql.NullTime{Time: parseDate(c.FormValue("date_initial_onset")), Valid: true},
		Fever:            sql.NullBool{Bool: parseBool(c.FormValue("fever")), Valid: true},
		DateFever:        sql.NullTime{Time: parseDate(c.FormValue("date_fever")), Valid: true},
		Temperature:      sql.NullFloat64{Float64: parseFloat(c.FormValue("temperature")), Valid: true},
		Vomiting:         sql.NullBool{Bool: parseBool(c.FormValue("vomiting")), Valid: true},
		DateVomiting:     sql.NullTime{Time: parseDate(c.FormValue("date_vomiting")), Valid: true},
		Diarrhea:         sql.NullBool{Bool: parseBool(c.FormValue("diarrhea")), Valid: true},
		DateDiarrhea:     sql.NullTime{Time: parseDate(c.FormValue("date_diarrhea")), Valid: true},
		CreatedAt:        time.Now(),
	}

	if err := models.SaveVHFClinicalSigns(db, clinicalSigns); err != nil {
		sl.Error("Failed to save clinical signs", "error", err)
		return c.Status(500).SendString("Failed to save clinical signs")
	}

	// Parse and save hospitalization data
	hospitalization := &models.VHFHospitalization{
		PatientID:          patient.ID,
		Hospitalized:       parseBool(c.FormValue("hospitalized")),
		AdmissionDate:      sql.NullTime{Time: parseDate(c.FormValue("admission_date")), Valid: true},
		HealthFacilityName: c.FormValue("health_facility_name"),
		InIsolation:        parseBool(c.FormValue("in_isolation")),
		IsolationDate:      sql.NullTime{Time: parseDate(c.FormValue("isolation_date")), Valid: true},
		CreatedAt:          time.Now(),
	}

	if err := models.SaveVHFHospitalization(db, hospitalization); err != nil {
		sl.Error("Failed to save hospitalization data", "error", err)
		return c.Status(500).SendString("Failed to save hospitalization data")
	}

	// Parse and save risk factors
	riskFactors := &models.VHFRiskFactors{
		PatientID:       patient.ID,
		ContactWithCase: sql.NullBool{Bool: parseBool(c.FormValue("contactWithCase")), Valid: true},
		ContactName:     c.FormValue("contact_name"),
		ContactRelation: c.FormValue("contact_relation"),
		ContactDates:    c.FormValue("contact_dates"),
		ContactVillage:  c.FormValue("contact_village"),
		ContactDistrict: c.FormValue("contact_district"),
		ContactStatus:   c.FormValue("contact_status"),
		ContactTypes:    c.FormValue("contact_types"),
		CreatedAt:       time.Now(),
	}

	// Parse contact death date if provided
	if deathDate := c.FormValue("contact_death_date"); deathDate != "" {
		if t, err := time.Parse("2006-01-02", deathDate); err == nil {
			riskFactors.ContactDeathDate = sql.NullTime{Time: t, Valid: true}
		}
	}

	if err := models.SaveVHFRiskFactors(db, riskFactors); err != nil {
		sl.Error("Failed to save risk factors", "error", err)
		return c.Status(500).SendString("Failed to save risk factors")
	}

	// Parse and save laboratory data
	laboratory := &models.VHFLaboratory{
		PatientID:            patient.ID,
		SampleCollectionDate: sql.NullTime{Time: parseDate(c.FormValue("sample_collection_date")), Valid: true},
		SampleCollectionTime: sql.NullString{String: c.FormValue("sample_collection_time"), Valid: true},
		SampleType:           c.FormValue("sample_type"),
		OtherSampleType:      c.FormValue("other_sample_type"),
		CreatedAt:            time.Now(),
	}

	if err := models.SaveVHFLaboratory(db, laboratory); err != nil {
		sl.Error("Failed to save laboratory data", "error", err)
		return c.Status(500).SendString("Failed to save laboratory data")
	}

	// Parse and save investigator data
	investigator := &models.VHFInvestigator{
		PatientID:         patient.ID,
		InvestigatorName:  c.FormValue("investigator_name"),
		Phone:             c.FormValue("phone"),
		Email:             c.FormValue("email"),
		Position:          c.FormValue("position"),
		District:          c.FormValue("district"),
		HealthFacility:    c.FormValue("health_facility"),
		InformationSource: c.FormValue("information_source"),
		ProxyName:         c.FormValue("proxy_name"),
		ProxyRelation:     c.FormValue("proxy_relation"),
		CreatedAt:         time.Now(),
	}

	if err := models.SaveVHFInvestigator(db, investigator); err != nil {
		sl.Error("Failed to save investigator data", "error", err)
		return c.Status(500).SendString("Failed to save investigator data")
	}

	// Send SMS notification if phone number is provided
	if patient.DataCapturerPhone != "" {
		message := fmt.Sprintf("VHF Case %s has been successfully registered. Patient: %s %s",
			patient.CaseCode, patient.Surname, patient.OtherNames)
		if err := smsService.SendSMS(patient.DataCapturerPhone, message); err != nil {
			sl.Error("Failed to send SMS notification", "error", err)
			// Don't return error here, as the form was still saved successfully
		}
	}

	// Redirect to success page with case code
	return c.Redirect(fmt.Sprintf("/vhf-cif/success?case_code=%s", patient.CaseCode))
}

// Helper functions for parsing form values
func parseDate(dateStr string) time.Time {
	if dateStr == "" {
		return time.Time{}
	}
	t, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return time.Time{}
	}
	return t
}

func parseInt32(str string) int32 {
	if str == "" {
		return 0
	}
	i, err := strconv.ParseInt(str, 10, 32)
	if err != nil {
		return 0
	}
	return int32(i)
}

func parseFloat(str string) float64 {
	if str == "" {
		return 0
	}
	f, err := strconv.ParseFloat(str, 64)
	if err != nil {
		return 0
	}
	return f
}

func parseBool(str string) bool {
	return str == "Yes"
}

// HandlerVHFClinicalSignsSubmit handles the submission of clinical signs
func HandlerVHFClinicalSignsSubmit(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	patientID, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(400).SendString("Invalid patient ID")
	}

	signs := &models.VHFClinicalSigns{
		PatientID:  patientID,
		TempSource: c.FormValue("temp_source"),
	}

	// Parse date fields
	if onset := c.FormValue("date_initial_onset"); onset != "" {
		if t, err := time.Parse("2006-01-02", onset); err == nil {
			signs.DateInitialOnset = sql.NullTime{Time: t, Valid: true}
		}
	}

	// Parse boolean fields
	signs.Fever = sql.NullBool{Bool: c.FormValue("fever") == "Yes", Valid: true}
	signs.Vomiting = sql.NullBool{Bool: c.FormValue("vomiting") == "Yes", Valid: true}
	signs.Nausea = sql.NullBool{Bool: c.FormValue("nausea") == "Yes", Valid: true}
	signs.Diarrhea = sql.NullBool{Bool: c.FormValue("diarrhea") == "Yes", Valid: true}
	signs.IntenseFatigueGeneralWeakness = sql.NullBool{Bool: c.FormValue("intense_fatigue_general_weakness") == "Yes", Valid: true}
	signs.EpigastricPain = sql.NullBool{Bool: c.FormValue("epigastric_pain") == "Yes", Valid: true}
	signs.LowerAbdominalPain = sql.NullBool{Bool: c.FormValue("lower_abdominal_pain") == "Yes", Valid: true}
	signs.ChestPain = sql.NullBool{Bool: c.FormValue("chest_pain") == "Yes", Valid: true}
	signs.MusclePain = sql.NullBool{Bool: c.FormValue("muscle_pain") == "Yes", Valid: true}
	signs.JointPain = sql.NullBool{Bool: c.FormValue("joint_pain") == "Yes", Valid: true}
	signs.Headache = sql.NullBool{Bool: c.FormValue("headache") == "Yes", Valid: true}
	signs.Cough = sql.NullBool{Bool: c.FormValue("cough") == "Yes", Valid: true}
	signs.DifficultyBreathing = sql.NullBool{Bool: c.FormValue("difficulty_breathing") == "Yes", Valid: true}
	signs.DifficultySwallowing = sql.NullBool{Bool: c.FormValue("difficulty_swallowing") == "Yes", Valid: true}
	signs.SoreThroat = sql.NullBool{Bool: c.FormValue("sore_throat") == "Yes", Valid: true}
	signs.OtherHemorrhagicSymptoms = sql.NullBool{Bool: c.FormValue("other_hemorrhagic_symptoms") == "Yes", Valid: true}

	// Parse temperature
	if temp := c.FormValue("temperature"); temp != "" {
		if t, err := strconv.ParseFloat(temp, 64); err == nil {
			signs.Temperature = sql.NullFloat64{Float64: t, Valid: true}
		}
	}

	// Save to database
	if err := models.SaveVHFClinicalSigns(db, signs); err != nil {
		sl.Error("Failed to save clinical signs", "error", err)
		return c.Status(500).SendString("Failed to save clinical signs")
	}

	// Store patient ID in session
	sess, err := store.Get(c)
	if err != nil {
		sl.Error("Failed to get session", "error", err)
		return c.Status(500).SendString("Failed to get session")
	}
	sess.Set("patient_id", patientID)
	if err := sess.Save(); err != nil {
		sl.Error("Failed to save session", "error", err)
		return c.Status(500).SendString("Failed to save session")
	}

	// Redirect to hospitalization form
	return c.Redirect(fmt.Sprintf("/vhf-cif/hospitalization/%d", patientID))
}

// HandlerVHFHospitalizationSubmit handles the submission of hospitalization information
func HandlerVHFHospitalizationSubmit(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	patientID, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(400).SendString("Invalid patient ID")
	}

	hospitalization := &models.VHFHospitalization{
		PatientID:          patientID,
		Hospitalized:       c.FormValue("hospitalized") == "Yes",
		HealthFacilityName: c.FormValue("health_facility_name"),
		InIsolation:        c.FormValue("isolation") == "Yes",
	}

	// Parse date fields
	if admission := c.FormValue("admission_date"); admission != "" {
		if t, err := time.Parse("2006-01-02", admission); err == nil {
			hospitalization.AdmissionDate = sql.NullTime{Time: t, Valid: true}
		}
	}

	if isolation := c.FormValue("isolation_date"); isolation != "" {
		if t, err := time.Parse("2006-01-02", isolation); err == nil {
			hospitalization.IsolationDate = sql.NullTime{Time: t, Valid: true}
		}
	}

	// Save hospitalization data
	if err := models.SaveVHFHospitalization(db, hospitalization); err != nil {
		sl.Error("Failed to save hospitalization data", "error", err)
		return c.Status(500).SendString("Failed to save hospitalization data")
	}

	// Store patient ID in session
	sess, err := store.Get(c)
	if err != nil {
		sl.Error("Failed to get session", "error", err)
		return c.Status(500).SendString("Failed to get session")
	}
	sess.Set("patient_id", patientID)
	if err := sess.Save(); err != nil {
		sl.Error("Failed to save session", "error", err)
		return c.Status(500).SendString("Failed to save session")
	}

	// Redirect to risk factors form
	return c.Redirect(fmt.Sprintf("/vhf-cif/risk-factors/%d", patientID))
}

// HandlerVHFRiskFactorsSubmit handles the submission of risk factors
func HandlerVHFRiskFactorsSubmit(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	patientID, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(400).SendString("Invalid patient ID")
	}

	riskFactors := &models.VHFRiskFactors{
		PatientID:       patientID,
		ContactWithCase: sql.NullBool{Bool: parseBool(c.FormValue("contactWithCase")), Valid: true},
		ContactName:     c.FormValue("contact_name"),
		ContactRelation: c.FormValue("contact_relation"),
		ContactDates:    c.FormValue("contact_dates"),
		ContactVillage:  c.FormValue("contact_village"),
		ContactDistrict: c.FormValue("contact_district"),
		ContactStatus:   c.FormValue("contact_status"),
		ContactTypes:    c.FormValue("contact_types"),
		CreatedAt:       time.Now(),
	}

	// Parse contact death date if provided
	if deathDate := c.FormValue("contact_death_date"); deathDate != "" {
		if t, err := time.Parse("2006-01-02", deathDate); err == nil {
			riskFactors.ContactDeathDate = sql.NullTime{Time: t, Valid: true}
		}
	}

	// Save risk factors data
	if err := models.SaveVHFRiskFactors(db, riskFactors); err != nil {
		sl.Error("Failed to save risk factors", "error", err)
		return c.Status(500).SendString("Failed to save risk factors")
	}

	// Store patient ID in session
	sess, err := store.Get(c)
	if err != nil {
		sl.Error("Failed to get session", "error", err)
		return c.Status(500).SendString("Failed to get session")
	}
	sess.Set("patient_id", patientID)
	if err := sess.Save(); err != nil {
		sl.Error("Failed to save session", "error", err)
		return c.Status(500).SendString("Failed to save session")
	}

	// Redirect to laboratory form
	return c.Redirect(fmt.Sprintf("/vhf-cif/laboratory/%d", patientID))
}

// HandlerVHFLaboratorySubmit handles the submission of laboratory information
func HandlerVHFLaboratorySubmit(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	patientID, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(400).SendString("Invalid patient ID")
	}

	laboratory := &models.VHFLaboratory{
		PatientID:       patientID,
		SampleType:      c.FormValue("sample_type"),
		OtherSampleType: c.FormValue("other_sample_type"),
		RequestedTest:   c.FormValue("requested_test"),
		Serology:        c.FormValue("serology"),
		MalariaRDT:      c.FormValue("malaria_rdt"),
		HIVRDT:          c.FormValue("hiv_rdt"),
	}

	// Parse date and time fields
	if collectionDate := c.FormValue("sample_collection_date"); collectionDate != "" {
		if t, err := time.Parse("2006-01-02", collectionDate); err == nil {
			laboratory.SampleCollectionDate = sql.NullTime{Time: t, Valid: true}
		}
	}

	if collectionTime := c.FormValue("sample_collection_time"); collectionTime != "" {
		laboratory.SampleCollectionTime = sql.NullString{String: collectionTime, Valid: true}
	}

	// Save laboratory data
	if err := models.SaveVHFLaboratory(db, laboratory); err != nil {
		sl.Error("Failed to save laboratory data", "error", err)
		return c.Status(500).SendString("Failed to save laboratory data")
	}

	// Store patient ID in session
	sess, err := store.Get(c)
	if err != nil {
		sl.Error("Failed to get session", "error", err)
		return c.Status(500).SendString("Failed to get session")
	}
	sess.Set("patient_id", patientID)
	if err := sess.Save(); err != nil {
		sl.Error("Failed to save session", "error", err)
		return c.Status(500).SendString("Failed to save session")
	}

	// Redirect to investigator form
	return c.Redirect(fmt.Sprintf("/vhf-cif/investigator/%d", patientID))
}

// HandlerVHFInvestigatorSubmit handles the submission of investigator information
func HandlerVHFInvestigatorSubmit(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	patientID, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(400).SendString("Invalid patient ID")
	}

	investigator := &models.VHFInvestigator{
		PatientID:         patientID,
		InvestigatorName:  c.FormValue("investigator_name"),
		Phone:             c.FormValue("phone"),
		Email:             c.FormValue("email"),
		Position:          c.FormValue("position"),
		District:          c.FormValue("district"),
		HealthFacility:    c.FormValue("health_facility"),
		InformationSource: c.FormValue("information_source"),
		ProxyName:         c.FormValue("proxy_name"),
		ProxyRelation:     c.FormValue("proxy_relation"),
	}

	if err := models.SaveVHFInvestigator(db, investigator); err != nil {
		sl.Error("Failed to save investigator information", "error", err)
		return c.Status(500).SendString("Failed to save investigator information")
	}

	// Get the case code from the patient record
	patient, err := models.GetVHFPatient(db, patientID)
	if err != nil {
		sl.Error("Failed to get patient", "error", err)
		return c.Status(500).SendString("Failed to get patient information")
	}

	// Redirect to success page with case code
	return c.Redirect(fmt.Sprintf("/vhf-cif/success?case_code=%s", patient.CaseCode))
}

// HandlerVHFList handles the listing of all VHF cases
func HandlerVHFList(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	patients, err := models.ListVHFPatients(db)
	if err != nil {
		sl.Error("Failed to list patients", "error", err)
		return c.Status(500).SendString("Failed to retrieve patient list")
	}

	return GenerateHTML(c, db, fiber.Map{
		"Patients": patients,
	}, "vhf_list")
}

// HandlerVHFView handles viewing a single VHF case
func HandlerVHFView(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(400).SendString("Invalid patient ID")
	}

	patient, err := models.GetVHFPatient(db, id)
	if err != nil {
		sl.Error("Failed to get patient", "error", err)
		return c.Status(500).SendString("Failed to retrieve patient information")
	}

	// TODO: Get all related information (clinical signs, hospitalization, etc.)

	return GenerateHTML(c, db, patient, "vhf_view")
}

// HandlerVHFSuccess handles the success page after form submission
func HandlerVHFSuccess(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	// Get the case code from the query parameters
	caseCode := c.Query("case_code")
	if caseCode == "" {
		sl.Error("No case code provided in success page")
		return c.Status(400).SendString("No case code provided")
	}

	return GenerateHTML(c, db, fiber.Map{
		"CaseCode": caseCode,
	}, "vhf_success")
}
