package handlers

import (
	"case/internal/config"
	"case/internal/models"
	"database/sql"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/lib/pq"
)

func getFormValues(c *fiber.Ctx, key string) []string {
	args := c.Request().PostArgs()
	var vals []string
	args.VisitAll(func(k, v []byte) {
		if string(k) == key {
			vals = append(vals, string(v))
		}
	})
	return vals
}

// Helper function to get array values from form
func getArrayValues(c *fiber.Ctx, key string) pq.StringArray {
	values := c.FormValue(key)
	if values == "" {
		return nil
	}
	return pq.StringArray(strings.Split(values, ","))
}

// HandlerMpoxCIFSubmit handles the submission of the MPOX Case Investigation Form
func HandlerMpoxCIFSubmit(c *fiber.Ctx, db *sql.DB, logger *slog.Logger) error {
	// Start a transaction
	tx, err := db.Begin()
	if err != nil {
		logger.Error("Failed to start transaction", "error", err)
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to start database transaction",
		})
	}
	defer tx.Rollback()

	// Create case investigation record
	caseInvestigation := &models.MpoxCaseInvestigation{
		CaseID: c.FormValue("case_id"),
		Date:   time.Now(),
		CaseStatus: sql.NullString{
			String: c.FormValue("case_status"),
			Valid:  true,
		},
		CaseClassification: sql.NullString{
			String: c.FormValue("case_classification"),
			Valid:  true,
		},
	}

	// Log the case ID for debugging
	logger.Info("Processing MPOX case investigation", "case_id", caseInvestigation.CaseID)

	// Insert the case investigation
	if err := caseInvestigation.Insert(tx); err != nil {
		// Check if it's a duplicate key error
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			logger.Error("Case ID already exists", "case_id", caseInvestigation.CaseID)
			return c.Status(400).JSON(fiber.Map{
				"error": "A case with this ID already exists. Please use a different case ID.",
			})
		}
		logger.Error("Failed to insert case investigation", "error", err, "case_id", caseInvestigation.CaseID)
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to save case investigation",
		})
	}

	// Create patient demographics record
	patientDemographics := &models.MpoxPatientDemographics{
		CaseID: caseInvestigation.CaseID,
		HealthFacilityCaseID: sql.NullString{
			String: c.FormValue("health_facility_case_id"),
			Valid:  true,
		},
		Surname: c.FormValue("surname"),
		OtherNames: sql.NullString{
			String: c.FormValue("other_names"),
			Valid:  true,
		},
		Sex: c.FormValue("sex"),
	}

	// Parse date of birth
	if dob, err := time.Parse("2006-01-02", c.FormValue("date_of_birth")); err == nil {
		patientDemographics.DateOfBirth = dob
	}

	// Parse age
	if age, err := strconv.Atoi(c.FormValue("age")); err == nil {
		patientDemographics.Age = age
	}

	patientDemographics.Parish = sql.NullString{String: c.FormValue("parish"), Valid: true}
	patientDemographics.SubCounty = sql.NullString{String: c.FormValue("sub_county"), Valid: true}
	patientDemographics.PhysicalAddress = sql.NullString{String: c.FormValue("physical_address"), Valid: true}
	patientDemographics.ContactTelephone = sql.NullString{String: c.FormValue("contact_telephone"), Valid: true}
	patientDemographics.Occupation = sql.NullString{String: c.FormValue("occupation"), Valid: true}
	patientDemographics.Nationality = sql.NullString{String: c.FormValue("nationality"), Valid: true}
	patientDemographics.VaccinationStatus = sql.NullString{String: c.FormValue("vaccination_status"), Valid: true}

	if dov, err := time.Parse("2006-01-02", c.FormValue("date_of_vaccination")); err == nil {
		patientDemographics.DateOfVaccination = sql.NullTime{Time: dov, Valid: true}
	}

	patientDemographics.NextOfKin = sql.NullString{String: c.FormValue("next_of_kin"), Valid: true}
	patientDemographics.NextOfKinContact = sql.NullString{String: c.FormValue("next_of_kin_contact"), Valid: true}
	patientDemographics.MaritalStatus = sql.NullString{String: c.FormValue("marital_status"), Valid: true}

	if dod, err := time.Parse("2006-01-02", c.FormValue("if_dead_date_of_death")); err == nil {
		patientDemographics.IfDeadDateOfDeath = sql.NullTime{Time: dod, Valid: true}
	}

	if ad, err := time.Parse("2006-01-02", c.FormValue("admission_date")); err == nil {
		patientDemographics.AdmissionDate = sql.NullTime{Time: ad, Valid: true}
	}

	if od, err := time.Parse("2006-01-02", c.FormValue("onset_date")); err == nil {
		patientDemographics.OnsetDate = sql.NullTime{Time: od, Valid: true}
	}

	if rod, err := time.Parse("2006-01-02", c.FormValue("rash_onset_date")); err == nil {
		patientDemographics.RashOnsetDate = sql.NullTime{Time: rod, Valid: true}
	}

	// Log patient demographics for debugging
	logger.Info("Inserting patient demographics",
		"case_id", caseInvestigation.CaseID,
		"surname", patientDemographics.Surname,
		"other_names", patientDemographics.OtherNames.String)

	if err := patientDemographics.Insert(tx); err != nil {
		logger.Error("Failed to insert patient demographics", "error", err, "case_id", caseInvestigation.CaseID)
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to save case investigation",
		})
	}

	// Create clinician info record
	clinicianInfo := &models.MpoxClinicianInfo{
		CaseID: caseInvestigation.CaseID,
		ClinicianName: sql.NullString{
			String: c.FormValue("clinician_name"),
			Valid:  true,
		},
		ClinicianContact: sql.NullString{
			String: c.FormValue("clinician_contact"),
			Valid:  true,
		},
		FacilityName: sql.NullString{
			String: c.FormValue("facility_name"),
			Valid:  true,
		},
		ClinicianEmail: sql.NullString{
			String: c.FormValue("clinician_email"),
			Valid:  true,
		},
		FacilityDistrict: sql.NullString{
			String: c.FormValue("facility_district"),
			Valid:  true,
		},
		PDPIDNumber: sql.NullString{
			String: c.FormValue("pdpid_number"),
			Valid:  true,
		},
	}

	if ad, err := time.Parse("2006-01-02", c.FormValue("admission_date")); err == nil {
		clinicianInfo.AdmissionDate = sql.NullTime{Time: ad, Valid: true}
	}

	clinicianInfo.Ward = sql.NullString{String: c.FormValue("ward"), Valid: true}

	// Log clinician info for debugging
	logger.Info("Inserting clinician info",
		"case_id", caseInvestigation.CaseID,
		"clinician_name", clinicianInfo.ClinicianName.String,
		"facility_name", clinicianInfo.FacilityName.String)

	if err := clinicianInfo.Insert(tx); err != nil {
		logger.Error("Failed to insert clinician info", "error", err, "case_id", caseInvestigation.CaseID)
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to save case investigation",
		})
	}

	// Create exposure history record
	exposureHistory := &models.MpoxCaseExposureHistory{
		CaseID: caseInvestigation.CaseID,
		TraveledCountryReportedMpox: sql.NullString{
			String: c.FormValue("traveled_country_reported_mpox"),
			Valid:  true,
		},
		CloseContactMpox: sql.NullString{
			String: c.FormValue("close_contact_mpox"),
			Valid:  true,
		},
		IntlTravel: sql.NullString{
			String: c.FormValue("intl_travel"),
			Valid:  true,
		},
		ContactAnimals: sql.NullString{
			String: c.FormValue("contact_animals"),
			Valid:  true,
		},
		DomesticWildAnimals: sql.NullString{
			String: c.FormValue("domestic_wild_animals"),
			Valid:  true,
		},
		SexualExposure: sql.NullString{
			String: c.FormValue("sexual_exposure"),
			Valid:  true,
		},
	}

	// Log exposure history for debugging
	logger.Info("Inserting exposure history",
		"case_id", caseInvestigation.CaseID,
		"traveled_country", exposureHistory.TraveledCountryReportedMpox.String,
		"close_contact", exposureHistory.CloseContactMpox.String)

	if err := exposureHistory.Insert(tx); err != nil {
		logger.Error("Failed to insert exposure history", "error", err, "case_id", caseInvestigation.CaseID)
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to save case investigation",
		})
	}

	// Create clinical manifestations record
	clinicalManifestations := &models.MpoxClinicalManifestations{
		CaseID: caseInvestigation.CaseID,
	}

	if od, err := time.Parse("2006-01-02", c.FormValue("onset_date")); err == nil {
		clinicalManifestations.OnsetDate = sql.NullTime{Time: od, Valid: true}
	}

	clinicalManifestations.Fever = sql.NullString{String: c.FormValue("fever"), Valid: true}
	clinicalManifestations.FeverTemperature = sql.NullString{String: c.FormValue("fever_temperature"), Valid: true}
	clinicalManifestations.Lymphadenopathy = sql.NullString{String: c.FormValue("lymphadenopathy"), Valid: true}
	clinicalManifestations.Symptoms = getArrayValues(c, "symptoms")
	clinicalManifestations.SymptomOtherSpecify = sql.NullString{String: c.FormValue("symptom_other_specify"), Valid: true}
	clinicalManifestations.NauseaVomiting = sql.NullString{String: c.FormValue("nausea_vomiting"), Valid: true}
	clinicalManifestations.Pregnant = sql.NullString{String: c.FormValue("pregnant"), Valid: true}
	clinicalManifestations.PregnantTrimester = sql.NullString{String: c.FormValue("pregnant_trimester"), Valid: true}
	clinicalManifestations.Vaccinated = sql.NullString{String: c.FormValue("vaccinated"), Valid: true}
	clinicalManifestations.VaccinationDate = sql.NullString{String: c.FormValue("vaccination_date"), Valid: true}
	clinicalManifestations.Rash = sql.NullString{String: c.FormValue("rash"), Valid: true}

	if rod, err := time.Parse("2006-01-02", c.FormValue("rash_onset_date")); err == nil {
		clinicalManifestations.RashOnsetDate = sql.NullTime{Time: rod, Valid: true}
	}

	clinicalManifestations.RashDistribution = getArrayValues(c, "rash_distribution")
	clinicalManifestations.RashType = getArrayValues(c, "rash_type")
	clinicalManifestations.UnderlyingIllness = sql.NullString{String: c.FormValue("underlying_illness"), Valid: true}
	clinicalManifestations.UnderlyingIllnessDetails = sql.NullString{String: c.FormValue("underlying_illness_details"), Valid: true}

	// Log clinical manifestations for debugging
	logger.Info("Inserting clinical manifestations",
		"case_id", caseInvestigation.CaseID,
		"fever", clinicalManifestations.Fever.String,
		"rash", clinicalManifestations.Rash.String)

	if err := clinicalManifestations.Insert(tx); err != nil {
		logger.Error("Failed to insert clinical manifestations", "error", err, "case_id", caseInvestigation.CaseID)
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to save case investigation",
		})
	}

	// Create travel history record
	travelHistory := &models.MpoxTravelHistory{
		CaseID: caseInvestigation.CaseID,
		TravelOutsideUganda: sql.NullString{
			String: c.FormValue("travel_outside_uganda"),
			Valid:  true,
		},
		CountryVisited:     getArrayValues(c, "country_visited"),
		LocationVisited:    getArrayValues(c, "location_visited"),
		DateArrival:        getArrayValues(c, "date_arrival"),
		DateDeparture:      getArrayValues(c, "date_departure"),
		ActivitiesLocation: getArrayValues(c, "activities_location"),
	}

	// Log travel history for debugging
	logger.Info("Inserting travel history",
		"case_id", caseInvestigation.CaseID,
		"travel_outside", travelHistory.TravelOutsideUganda.String,
		"countries", travelHistory.CountryVisited)

	if err := travelHistory.Insert(tx); err != nil {
		logger.Error("Failed to insert travel history", "error", err, "case_id", caseInvestigation.CaseID)
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to save case investigation",
		})
	}

	// Create lab investigation record
	labInvestigation := &models.MpoxLabInvestigation{
		CaseID: caseInvestigation.CaseID,
		LabID: sql.NullString{
			String: c.FormValue("lab_id"),
			Valid:  true,
		},
		SampleCollected:    getArrayValues(c, "sample_collected"),
		SampleOtherSpecify: sql.NullString{String: c.FormValue("sample_other_specify"), Valid: true},
		TestRequested:      getArrayValues(c, "test_requested"),
		TestOtherSpecify:   sql.NullString{String: c.FormValue("test_other_specify"), Valid: true},
	}

	if dsc, err := time.Parse("2006-01-02", c.FormValue("date_sample_collection")); err == nil {
		labInvestigation.DateSampleCollection = sql.NullTime{Time: dsc, Valid: true}
	}

	if tsc, err := time.Parse("15:04", c.FormValue("time_sample_collection")); err == nil {
		labInvestigation.TimeSampleCollection = sql.NullTime{Time: tsc, Valid: true}
	}

	if dsd, err := time.Parse("2006-01-02", c.FormValue("date_sample_dispatch")); err == nil {
		labInvestigation.DateSampleDispatch = sql.NullTime{Time: dsd, Valid: true}
	}

	labInvestigation.SampleCollectorName = sql.NullString{String: c.FormValue("sample_collector_name"), Valid: true}
	labInvestigation.SampleCollectorPhone = sql.NullString{String: c.FormValue("sample_collector_phone"), Valid: true}

	if dsr, err := time.Parse("2006-01-02", c.FormValue("date_sample_reception")); err == nil {
		labInvestigation.DateSampleReception = sql.NullTime{Time: dsr, Valid: true}
	}

	if tsr, err := time.Parse("15:04", c.FormValue("time_sample_reception")); err == nil {
		labInvestigation.TimeSampleReception = sql.NullTime{Time: tsr, Valid: true}
	}

	labInvestigation.SampleRecipientName = sql.NullString{String: c.FormValue("sample_recipient_name"), Valid: true}
	labInvestigation.SampleRecipientPhone = sql.NullString{String: c.FormValue("sample_recipient_phone"), Valid: true}
	labInvestigation.GenomicCharacterization = sql.NullString{String: c.FormValue("genomic_characterization"), Valid: true}
	labInvestigation.Clade = getArrayValues(c, "clade")
	labInvestigation.AccessionNumber = sql.NullString{String: c.FormValue("accession_number"), Valid: true}

	// Log lab investigation for debugging
	logger.Info("Inserting lab investigation",
		"case_id", caseInvestigation.CaseID,
		"lab_id", labInvestigation.LabID.String,
		"sample_collected", labInvestigation.SampleCollected)

	if err := labInvestigation.Insert(tx); err != nil {
		logger.Error("Failed to insert lab investigation", "error", err, "case_id", caseInvestigation.CaseID)
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to save case investigation",
		})
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		logger.Error("Failed to commit transaction", "error", err, "case_id", caseInvestigation.CaseID)
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to save case investigation",
		})
	}

	// Log successful submission
	logger.Info("Successfully saved MPOX case investigation", "case_id", caseInvestigation.CaseID)

	// Redirect to success page
	return c.Redirect("/mpox-cif/success?case_id=" + caseInvestigation.CaseID)
}

// HandlerMpoxCIFView handles viewing an mpox CIF
func HandlerMpoxCIFView(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config config.Config) error {
	idParam := c.Params("id")
	if idParam == "" {
		return c.Status(400).SendString("Case ID required")
	}

	// Determine if this is a numeric ID (primary key) or case_id (string)
	// Try numeric first (database id), then fall back to case_id
	var ci *models.MpoxCaseInvestigation
	var caseID string

	ci = &models.MpoxCaseInvestigation{}
	// Try querying by numeric ID first
	if idInt, err := strconv.Atoi(idParam); err == nil {
		// It's a numeric ID - query by primary key
		if err := db.QueryRow(`SELECT id, case_id, date, case_status, case_classification FROM mpox_case_investigation WHERE id = $1`, idInt).
			Scan(&ci.ID, &ci.CaseID, &ci.Date, &ci.CaseStatus, &ci.CaseClassification); err != nil {
			if err == sql.ErrNoRows {
				return c.Status(404).SendString("CIF not found")
			}
			sl.Error("Failed to load case investigation", "error", err, "id", idParam)
			return c.Status(500).SendString("Failed to load case")
		}
		caseID = ci.CaseID
	} else {
		// It's a string - query by case_id
		caseID = idParam
		if err := db.QueryRow(`SELECT id, case_id, date, case_status, case_classification FROM mpox_case_investigation WHERE case_id = $1`, caseID).
			Scan(&ci.ID, &ci.CaseID, &ci.Date, &ci.CaseStatus, &ci.CaseClassification); err != nil {
			if err == sql.ErrNoRows {
				return c.Status(404).SendString("CIF not found")
			}
			sl.Error("Failed to load case investigation", "error", err, "case_id", caseID)
			return c.Status(500).SendString("Failed to load case")
		}
	}

	// Now use caseID to fetch related data
	var demo *models.MpoxPatientDemographics
	var clin *models.MpoxClinicianInfo
	var exp *models.MpoxCaseExposureHistory
	var man *models.MpoxClinicalManifestations
	var travel *models.MpoxTravelHistory
	var lab *models.MpoxLabInvestigation

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

	data := NewTemplateData(c, store)
	data.Form = fiber.Map{
		"Case":           ci,
		"Demographics":   demo,
		"Clinician":      clin,
		"Exposure":       exp,
		"Manifestations": man,
		"TravelHistory":  travel,
		"Laboratory":     lab,
		"IsView":         true,
	}
	return GenerateHTML(c, db, data, "mpox_cif")
}

// HandlerMpoxCIFEdit handles editing an mpox CIF
func HandlerMpoxCIFEdit(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config config.Config) error {
	idParam := c.Params("id")
	if idParam == "" {
		return c.Status(400).SendString("Case ID required")
	}

	// Determine if this is a numeric ID (primary key) or case_id (string)
	var ci *models.MpoxCaseInvestigation
	var caseID string

	ci = &models.MpoxCaseInvestigation{}
	// Try querying by numeric ID first
	if idInt, err := strconv.Atoi(idParam); err == nil {
		// It's a numeric ID - query by primary key
		if err := db.QueryRow(`SELECT id, case_id, date, case_status, case_classification FROM mpox_case_investigation WHERE id = $1`, idInt).
			Scan(&ci.ID, &ci.CaseID, &ci.Date, &ci.CaseStatus, &ci.CaseClassification); err != nil {
			if err == sql.ErrNoRows {
				return c.Status(404).SendString("CIF not found")
			}
			sl.Error("Failed to load case investigation", "error", err, "id", idParam)
			return c.Status(500).SendString("Failed to load case")
		}
		caseID = ci.CaseID
	} else {
		// It's a string - query by case_id
		caseID = idParam
		if err := db.QueryRow(`SELECT id, case_id, date, case_status, case_classification FROM mpox_case_investigation WHERE case_id = $1`, caseID).
			Scan(&ci.ID, &ci.CaseID, &ci.Date, &ci.CaseStatus, &ci.CaseClassification); err != nil {
			if err == sql.ErrNoRows {
				return c.Status(404).SendString("CIF not found")
			}
			sl.Error("Failed to load case investigation", "error", err, "case_id", caseID)
			return c.Status(500).SendString("Failed to load case")
		}
	}

	// Now use caseID to fetch related data
	var demo *models.MpoxPatientDemographics
	var clin *models.MpoxClinicianInfo
	var exp *models.MpoxCaseExposureHistory
	var man *models.MpoxClinicalManifestations
	var travel *models.MpoxTravelHistory
	var lab *models.MpoxLabInvestigation

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

	data := NewTemplateData(c, store)
	data.Form = fiber.Map{
		"Case":           ci,
		"Demographics":   demo,
		"Clinician":      clin,
		"Exposure":       exp,
		"Manifestations": man,
		"TravelHistory":  travel,
		"Laboratory":     lab,
		"IsEdit":         true,
		"case_id":        caseID,
	}
	return GenerateHTML(c, db, data, "mpox_cif")
}

func HandlerMpoxCIFSuccess(c *fiber.Ctx, db *sql.DB, logger *slog.Logger, store *session.Store) error {
	caseID := c.Query("case_id")
	data := NewTemplateData(c, store)
	data.Form = fiber.Map{"case_id": caseID}
	return GenerateHTML(c, db, data, "mpox_cif_success")
}

// HandlerMpoxList handles the listing of all MPOX cases
func HandlerMpoxList(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config config.Config) error {
	// Get current user information (for future use if needed)
	_, _ = GetUser(c, sl, store)

	// Build the query to get MPOX cases
	query := `
		SELECT 
			mci.id,
			mci.case_id,
			mci.case_status,
			mci.case_classification,
			mci.date,
			mpd.surname,
			mpd.other_names,
			mpd.sex,
			mpd.age,
			mpd.parish,
			mpd.sub_county,
			mpd.onset_date,
			mpd.rash_onset_date,
			CASE WHEN mli.id IS NOT NULL THEN true ELSE false END as lab_status
		FROM mpox_case_investigation mci
		LEFT JOIN mpox_patient_demographics mpd ON mci.case_id = mpd.case_id
		LEFT JOIN mpox_lab_investigation mli ON mci.case_id = mli.case_id
		ORDER BY mci.date DESC`

	rows, err := db.QueryContext(c.Context(), query)
	if err != nil {
		sl.Error("Failed to query MPOX cases", "error", err)
		return c.Status(500).SendString("Failed to retrieve MPOX cases")
	}
	defer rows.Close()

	var cases []fiber.Map
	for rows.Next() {
		var (
			id                 int64
			caseID             string
			caseStatus         sql.NullString
			caseClassification sql.NullString
			date               time.Time
			surname            string
			otherNames         sql.NullString
			sex                string
			age                int
			parish             sql.NullString
			subCounty          sql.NullString
			onsetDate          sql.NullTime
			rashOnsetDate      sql.NullTime
			labStatus          bool
		)

		err := rows.Scan(
			&id,
			&caseID,
			&caseStatus,
			&caseClassification,
			&date,
			&surname,
			&otherNames,
			&sex,
			&age,
			&parish,
			&subCounty,
			&onsetDate,
			&rashOnsetDate,
			&labStatus,
		)
		if err != nil {
			sl.Error("Failed to scan MPOX case", "error", err)
			continue
		}

		// Format name
		name := surname
		if otherNames.Valid && otherNames.String != "" {
			name = fmt.Sprintf("%s %s", surname, otherNames.String)
		}

		// Format location
		location := ""
		if parish.Valid && parish.String != "" {
			location = parish.String
		}
		if subCounty.Valid && subCounty.String != "" {
			if location != "" {
				location += ", "
			}
			location += subCounty.String
		}

		// Determine status
		status := "Active"
		if caseStatus.Valid && caseStatus.String != "" {
			status = caseStatus.String
		}

		cases = append(cases, fiber.Map{
			"ID":             id,
			"CaseID":         caseID,
			"Name":           name,
			"Age":            fmt.Sprintf("%d years", age),
			"Gender":         sex,
			"Location":       location,
			"Status":         status,
			"Classification": caseClassification.String,
			"LabStatus":      labStatus,
			"OnsetDate":      onsetDate.Time.Format("2006-01-02"),
			"RashOnsetDate":  rashOnsetDate.Time.Format("2006-01-02"),
			"CreatedAt":      date.Format("2006-01-02 15:04"),
		})
	}

	if err = rows.Err(); err != nil {
		sl.Error("Error iterating MPOX cases", "error", err)
		return c.Status(500).SendString("Error retrieving MPOX cases")
	}

	// Check if this is an API request
	if c.Get("Accept") == "application/json" {
		return c.JSON(cases)
	}

	// Return HTML response
	data := NewTemplateData(c, store)
	data.Form = fiber.Map{
		"Title": "MPOX Cases",
		"Cases": cases,
	}
	return GenerateHTML(c, db, data, "mpox_list")
}
