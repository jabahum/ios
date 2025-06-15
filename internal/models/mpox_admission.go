package models

import (
	"database/sql"
	"time"
)

// MpoxDemographics represents the demographics information
type MpoxDemographics struct {
	ID                   int
	ClientID             sql.NullInt64
	Sex                  sql.NullString
	DateOfBirth          sql.NullTime
	AgeYears             sql.NullInt64
	AgeMonths            sql.NullInt64
	AgeDays              sql.NullInt64
	HealthCareWorker     sql.NullString
	LaboratoryWorker     sql.NullString
	PPEStatus            sql.NullString
	Tribe                sql.NullString
	Pregnant             sql.NullBool
	GestationalWeeks     sql.NullInt64
	LMNP                 sql.NullTime
	RecentlyPregnant     sql.NullBool
	Pregnant22_42        sql.NullBool
	TetanusVaccination   sql.NullBool
	Occupation           sql.NullString
	SiteOfFirstEncounter sql.NullString
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

// MpoxExposureHistory represents exposure and social history
type MpoxExposureHistory struct {
	ID             int
	DemographicsID int
	KnownLink      sql.NullBool
	SexuallyActive sql.NullBool
	SexOfPartners  sql.NullString
	RecentTravel   sql.NullBool
	TravelHighRisk sql.NullBool
	TravelDetails  sql.NullString
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// MpoxOnsetVitals represents date of onset and vital signs
type MpoxOnsetVitals struct {
	ID              int
	DemographicsID  int
	SymptomOnset    sql.NullTime
	Fever           sql.NullBool
	SoreThroat      sql.NullBool
	Headache        sql.NullBool
	MuscleAches     sql.NullBool
	Cough           sql.NullBool
	Fatigue         sql.NullBool
	OralPain        sql.NullBool
	Nausea          sql.NullBool
	Vomiting        sql.NullBool
	Diarrhea        sql.NullBool
	RectalPain      sql.NullBool
	Lesions         sql.NullBool
	Lymphadenopathy sql.NullBool
	Temperature     sql.NullFloat64
	HeartRate       sql.NullInt64
	RespiratoryRate sql.NullInt64
	BpSystolic      sql.NullInt64
	BpDiastolic     sql.NullInt64
	Dehydration     sql.NullBool
	AVPU            sql.NullString
	HeightCm        sql.NullFloat64
	WeightKg        sql.NullFloat64
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// MpoxComorbidities represents co-morbidities
type MpoxComorbidities struct {
	ID                   int
	DemographicsID       int
	CardiacDisease       sql.NullBool
	Hypertension         sql.NullBool
	PulmonaryDisease     sql.NullBool
	Asthma               sql.NullBool
	KidneyDisease        sql.NullBool
	LiverDisease         sql.NullBool
	NeurologicalDisorder sql.NullBool
	Diabetes             sql.NullBool
	TuberculosisActive   sql.NullBool
	TuberculosisPrevious sql.NullBool
	Asplenia             sql.NullBool
	Neoplasm             sql.NullBool
	AlcoholUseDisorder   sql.NullBool
	Immunosuppressants   sql.NullBool
	STI                  sql.NullBool
	HIVStatus            sql.NullString
	ARTRegimen           sql.NullString
	CD4                  sql.NullInt64
	ViralLoad            sql.NullString
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

// MpoxRashEvaluation represents rash evaluation
type MpoxRashEvaluation struct {
	ID                   int
	DemographicsID       int
	Severity             sql.NullString
	Face                 sql.NullBool
	Nares                sql.NullBool
	Mouth                sql.NullBool
	Chest                sql.NullBool
	Abdomen              sql.NullBool
	Back                 sql.NullBool
	Perianal             sql.NullBool
	Genitals             sql.NullBool
	Palms                sql.NullBool
	Arms                 sql.NullBool
	Forearms             sql.NullBool
	Thighs               sql.NullBool
	Legs                 sql.NullBool
	Soles                sql.NullBool
	Macule               sql.NullBool
	Papule               sql.NullBool
	EarlyVesicle         sql.NullBool
	SmallPustule         sql.NullBool
	UmbilicatedPustule   sql.NullBool
	UlceratedLesion      sql.NullBool
	Crusting             sql.NullBool
	PartiallyRemovedScab sql.NullBool
	PainAtLesion         sql.NullBool
	PainScore            sql.NullInt64
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

// MpoxLaboratoryInvestigations represents laboratory investigations
type MpoxLaboratoryInvestigations struct {
	ID              int
	DemographicsID  int
	ALT             sql.NullFloat64
	AST             sql.NullFloat64
	Creatinine      sql.NullFloat64
	Potassium       sql.NullFloat64
	Urea            sql.NullFloat64
	CreatineKinase  sql.NullFloat64
	Calcium         sql.NullFloat64
	Sodium          sql.NullFloat64
	CRP             sql.NullFloat64
	Glucose         sql.NullFloat64
	Lactate         sql.NullFloat64
	Haemoglobin     sql.NullFloat64
	TotalBilirubin  sql.NullFloat64
	WBCCount        sql.NullFloat64
	Platelets       sql.NullFloat64
	ProthrombinTime sql.NullFloat64
	APTT            sql.NullFloat64
	MalariaResult   sql.NullString
	SyphilisResult  sql.NullString
	MpoxResult      sql.NullString
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// MpoxDataEntrant represents the data entrant information
type MpoxDataEntrant struct {
	ID             int
	DemographicsID int
	Name           sql.NullString
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// Insert methods for each struct
func (d *MpoxDemographics) Insert(db *sql.DB) error {
	query := `
		INSERT INTO mpox_demographics (
			client_id, sex, date_of_birth, age_years, age_months, age_days,
			health_care_worker, laboratory_worker, ppe_status, tribe, pregnant,
			gestational_weeks, lmnp, recently_pregnant, pregnant_22_42,
			tetanus_vaccination, occupation, site_of_first_encounter,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15,
			$16, $17, $18, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
		) RETURNING id`
	return db.QueryRow(query,
		d.ClientID, d.Sex, d.DateOfBirth, d.AgeYears, d.AgeMonths, d.AgeDays,
		d.HealthCareWorker, d.LaboratoryWorker, d.PPEStatus, d.Tribe, d.Pregnant,
		d.GestationalWeeks, d.LMNP, d.RecentlyPregnant, d.Pregnant22_42,
		d.TetanusVaccination, d.Occupation, d.SiteOfFirstEncounter).Scan(&d.ID)
}

func (e *MpoxExposureHistory) Insert(db *sql.DB) error {
	query := `
		INSERT INTO mpox_exposure_history (
			demographics_id, known_link, sexually_active, sex_of_partners,
			recent_travel, travel_high_risk, travel_details,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7,
			CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
		) RETURNING id`
	return db.QueryRow(query,
		e.DemographicsID, e.KnownLink, e.SexuallyActive, e.SexOfPartners,
		e.RecentTravel, e.TravelHighRisk, e.TravelDetails).Scan(&e.ID)
}

func (v *MpoxOnsetVitals) Insert(db *sql.DB) error {
	query := `
		INSERT INTO mpox_onset_vitals (
			demographics_id, symptom_onset, fever, sore_throat, headache,
			muscle_aches, cough, fatigue, oral_pain, nausea, vomiting,
			diarrhea, rectal_pain, lesions, lymphadenopathy, temperature,
			heart_rate, respiratory_rate, bp_systolic, bp_diastolic,
			dehydration, avpu, height_cm, weight_kg,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15,
			$16, $17, $18, $19, $20, $21, $22, $23, $24,
			CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
		) RETURNING id`
	return db.QueryRow(query,
		v.DemographicsID, v.SymptomOnset, v.Fever, v.SoreThroat, v.Headache,
		v.MuscleAches, v.Cough, v.Fatigue, v.OralPain, v.Nausea, v.Vomiting,
		v.Diarrhea, v.RectalPain, v.Lesions, v.Lymphadenopathy, v.Temperature,
		v.HeartRate, v.RespiratoryRate, v.BpSystolic, v.BpDiastolic,
		v.Dehydration, v.AVPU, v.HeightCm, v.WeightKg).Scan(&v.ID)
}

func (c *MpoxComorbidities) Insert(db *sql.DB) error {
	query := `
		INSERT INTO mpox_comorbidities (
			demographics_id, cardiac_disease, hypertension, pulmonary_disease,
			asthma, kidney_disease, liver_disease, neurological_disorder,
			diabetes, tuberculosis_active, tuberculosis_previous, asplenia,
			neoplasm, alcohol_use_disorder, immunosuppressants, sti,
			hiv_status, art_regimen, cd4, viral_load,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15,
			$16, $17, $18, $19, $20,
			CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
		) RETURNING id`
	return db.QueryRow(query,
		c.DemographicsID, c.CardiacDisease, c.Hypertension, c.PulmonaryDisease,
		c.Asthma, c.KidneyDisease, c.LiverDisease, c.NeurologicalDisorder,
		c.Diabetes, c.TuberculosisActive, c.TuberculosisPrevious, c.Asplenia,
		c.Neoplasm, c.AlcoholUseDisorder, c.Immunosuppressants, c.STI,
		c.HIVStatus, c.ARTRegimen, c.CD4, c.ViralLoad).Scan(&c.ID)
}

func (r *MpoxRashEvaluation) Insert(db *sql.DB) error {
	query := `
		INSERT INTO mpox_rash_evaluation (
			demographics_id, severity, face, nares, mouth, chest, abdomen,
			back, perianal, genitals, palms, arms, forearms, thighs, legs,
			soles, macule, papule, early_vesicle, small_pustule,
			umbilicated_pustule, ulcerated_lesion, crusting,
			partially_removed_scab, pain_at_lesion, pain_score,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15,
			$16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26,
			CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
		) RETURNING id`
	return db.QueryRow(query,
		r.DemographicsID, r.Severity, r.Face, r.Nares, r.Mouth, r.Chest, r.Abdomen,
		r.Back, r.Perianal, r.Genitals, r.Palms, r.Arms, r.Forearms, r.Thighs, r.Legs,
		r.Soles, r.Macule, r.Papule, r.EarlyVesicle, r.SmallPustule,
		r.UmbilicatedPustule, r.UlceratedLesion, r.Crusting,
		r.PartiallyRemovedScab, r.PainAtLesion, r.PainScore).Scan(&r.ID)
}

func (l *MpoxLaboratoryInvestigations) Insert(db *sql.DB) error {
	query := `
		INSERT INTO mpox_laboratory_investigations (
			demographics_id, alt, ast, creatinine, potassium, urea,
			creatine_kinase, calcium, sodium, crp, glucose, lactate,
			haemoglobin, total_bilirubin, wbc_count, platelets,
			prothrombin_time, aptt, malaria_result, syphilis_result,
			mpox_result, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15,
			$16, $17, $18, $19, $20, $21,
			CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
		) RETURNING id`
	return db.QueryRow(query,
		l.DemographicsID, l.ALT, l.AST, l.Creatinine, l.Potassium, l.Urea,
		l.CreatineKinase, l.Calcium, l.Sodium, l.CRP, l.Glucose, l.Lactate,
		l.Haemoglobin, l.TotalBilirubin, l.WBCCount, l.Platelets,
		l.ProthrombinTime, l.APTT, l.MalariaResult, l.SyphilisResult,
		l.MpoxResult).Scan(&l.ID)
}

func (d *MpoxDataEntrant) Insert(db *sql.DB) error {
	query := `
		INSERT INTO mpox_data_entrant (
			demographics_id, name, created_at, updated_at
		) VALUES (
			$1, $2, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
		) RETURNING id`
	return db.QueryRow(query, d.DemographicsID, d.Name).Scan(&d.ID)
}
