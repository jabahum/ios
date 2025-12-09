package handlers

import (
	"case/internal/models"
	"database/sql"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
)

// HandlerPolioCIF handles the Polio CIF form display
func HandlerPolioCIF(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	data := NewTemplateData(c, store)
	data.Form = fiber.Map{
		"Title": "Polio Case Investigation Form",
	}
	return GenerateHTML(c, db, data, "polio_cif")
}

// HandlerPolioCIFSubmit handles the submission of the Polio CIF form
func HandlerPolioCIFSubmit(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	// Debug: Log that the handler was called
	sl.Info("Polio CIF submit handler called", "method", c.Method(), "path", c.Path())
	// Start a transaction
	tx, err := db.Begin()
	if err != nil {
		sl.Error("Failed to start transaction", "error", err)
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to start database transaction",
		})
	}
	defer tx.Rollback()

	caseID := c.FormValue("case_id")
	if caseID == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Case ID is required",
		})
	}

	// Create case investigation record
	caseInvestigation := &models.PolioCaseInvestigation{
		CaseID:         caseID,
		EpidNumber:     c.FormValue("epid_number"),
		Country:        c.FormValue("country"),
		RegionProvince: c.FormValue("region_province"),
		District:       c.FormValue("district"),
		YearOnset:      parseInt(c.FormValue("year_onset")),
		CaseNumber:     parseInt(c.FormValue("case_number")),
		ReceivedDate:   time.Now(),
		CreatedAt:      time.Now(),
	}

	// Parse received date if provided
	if receivedDate := c.FormValue("received_date"); receivedDate != "" {
		if t, err := time.Parse("2006-01-02", receivedDate); err == nil {
			caseInvestigation.ReceivedDate = t
		}
	}

	// Insert the case investigation
	if err := caseInvestigation.Insert(db); err != nil {
		sl.Error("Failed to insert case investigation", "error", err, "case_id", caseID)
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to save case investigation",
		})
	}

	// Create identification record
	identification := &models.PolioIdentification{
		CaseID:                caseID,
		District:              c.FormValue("district"),
		RegionProvince:        c.FormValue("region_province"),
		Address:               c.FormValue("address"),
		Village:               c.FormValue("village"),
		City:                  c.FormValue("city"),
		NearestHealthFacility: c.FormValue("nearest_health_facility"),
		PatientName:           c.FormValue("patient_name"),
		FatherMother:          c.FormValue("father_mother"),
		PhoneNumber:           c.FormValue("phone_number"),
		Sex:                   c.FormValue("sex"),
	}

	// Parse date of birth
	if dob := c.FormValue("date_of_birth"); dob != "" {
		if t, err := time.Parse("2006-01-02", dob); err == nil {
			identification.DateOfBirth = sql.NullTime{Time: t, Valid: true}
		}
	}

	// Parse age
	if ageYears := c.FormValue("age_years"); ageYears != "" {
		if age, err := strconv.Atoi(ageYears); err == nil {
			identification.AgeYears = sql.NullInt32{Int32: int32(age), Valid: true}
		}
	}
	if ageMonths := c.FormValue("age_months"); ageMonths != "" {
		if age, err := strconv.Atoi(ageMonths); err == nil {
			identification.AgeMonths = sql.NullInt32{Int32: int32(age), Valid: true}
		}
	}

	// Parse coordinates
	if longitude := c.FormValue("longitude"); longitude != "" {
		if lon, err := strconv.ParseFloat(longitude, 64); err == nil {
			identification.Longitude = sql.NullFloat64{Float64: lon, Valid: true}
		}
	}
	if latitude := c.FormValue("latitude"); latitude != "" {
		if lat, err := strconv.ParseFloat(latitude, 64); err == nil {
			identification.Latitude = sql.NullFloat64{Float64: lat, Valid: true}
		}
	}

	if err := identification.Insert(db); err != nil {
		sl.Error("Failed to insert identification", "error", err)
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to save identification",
		})
	}

	// Create notification/investigation record
	notification := &models.PolioNotificationInvestigation{
		CaseID:     c.FormValue("case_id"),
		NotifiedBy: c.FormValue("notified_by"),
	}

	// Parse notification date
	if notifDate := c.FormValue("date_of_notification"); notifDate != "" {
		if t, err := time.Parse("2006-01-02", notifDate); err == nil {
			notification.DateOfNotification = sql.NullTime{Time: t, Valid: true}
		}
	}

	// Parse investigation date
	if invDate := c.FormValue("date_of_investigation"); invDate != "" {
		if t, err := time.Parse("2006-01-02", invDate); err == nil {
			notification.DateOfInvestigation = sql.NullTime{Time: t, Valid: true}
		}
	}

	if err := notification.Insert(db); err != nil {
		sl.Error("Failed to insert notification", "error", err)
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to save notification",
		})
	}

	// Create hospitalization record
	hospitalization := &models.PolioHospitalization{
		CaseID:               caseID,
		Hospitalized:         c.FormValue("hospitalized") == "on",
		HospitalRecordNumber: c.FormValue("hospital_record_number"),
		HospitalNameAddress:  c.FormValue("hospital_name_address"),
	}

	// Parse admission date
	if admDate := c.FormValue("date_of_admission"); admDate != "" {
		if t, err := time.Parse("2006-01-02", admDate); err == nil {
			hospitalization.DateOfAdmission = sql.NullTime{Time: t, Valid: true}
		}
	}

	if err := hospitalization.Insert(db); err != nil {
		sl.Error("Failed to insert hospitalization", "error", err)
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to save hospitalization",
		})
	}

	// Create clinical history record
	clinical := &models.PolioClinicalHistory{
		CaseID:                 caseID,
		FeverAtOnset:           sql.NullString{String: c.FormValue("fever_at_onset"), Valid: c.FormValue("fever_at_onset") != ""},
		ProgressiveParalysis:   sql.NullString{String: c.FormValue("progressive_paralysis"), Valid: c.FormValue("progressive_paralysis") != ""},
		FlaccidAcuteParalysis:  sql.NullString{String: c.FormValue("flaccid_acute_paralysis"), Valid: c.FormValue("flaccid_acute_paralysis") != ""},
		SensationLoss:          sql.NullString{String: c.FormValue("sensation_loss"), Valid: c.FormValue("sensation_loss") != ""},
		SuddenOnset:            c.FormValue("sudden_onset") == "on",
		Asymmetric:             c.FormValue("asymmetric") == "on",
		LeftArmParalysis:       c.FormValue("left_arm_paralysis") == "on",
		RightArmParalysis:      c.FormValue("right_arm_paralysis") == "on",
		LeftLegParalysis:       c.FormValue("left_leg_paralysis") == "on",
		RightLegParalysis:      c.FormValue("right_leg_paralysis") == "on",
		DiminishedReflexes:     c.FormValue("diminished_reflexes") == "on",
		DiminishedMuscleTone:   c.FormValue("diminished_muscle_tone") == "on",
		MuscleWasting:          c.FormValue("muscle_wasting") == "on",
		MuscleWeakness:         c.FormValue("muscle_weakness") == "on",
		RespiratoryMuscles:     c.FormValue("respiratory_muscles") == "on",
		Face:                   c.FormValue("face") == "on",
		StiffNeck:              c.FormValue("stiff_neck") == "on",
		Convulsions:            c.FormValue("convulsions") == "on",
		Headache:               c.FormValue("headache") == "on",
		Vomiting:               c.FormValue("vomiting") == "on",
		Diarrhoea:              c.FormValue("diarrhoea") == "on",
		OtherSites:             c.FormValue("other_sites"),
		RecentInjection:        sql.NullString{String: c.FormValue("recent_injection"), Valid: c.FormValue("recent_injection") != ""},
		InjectionType:          c.FormValue("injection_type"),
		ParalyzedLimbSensitive: sql.NullString{String: c.FormValue("paralyzed_limb_sensitive"), Valid: c.FormValue("paralyzed_limb_sensitive") != ""},
		InjectionFacilityName:  c.FormValue("injection_facility_name"),
		ProvisionalDiagnosis:   c.FormValue("provisional_diagnosis"),
		TrueAFPCase:            sql.NullString{String: c.FormValue("true_afp_case"), Valid: c.FormValue("true_afp_case") != ""},
	}

	// Parse clinical dates
	if feverDate := c.FormValue("date_onset_of_fever"); feverDate != "" {
		if t, err := time.Parse("2006-01-02", feverDate); err == nil {
			clinical.DateOnsetOfFever = sql.NullTime{Time: t, Valid: true}
		}
	}
	if paralysisDate := c.FormValue("date_onset_of_paralysis"); paralysisDate != "" {
		if t, err := time.Parse("2006-01-02", paralysisDate); err == nil {
			clinical.DateOnsetOfParalysis = sql.NullTime{Time: t, Valid: true}
		}
	}

	// Parse total injections
	if totalInjections := c.FormValue("total_injections"); totalInjections != "" {
		if total, err := strconv.Atoi(totalInjections); err == nil {
			clinical.TotalInjections = sql.NullInt32{Int32: int32(total), Valid: true}
		}
	}

	if err := clinical.Insert(db); err != nil {
		sl.Error("Failed to insert clinical history", "error", err)
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to save clinical history",
		})
	}

	// Create immunization history record
	immunization := &models.PolioImmunizationHistory{
		CaseID:                 caseID,
		ExcludeDoseAtBirth:     c.FormValue("exclude_dose_at_birth") == "on",
		SourceOfRIVaccination:  c.FormValue("source_of_ri_vaccination"),
		UnknownZeroDoseReasons: c.FormValue("unknown_zero_dose_reasons"),
	}

	immunization.OPVDoseAtBirth = ParseNullTime(c.FormValue("opv_dose_at_birth"))
	immunization.OPVDose1 = ParseNullTime(c.FormValue("opv_dose1"))
	immunization.OPVDose2 = ParseNullTime(c.FormValue("opv_dose2"))
	immunization.OPVDose3 = ParseNullTime(c.FormValue("opv_dose3"))
	immunization.OPVDose4 = ParseNullTime(c.FormValue("opv_dose4"))
	immunization.OPVDoseMoreThan4 = ParseNullTime(c.FormValue("opv_dose_more_than4"))
	immunization.LastOPVDose = ParseNullTime(c.FormValue("last_opv_dose"))
	immunization.LastOPVSIA = ParseNullTime(c.FormValue("last_opv_sia"))
	immunization.LastIPVSIA = ParseNullTime(c.FormValue("last_ipv_sia"))

	// Parse numeric fields
	parseInt32 := func(value string) sql.NullInt32 {
		if value == "" {
			return sql.NullInt32{Valid: false}
		}
		if val, err := strconv.Atoi(value); err == nil {
			return sql.NullInt32{Int32: int32(val), Valid: true}
		}
		return sql.NullInt32{Valid: false}
	}

	immunization.TotalPolioDoses = parseInt32(c.FormValue("total_polio_doses"))
	immunization.TotalOPVSIA = parseInt32(c.FormValue("total_opv_sia"))
	immunization.TotalOPVRI = parseInt32(c.FormValue("total_opv_ri"))
	immunization.TotalIPVSIA = parseInt32(c.FormValue("total_ipv_sia"))
	immunization.TotalIPVRI = parseInt32(c.FormValue("total_ipv_ri"))

	if err := immunization.Insert(db); err != nil {
		sl.Error("Failed to insert immunization history", "error", err)
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to save immunization history",
		})
	}

	// Create stool specimen collection record
	specimenCollection := &models.PolioStoolSpecimenCollection{
		CaseID: caseID,
	}

	specimenCollection.DateFirstSpecimen = ParseNullTime(c.FormValue("date_first_specimen"))
	specimenCollection.DateSecondSpecimen = ParseNullTime(c.FormValue("date_second_specimen"))
	specimenCollection.DateSpecimenSentNational = ParseNullTime(c.FormValue("date_specimen_sent_national"))
	specimenCollection.DateSpecimenReceivedNational = ParseNullTime(c.FormValue("date_specimen_received_national"))
	specimenCollection.DateSpecimenSentLab = ParseNullTime(c.FormValue("date_specimen_sent_lab"))

	if err := specimenCollection.Insert(db); err != nil {
		sl.Error("Failed to insert specimen collection", "error", err)
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to save specimen collection",
		})
	}

	// Create stool specimen results record
	specimenResults := &models.PolioStoolSpecimenResults{
		CaseID: caseID,
	}

	specimenResults.DateReceivedAtLab = ParseNullTime(c.FormValue("date_received_at_lab"))
	specimenResults.SpecimenStatusAtReception = sql.NullString{String: c.FormValue("specimen_status_at_reception"), Valid: c.FormValue("specimen_status_at_reception") != ""}
	specimenResults.DateCombinedCellCulture = ParseNullTime(c.FormValue("date_combined_cell_culture"))
	specimenResults.DateResultsSentToEPI = ParseNullTime(c.FormValue("date_results_sent_to_epi"))
	specimenResults.DateResultsReceivedAtEPI = ParseNullTime(c.FormValue("date_results_received_at_epi"))
	specimenResults.FinalCellCultureResults = sql.NullString{String: c.FormValue("final_cell_culture_results"), Valid: c.FormValue("final_cell_culture_results") != ""}
	specimenResults.W1 = sql.NullString{String: c.FormValue("w1"), Valid: c.FormValue("w1") != ""}
	specimenResults.W2 = sql.NullString{String: c.FormValue("w2"), Valid: c.FormValue("w2") != ""}
	specimenResults.W3 = sql.NullString{String: c.FormValue("w3"), Valid: c.FormValue("w3") != ""}
	specimenResults.DiscordantSabin = sql.NullString{String: c.FormValue("discordant_sabin"), Valid: c.FormValue("discordant_sabin") != ""}
	specimenResults.SL1 = sql.NullString{String: c.FormValue("sl1"), Valid: c.FormValue("sl1") != ""}
	specimenResults.SL2 = sql.NullString{String: c.FormValue("sl2"), Valid: c.FormValue("sl2") != ""}
	specimenResults.SL3 = sql.NullString{String: c.FormValue("sl3"), Valid: c.FormValue("sl3") != ""}
	specimenResults.RNPENT = sql.NullString{String: c.FormValue("r_npent"), Valid: c.FormValue("r_npent") != ""}
	specimenResults.NEV = sql.NullString{String: c.FormValue("nev"), Valid: c.FormValue("nev") != ""}
	specimenResults.DateSentToRegionalLab = ParseNullTime(c.FormValue("date_sent_to_regional_lab"))
	specimenResults.DateITDifferentiationSent = ParseNullTime(c.FormValue("date_it_differentiation_sent"))
	specimenResults.DateITDifferentiationReceived = ParseNullTime(c.FormValue("date_it_differentiation_received"))
	specimenResults.DateIsolateSentSequencing = ParseNullTime(c.FormValue("date_isolate_sent_sequencing"))
	specimenResults.DateSeqResultsSentProgram = ParseNullTime(c.FormValue("date_seq_results_sent_program"))

	if err := specimenResults.Insert(db); err != nil {
		sl.Error("Failed to insert specimen results", "error", err)
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to save specimen results",
		})
	}

	// Create follow-up examination record
	followUp := &models.PolioFollowUpExamination{
		CaseID:              caseID,
		ResidualParalysisLA: c.FormValue("residual_paralysis_la") == "on",
		ResidualParalysisRA: c.FormValue("residual_paralysis_ra") == "on",
		ResidualParalysisLL: c.FormValue("residual_paralysis_ll") == "on",
		ResidualParalysisRL: c.FormValue("residual_paralysis_rl") == "on",
		CVDPV:               c.FormValue("cvdpv") == "on",
		AVDPV:               c.FormValue("avdpv") == "on",
		IVDPV:               c.FormValue("ivdpv") == "on",
	}

	followUp.DateOfFollowUp = ParseNullTime(c.FormValue("date_of_follow_up"))
	followUp.ResultsOfExam = sql.NullString{String: c.FormValue("results_of_exam"), Valid: c.FormValue("results_of_exam") != ""}
	followUp.ImmunocompromisedStatus = sql.NullString{String: c.FormValue("immunocompromised_status"), Valid: c.FormValue("immunocompromised_status") != ""}
	followUp.FinalClassification = sql.NullString{String: c.FormValue("final_classification"), Valid: c.FormValue("final_classification") != ""}
	followUp.Serotype = sql.NullString{String: c.FormValue("serotype"), Valid: c.FormValue("serotype") != ""}

	if err := followUp.Insert(db); err != nil {
		sl.Error("Failed to insert follow-up examination", "error", err)
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to save follow-up examination",
		})
	}

	// Create patient history record
	patientHistory := &models.PolioPatientHistory{
		CaseID: caseID,
		Place1: c.FormValue("place1"),
		Place2: c.FormValue("place2"),
		Place3: c.FormValue("place3"),
		Place4: c.FormValue("place4"),
	}

	// Parse duration fields
	parseDuration := func(months, days string) (sql.NullInt32, sql.NullInt32) {
		var monthsVal, daysVal sql.NullInt32
		if months != "" {
			if val, err := strconv.Atoi(months); err == nil {
				monthsVal = sql.NullInt32{Int32: int32(val), Valid: true}
			}
		}
		if days != "" {
			if val, err := strconv.Atoi(days); err == nil {
				daysVal = sql.NullInt32{Int32: int32(val), Valid: true}
			}
		}
		return monthsVal, daysVal
	}

	patientHistory.Duration1Months, patientHistory.Duration1Days = parseDuration(c.FormValue("duration1_months"), c.FormValue("duration1_days"))
	patientHistory.Duration2Months, patientHistory.Duration2Days = parseDuration(c.FormValue("duration2_months"), c.FormValue("duration2_days"))
	patientHistory.Duration3Months, patientHistory.Duration3Days = parseDuration(c.FormValue("duration3_months"), c.FormValue("duration3_days"))
	patientHistory.Duration4Months, patientHistory.Duration4Days = parseDuration(c.FormValue("duration4_months"), c.FormValue("duration4_days"))

	if err := patientHistory.Insert(db); err != nil {
		sl.Error("Failed to insert patient history", "error", err)
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to save patient history",
		})
	}

	// Create investigator record
	investigator := &models.PolioInvestigator{
		CaseID:            caseID,
		InvestigatorName:  c.FormValue("investigator_name"),
		InvestigatorTitle: c.FormValue("investigator_title"),
		Unit:              c.FormValue("unit"),
		Address:           c.FormValue("investigator_address"),
		Telephone:         c.FormValue("investigator_telephone"),
	}

	if err := investigator.Insert(db); err != nil {
		sl.Error("Failed to insert investigator", "error", err)
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to save investigator",
		})
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		sl.Error("Failed to commit transaction", "error", err)
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to save case data",
		})
	}

	// Redirect to success page
	return c.Redirect(fmt.Sprintf("/polio-cif/success?case_id=%s", caseID))
}

// HandlerPolioCIFView handles viewing a polio CIF
func HandlerPolioCIFView(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	idParam := c.Params("id")
	if idParam == "" {
		return c.Status(400).SendString("Case ID required")
	}

	// Determine if this is a numeric ID (primary key) or case_id (string)
	var main models.PolioCaseInvestigation
	var caseID string
	
	// Try querying by numeric ID first
	if idInt, err := strconv.Atoi(idParam); err == nil {
		// It's a numeric ID - query by primary key
		if err := db.QueryRow(`SELECT id, case_id, epid_number, country, region_province, district, year_onset, case_number, received_date, created_at FROM polio_case_investigation WHERE id = $1`, idInt).
			Scan(&main.ID, &main.CaseID, &main.EpidNumber, &main.Country, &main.RegionProvince, &main.District, &main.YearOnset, &main.CaseNumber, &main.ReceivedDate, &main.CreatedAt); err != nil {
			if err == sql.ErrNoRows {
				return c.Status(404).SendString("CIF not found")
			}
			sl.Error("Failed to load case investigation", "error", err, "id", idParam)
			return c.Status(500).SendString("Failed to load case")
		}
		caseID = main.CaseID
	} else {
		// It's a string - query by case_id
		caseID = idParam
		if err := db.QueryRow(`SELECT id, case_id, epid_number, country, region_province, district, year_onset, case_number, received_date, created_at FROM polio_case_investigation WHERE case_id = $1`, caseID).
			Scan(&main.ID, &main.CaseID, &main.EpidNumber, &main.Country, &main.RegionProvince, &main.District, &main.YearOnset, &main.CaseNumber, &main.ReceivedDate, &main.CreatedAt); err != nil {
			if err == sql.ErrNoRows {
				return c.Status(404).SendString("CIF not found")
			}
			sl.Error("Failed to load case investigation", "error", err, "case_id", caseID)
			return c.Status(500).SendString("Failed to load case")
		}
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

	data := NewTemplateData(c, store)
	data.Form = fiber.Map{
		"Case":            main,
		"Identification":  ident,
		"Notification":    notif,
		"Hospitalization": hosp,
		"ClinicalHistory": clin,
		"Immunization":    imm,
		"StoolCollection": stool,
		"StoolResults":    stoolRes,
		"FollowUp":        follow,
		"PatientHistory":  history,
		"Investigator":    investigator,
		"IsView":          true,
	}
	return GenerateHTML(c, db, data, "polio_cif")
}

// HandlerPolioCIFEdit handles editing a polio CIF
func HandlerPolioCIFEdit(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	idParam := c.Params("id")
	if idParam == "" {
		return c.Status(400).SendString("Case ID required")
	}

	// Determine if this is a numeric ID (primary key) or case_id (string)
	var main models.PolioCaseInvestigation
	var caseID string
	
	// Try querying by numeric ID first
	if idInt, err := strconv.Atoi(idParam); err == nil {
		// It's a numeric ID - query by primary key
		if err := db.QueryRow(`SELECT id, case_id, epid_number, country, region_province, district, year_onset, case_number, received_date, created_at FROM polio_case_investigation WHERE id = $1`, idInt).
			Scan(&main.ID, &main.CaseID, &main.EpidNumber, &main.Country, &main.RegionProvince, &main.District, &main.YearOnset, &main.CaseNumber, &main.ReceivedDate, &main.CreatedAt); err != nil {
			if err == sql.ErrNoRows {
				return c.Status(404).SendString("CIF not found")
			}
			sl.Error("Failed to load case investigation", "error", err, "id", idParam)
			return c.Status(500).SendString("Failed to load case")
		}
		caseID = main.CaseID
	} else {
		// It's a string - query by case_id
		caseID = idParam
		if err := db.QueryRow(`SELECT id, case_id, epid_number, country, region_province, district, year_onset, case_number, received_date, created_at FROM polio_case_investigation WHERE case_id = $1`, caseID).
			Scan(&main.ID, &main.CaseID, &main.EpidNumber, &main.Country, &main.RegionProvince, &main.District, &main.YearOnset, &main.CaseNumber, &main.ReceivedDate, &main.CreatedAt); err != nil {
			if err == sql.ErrNoRows {
				return c.Status(404).SendString("CIF not found")
			}
			sl.Error("Failed to load case investigation", "error", err, "case_id", caseID)
			return c.Status(500).SendString("Failed to load case")
		}
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

	data := NewTemplateData(c, store)
	data.Form = fiber.Map{
		"Case":            main,
		"Identification":  ident,
		"Notification":    notif,
		"Hospitalization": hosp,
		"ClinicalHistory": clin,
		"Immunization":    imm,
		"StoolCollection": stool,
		"StoolResults":    stoolRes,
		"FollowUp":        follow,
		"PatientHistory":  history,
		"Investigator":    investigator,
		"IsEdit":          true,
		"Title":           "Polio Case Investigation Form",
	}
	return GenerateHTML(c, db, data, "polio_cif")
}

// HandlerPolioCIFSuccess handles the success page after Polio CIF submission
func HandlerPolioCIFSuccess(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	caseID := c.Query("case_id")

	data := NewTemplateData(c, store)
	data.Form = fiber.Map{
		"Title":   "Polio CIF Submitted Successfully",
		"CaseID":  caseID,
		"Message": "The Polio Case Investigation Form has been submitted successfully.",
	}
	return GenerateHTML(c, db, data, "polio_cif_success")
}
