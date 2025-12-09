package handlers

import (
	"fmt"
	"html/template"
	"net/http"
	"time"

	"database/sql"

	"case/internal/models"

	"log/slog"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
)

func MeaslesCIFHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		// Parse form and populate model (stub, no DB yet)
		// TODO: Implement form parsing and DB save
		http.Redirect(w, r, "/measles_cif?success=1", http.StatusSeeOther)
		return
	}
	tmpl, err := template.ParseFiles("ui/html/measles_cif.html")
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)
}

func HandlerMeaslesCIF(c *fiber.Ctx, db *sql.DB, store *session.Store) error {
	measlesCode := c.FormValue("measles_code")
	if c.Method() == fiber.MethodPost {
		pid := c.FormValue("patient_id")
		// Patch for DOB
		var dob sql.NullTime
		if v := c.FormValue("dob"); v != "" {
			t, err := time.Parse("2006-01-02", v)
			if err == nil {
				dob = sql.NullTime{Time: t, Valid: true}
			}
		}
		// Save patient core info
		patient := &models.MeaslesPatient{
			PatientID:   pid,
			MeaslesCode: measlesCode,
			PatientName: c.FormValue("patient_name"),
			Sex:         c.FormValue("sex"),
			DOB:         dob,
			CreatedAt:   time.Now().Format("2006-01-02 15:04:05"),
		}
		// Save demographics
		demographics := &models.MeaslesDemographics{
			PatientID:          pid,
			OnsetDistrict:      c.FormValue("onset_district"),
			ReportingUnit:      c.FormValue("reporting_unit"),
			AgeMonths:          parseInt(c.FormValue("age_months")),
			HeadOfHousehold:    c.FormValue("head_of_household"),
			GuardianOccupation: c.FormValue("guardian_occupation"),
			HomeDistrict:       c.FormValue("home_district"),
			Subcounty:          c.FormValue("subcounty"),
			Parish:             c.FormValue("parish"),
			LC1Zone:            c.FormValue("lc1_zone"),
			LC1Chairman:        c.FormValue("lc1_chairman"),
			LC1Tel:             c.FormValue("lc1_tel"),
		}
		// Save clinical history
		clinical := &models.MeaslesClinicalHistory{
			PatientID:              pid,
			Fever:                  c.FormValue("fever") == "on",
			FeverOnset:             c.FormValue("fever_onset"),
			Temperature:            parseFloat(c.FormValue("temperature")),
			Rash:                   c.FormValue("rash") == "on",
			RashOnset:              c.FormValue("rash_onset"),
			Cough:                  c.FormValue("cough") == "on",
			RedEyes:                c.FormValue("red_eyes") == "on",
			RunningNose:            c.FormValue("running_nose") == "on",
			OtherComplications:     c.FormValue("other_complications") == "on",
			ComplicationsSpecify:   c.FormValue("complications_specify"),
			Outcome:                parseInt(c.FormValue("outcome")),
			VitaminA:               c.FormValue("vitamin_a") == "on",
			VitaminADoses:          parseInt(c.FormValue("vitamin_a_doses")),
			ImmunisationCardSeen:   c.FormValue("immunisation_card_seen") == "on",
			MeaslesDoses:           parseInt(c.FormValue("measles_doses")),
			LastMeaslesVaccination: c.FormValue("last_measles_vaccination"),
			VaccinationReason:      c.FormValue("vaccination_reason"),
			Diagnosis:              c.FormValue("diagnosis"),
		}

		var bloodCollectionDate = ParseNullTime(c.FormValue("blood_collection_date"))
		var bloodSentDate = ParseNullTime(c.FormValue("blood_sent_date"))
		var bloodReceivedDate = ParseNullTime(c.FormValue("blood_received_date"))
		var urineCollectionDate = ParseNullTime(c.FormValue("urine_collection_date"))
		var urineSentDate = ParseNullTime(c.FormValue("urine_sent_date"))
		var urineReceivedDate = ParseNullTime(c.FormValue("urine_received_date"))
		var formSentDate = ParseNullTime(c.FormValue("form_sent_date"))
		var formReceivedDate = ParseNullTime(c.FormValue("form_received_date"))
		// Save specimens
		specimens := &models.MeaslesSpecimens{
			PatientID:           pid,
			BloodCollectionDate: bloodCollectionDate,
			BloodSentDate:       bloodSentDate,
			BloodReceivedDate:   bloodReceivedDate,
			BloodCondition:      c.FormValue("blood_condition"),
			UrineCollectionDate: urineCollectionDate,
			UrineSentDate:       urineSentDate,
			UrineReceivedDate:   urineReceivedDate,
			UrineCondition:      c.FormValue("urine_condition"),
			FormSentDate:        formSentDate,
			FormReceivedDate:    formReceivedDate,
		}
		// Save investigators
		var investigatorDate = ParseNullTime(c.FormValue("investigator_date"))
		investigators := &models.MeaslesInvestigators{
			PatientID:         pid,
			InvestigatorName:  c.FormValue("investigator_name"),
			InvestigatorTitle: c.FormValue("investigator_title"),
			InvestigatorDate:  investigatorDate,
		}
		// Save results
		var serologyDate = ParseNullTime(c.FormValue("serology_date"))
		var serologyEpiSentDate = ParseNullTime(c.FormValue("serology_epi_sent_date"))
		var virusIsolationDate = ParseNullTime(c.FormValue("virus_isolation_date"))
		var resultsSentDate = ParseNullTime(c.FormValue("results_sent_date"))
		results := &models.MeaslesResults{
			PatientID:           pid,
			SerologyIgM:         c.FormValue("serology_igm"),
			SerologyDate:        serologyDate,
			SerologyEpiSentDate: serologyEpiSentDate,
			VirusIsolationUrine: c.FormValue("virus_isolation_urine"),
			VirusIsolationDate:  virusIsolationDate,
			FinalClassification: parseInt(c.FormValue("final_classification")),
			ResultsSentDate:     resultsSentDate,
		}
		// Insert all
		if err := insertAllMeaslesSections(db, patient, demographics, clinical, specimens, investigators, results); err != nil {
			return c.Status(500).SendString("Failed to save form: " + err.Error())
		}
		return c.Redirect("/measles_success?measles_code=" + measlesCode)
	}
	// Render the form using GenerateHTML helper
	data := NewTemplateData(c, store)
	data.Form = map[string]interface{}{"MeaslesCode": measlesCode}
	data.Optionz = Get_Client_Optionz()
	return GenerateHTML(c, db, data, "measles_cif")
}

func HandlerMeaslesSuccess(c *fiber.Ctx, store *session.Store) error {
	code := c.Query("measles_code")
	data := NewTemplateData(c, store)
	data.Form = map[string]interface{}{"MeaslesCode": code}
	data.Optionz = Get_Client_Optionz()
	return GenerateHTML(c, nil, data, "measles_success")
}

func HandlerMeaslesList(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	// Get current user information (for future use if needed)
	_, _ = GetUser(c, sl, store)

	// Build the query to get measles cases
	query := `
		SELECT 
			mp.patient_id,
			mp.measles_code,
			mp.patient_name,
			mp.sex,
			mp.dob,
			md.age_months,
			md.onset_district,
			mp.created_at,
			CASE WHEN mr.id IS NOT NULL THEN true ELSE false END as results_status
		FROM measles_patients mp
		LEFT JOIN measles_demographics md ON mp.patient_id = md.patient_id
		LEFT JOIN measles_results mr ON mp.patient_id = mr.patient_id
		ORDER BY mp.created_at DESC`

	rows, err := db.QueryContext(c.Context(), query)
	if err != nil {
		sl.Error("Failed to query measles cases", "error", err)
		return c.Status(500).SendString("Failed to retrieve measles cases")
	}
	defer rows.Close()

	var cases []fiber.Map
	for rows.Next() {
		var (
			patientID     string
			measlesCode   sql.NullString
			patientName   string
			sex           string
			dob           sql.NullTime
			ageMonths     sql.NullInt32
			onsetDistrict sql.NullString
			createdAt     time.Time
			resultsStatus bool
		)

		err := rows.Scan(
			&patientID,
			&measlesCode,
			&patientName,
			&sex,
			&dob,
			&ageMonths,
			&onsetDistrict,
			&createdAt,
			&resultsStatus,
		)
		if err != nil {
			sl.Error("Failed to scan measles case", "error", err)
			continue
		}

		// Calculate age display
		var ageDisplay string
		if ageMonths.Valid {
			ageDisplay = fmt.Sprintf("%d months", ageMonths.Int32)
		} else if dob.Valid {
			age := time.Now().Year() - dob.Time.Year()
			if time.Now().YearDay() < dob.Time.YearDay() {
				age--
			}
			ageDisplay = fmt.Sprintf("%d years", age)
		}

		cases = append(cases, fiber.Map{
			"ID":            patientID,
			"MeaslesCode":   measlesCode.String,
			"Name":          patientName,
			"Age":           ageDisplay,
			"Gender":        sex,
			"District":      onsetDistrict.String,
			"Status":        "Active", // Default status for measles cases
			"ResultsStatus": resultsStatus,
			"CreatedAt":     createdAt.Format("2006-01-02 15:04"),
		})
	}

	if err = rows.Err(); err != nil {
		sl.Error("Error iterating measles cases", "error", err)
		return c.Status(500).SendString("Error retrieving measles cases")
	}

	// Check if this is an API request
	if c.Get("Accept") == "application/json" {
		return c.JSON(cases)
	}

	// Return HTML response
	data := NewTemplateData(c, store)
	data.Form = fiber.Map{
		"Title": "Measles Cases",
		"Cases": cases,
	}
	return GenerateHTML(c, db, data, "measles_list")
}

func insertAllMeaslesSections(db *sql.DB, patient *models.MeaslesPatient, demographics *models.MeaslesDemographics, clinical *models.MeaslesClinicalHistory, specimens *models.MeaslesSpecimens, investigators *models.MeaslesInvestigators, results *models.MeaslesResults) error {
	if err := patient.Insert(db); err != nil {
		return err
	}
	if err := demographics.Insert(db); err != nil {
		return err
	}
	if err := clinical.Insert(db); err != nil {
		return err
	}
	if err := specimens.Insert(db); err != nil {
		return err
	}
	if err := investigators.Insert(db); err != nil {
		return err
	}
	if err := results.Insert(db); err != nil {
		return err
	}
	return nil
}

func parseInt(s string) int {
	var i int
	fmt.Sscanf(s, "%d", &i)
	return i
}

// HandlerMeaslesCIFView handles viewing a measles CIF
func HandlerMeaslesCIFView(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	pid := c.Params("id")
	if pid == "" {
		return c.Status(400).SendString("Patient ID required")
	}

	// Fetch patient data
	var patient models.MeaslesPatient
	err := db.QueryRow(`SELECT patient_id, measles_code, patient_name, sex, dob, created_at FROM measles_patients WHERE patient_id = $1`, pid).
		Scan(&patient.PatientID, &patient.MeaslesCode, &patient.PatientName, &patient.Sex, &patient.DOB, &patient.CreatedAt)
	if err == sql.ErrNoRows {
		return c.Status(404).SendString("CIF not found")
	}
	if err != nil {
		sl.Error("Failed to load patient", "error", err, "patient_id", pid)
		return c.Status(500).SendString("Failed to load patient")
	}

	// Fetch all related data
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

	// Render template
	data := NewTemplateData(c, store)
	data.Form = fiber.Map{
		"Patient":          patient,
		"Demographics":     demo,
		"Investigators":    inv,
		"ClinicalHistory":  hist,
		"Results":          res,
		"Specimens":        spec,
		"IsView":           true,
	}
	data.Optionz = Get_Client_Optionz()
	return GenerateHTML(c, db, data, "measles_cif")
}

// HandlerMeaslesCIFEdit handles editing a measles CIF
func HandlerMeaslesCIFEdit(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	pid := c.Params("id")
	if pid == "" {
		return c.Status(400).SendString("Patient ID required")
	}

	// Fetch patient data (same as view)
	var patient models.MeaslesPatient
	err := db.QueryRow(`SELECT patient_id, measles_code, patient_name, sex, dob, created_at FROM measles_patients WHERE patient_id = $1`, pid).
		Scan(&patient.PatientID, &patient.MeaslesCode, &patient.PatientName, &patient.Sex, &patient.DOB, &patient.CreatedAt)
	if err == sql.ErrNoRows {
		return c.Status(404).SendString("CIF not found")
	}
	if err != nil {
		sl.Error("Failed to load patient", "error", err, "patient_id", pid)
		return c.Status(500).SendString("Failed to load patient")
	}

	// Fetch all related data (same as view)
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

	// Render template with edit mode
	data := NewTemplateData(c, store)
	data.Form = fiber.Map{
		"Patient":          patient,
		"Demographics":     demo,
		"Investigators":    inv,
		"ClinicalHistory":  hist,
		"Results":          res,
		"Specimens":        spec,
		"IsEdit":           true,
		"MeaslesCode":      patient.MeaslesCode,
	}
	data.Optionz = Get_Client_Optionz()
	return GenerateHTML(c, db, data, "measles_cif")
}
