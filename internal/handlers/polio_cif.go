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

	// Parse immunization dates
	parseDate := func(dateStr string) sql.NullTime {
		if dateStr == "" {
			return sql.NullTime{Valid: false}
		}
		if t, err := time.Parse("2006-01-02", dateStr); err == nil {
			return sql.NullTime{Time: t, Valid: true}
		}
		return sql.NullTime{Valid: false}
	}

	immunization.OPVDoseAtBirth = parseDate(c.FormValue("opv_dose_at_birth"))
	immunization.OPVDose1 = parseDate(c.FormValue("opv_dose1"))
	immunization.OPVDose2 = parseDate(c.FormValue("opv_dose2"))
	immunization.OPVDose3 = parseDate(c.FormValue("opv_dose3"))
	immunization.OPVDose4 = parseDate(c.FormValue("opv_dose4"))
	immunization.OPVDoseMoreThan4 = parseDate(c.FormValue("opv_dose_more_than4"))
	immunization.LastOPVDose = parseDate(c.FormValue("last_opv_dose"))
	immunization.LastOPVSIA = parseDate(c.FormValue("last_opv_sia"))
	immunization.LastIPVSIA = parseDate(c.FormValue("last_ipv_sia"))

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

	specimenCollection.DateFirstSpecimen = parseDate(c.FormValue("date_first_specimen"))
	specimenCollection.DateSecondSpecimen = parseDate(c.FormValue("date_second_specimen"))
	specimenCollection.DateSpecimenSentNational = parseDate(c.FormValue("date_specimen_sent_national"))
	specimenCollection.DateSpecimenReceivedNational = parseDate(c.FormValue("date_specimen_received_national"))
	specimenCollection.DateSpecimenSentLab = parseDate(c.FormValue("date_specimen_sent_lab"))

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

	specimenResults.DateReceivedAtLab = parseDate(c.FormValue("date_received_at_lab"))
	specimenResults.SpecimenStatusAtReception = sql.NullString{String: c.FormValue("specimen_status_at_reception"), Valid: c.FormValue("specimen_status_at_reception") != ""}
	specimenResults.DateCombinedCellCulture = parseDate(c.FormValue("date_combined_cell_culture"))
	specimenResults.DateResultsSentToEPI = parseDate(c.FormValue("date_results_sent_to_epi"))
	specimenResults.DateResultsReceivedAtEPI = parseDate(c.FormValue("date_results_received_at_epi"))
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
	specimenResults.DateSentToRegionalLab = parseDate(c.FormValue("date_sent_to_regional_lab"))
	specimenResults.DateITDifferentiationSent = parseDate(c.FormValue("date_it_differentiation_sent"))
	specimenResults.DateITDifferentiationReceived = parseDate(c.FormValue("date_it_differentiation_received"))
	specimenResults.DateIsolateSentSequencing = parseDate(c.FormValue("date_isolate_sent_sequencing"))
	specimenResults.DateSeqResultsSentProgram = parseDate(c.FormValue("date_seq_results_sent_program"))

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

	followUp.DateOfFollowUp = parseDate(c.FormValue("date_of_follow_up"))
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
