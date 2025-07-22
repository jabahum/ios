package models

import "database/sql"

type MeaslesPatient struct {
	PatientID   string
	MeaslesCode string
	PatientName string
	Sex         string
	DOB         string
	CreatedAt   string
}

type MeaslesClinicalHistory struct {
	ID                     int
	PatientID              string
	Fever                  bool
	FeverOnset             string
	Temperature            float64
	Rash                   bool
	RashOnset              string
	Cough                  bool
	RedEyes                bool
	RunningNose            bool
	OtherComplications     bool
	ComplicationsSpecify   string
	Outcome                int
	VitaminA               bool
	VitaminADoses          int
	ImmunisationCardSeen   bool
	MeaslesDoses           int
	LastMeaslesVaccination string
	VaccinationReason      string
	Diagnosis              string
}

type MeaslesDemographics struct {
	ID                 int
	PatientID          string
	OnsetDistrict      string
	ReportingUnit      string
	AgeMonths          int
	HeadOfHousehold    string
	GuardianOccupation string
	HomeDistrict       string
	Subcounty          string
	Parish             string
	LC1Zone            string
	LC1Chairman        string
	LC1Tel             string
}

type MeaslesInvestigators struct {
	ID                int
	PatientID         string
	InvestigatorName  string
	InvestigatorTitle string
	InvestigatorDate  string
}

type MeaslesResults struct {
	ID                  int
	PatientID           string
	SerologyIgM         string
	SerologyDate        string
	SerologyEpiSentDate string
	VirusIsolationUrine string
	VirusIsolationDate  string
	FinalClassification int
	ResultsSentDate     string
}

type MeaslesSpecimens struct {
	ID                  int
	PatientID           string
	BloodCollectionDate string
	BloodSentDate       string
	BloodReceivedDate   string
	BloodCondition      string
	UrineCollectionDate string
	UrineSentDate       string
	UrineReceivedDate   string
	UrineCondition      string
	FormSentDate        string
	FormReceivedDate    string
}

func (m *MeaslesPatient) Insert(db *sql.DB) error {
	query := `INSERT INTO measles_patients (patient_id, measles_code, patient_name, sex, dob, created_at) VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := db.Exec(query, m.PatientID, m.MeaslesCode, m.PatientName, m.Sex, m.DOB, m.CreatedAt)
	return err
}
func (m *MeaslesInvestigators) Insert(db *sql.DB) error {
	query := `INSERT INTO measles_investigators (patient_id, investigator_name, investigator_title, investigator_date) VALUES ($1, $2, $3, $4)`
	_, err := db.Exec(query, m.PatientID, m.InvestigatorName, m.InvestigatorTitle, m.InvestigatorDate)
	return err
}

func (m *MeaslesClinicalHistory) Insert(db *sql.DB) error {
	query := `INSERT INTO measles_clinical_history (patient_id, fever, fever_onset, temperature, rash, rash_onset, cough, red_eyes, running_nose, other_complications, complications_specify, outcome, vitamin_a, vitamin_a_doses, immunisation_card_seen, measles_doses, last_measles_vaccination, vaccination_reason, diagnosis) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19)`
	_, err := db.Exec(query, m.PatientID, m.Fever, m.FeverOnset, m.Temperature, m.Rash, m.RashOnset, m.Cough, m.RedEyes, m.RunningNose, m.OtherComplications, m.ComplicationsSpecify, m.Outcome, m.VitaminA, m.VitaminADoses, m.ImmunisationCardSeen, m.MeaslesDoses, m.LastMeaslesVaccination, m.VaccinationReason, m.Diagnosis)
	return err
}

func (m *MeaslesDemographics) Insert(db *sql.DB) error {
	query := `INSERT INTO measles_demographics (patient_id, onset_district, reporting_unit, age_months, head_of_household, guardian_occupation, home_district, subcounty, parish, lc1_zone, lc1_chairman, lc1_tel) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`
	_, err := db.Exec(query, m.PatientID, m.OnsetDistrict, m.ReportingUnit, m.AgeMonths, m.HeadOfHousehold, m.GuardianOccupation, m.HomeDistrict, m.Subcounty, m.Parish, m.LC1Zone, m.LC1Chairman, m.LC1Tel)
	return err
}

func (m *MeaslesResults) Insert(db *sql.DB) error {
	query := `INSERT INTO measles_results (patient_id, serology_igm, serology_date, serology_epi_sent_date, virus_isolation_urine, virus_isolation_date, final_classification, results_sent_date) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	_, err := db.Exec(query, m.PatientID, m.SerologyIgM, m.SerologyDate, m.SerologyEpiSentDate, m.VirusIsolationUrine, m.VirusIsolationDate, m.FinalClassification, m.ResultsSentDate)
	return err
}

func (m *MeaslesSpecimens) Insert(db *sql.DB) error {
	query := `INSERT INTO measles_specimens (patient_id, blood_collection_date, blood_sent_date, blood_received_date, blood_condition, urine_collection_date, urine_sent_date, urine_received_date, urine_condition, form_sent_date, form_received_date) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`
	_, err := db.Exec(query, m.PatientID, m.BloodCollectionDate, m.BloodSentDate, m.BloodReceivedDate, m.BloodCondition, m.UrineCollectionDate, m.UrineSentDate, m.UrineReceivedDate, m.UrineCondition, m.FormSentDate, m.FormReceivedDate)
	return err
}
