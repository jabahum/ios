package handlers

import (
	"fmt"
	"html/template"
	"net/http"
	"time"

	"database/sql"

	"case/internal/models"

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
	if c.Method() == fiber.MethodPost {
		pid := c.FormValue("patient_id")
		measlesCode := c.FormValue("measles_code")
		// Save patient core info
		patient := &models.MeaslesPatient{
			PatientID:   pid,
			MeaslesCode: measlesCode,
			PatientName: c.FormValue("patient_name"),
			Sex:         c.FormValue("sex"),
			DOB:         c.FormValue("dob"),
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
		// Save specimens
		specimens := &models.MeaslesSpecimens{
			PatientID:           pid,
			BloodCollectionDate: c.FormValue("blood_collection_date"),
			BloodSentDate:       c.FormValue("blood_sent_date"),
			BloodReceivedDate:   c.FormValue("blood_received_date"),
			BloodCondition:      c.FormValue("blood_condition"),
			UrineCollectionDate: c.FormValue("urine_collection_date"),
			UrineSentDate:       c.FormValue("urine_sent_date"),
			UrineReceivedDate:   c.FormValue("urine_received_date"),
			UrineCondition:      c.FormValue("urine_condition"),
			FormSentDate:        c.FormValue("form_sent_date"),
			FormReceivedDate:    c.FormValue("form_received_date"),
		}
		// Save investigators
		investigators := &models.MeaslesInvestigators{
			PatientID:         pid,
			InvestigatorName:  c.FormValue("investigator_name"),
			InvestigatorTitle: c.FormValue("investigator_title"),
			InvestigatorDate:  c.FormValue("investigator_date"),
		}
		// Save results
		results := &models.MeaslesResults{
			PatientID:           pid,
			SerologyIgM:         c.FormValue("serology_igm"),
			SerologyDate:        c.FormValue("serology_date"),
			SerologyEpiSentDate: c.FormValue("serology_epi_sent_date"),
			VirusIsolationUrine: c.FormValue("virus_isolation_urine"),
			VirusIsolationDate:  c.FormValue("virus_isolation_date"),
			FinalClassification: parseInt(c.FormValue("final_classification")),
			ResultsSentDate:     c.FormValue("results_sent_date"),
		}
		// Insert all
		if err := insertAllMeaslesSections(db, patient, demographics, clinical, specimens, investigators, results); err != nil {
			return c.Status(500).SendString("Failed to save form: " + err.Error())
		}
		return c.Redirect("/measles_success?measles_code=" + measlesCode)
	}
	// Render the form using GenerateHTML helper
	data := NewTemplateData(c, store)
	return GenerateHTML(c, db, data, "measles_cif")
}

func HandlerMeaslesSuccess(c *fiber.Ctx, store *session.Store) error {
	code := c.Query("measles_code")
	data := NewTemplateData(c, store)
	data.Form = map[string]interface{}{"MeaslesCode": code}
	return GenerateHTML(c, nil, data, "measles_success")
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
