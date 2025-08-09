package models

import (
	"database/sql"
	"time"
)

// MpoxDemographics represents the demographics information
type MpoxDemographics struct {
	ID                        int
	ClientID                  sql.NullInt64
	Sex                       sql.NullString
	DateOfBirth               sql.NullTime
	AgeYears                  sql.NullInt64
	AgeMonths                 sql.NullInt64
	AgeDays                   sql.NullInt64
	HealthCareWorker          sql.NullString
	LaboratoryWorker          sql.NullString
	PPEStatus                 sql.NullString
	Tribe                     sql.NullString
	Pregnant                  sql.NullBool
	GestationalWeeks          sql.NullInt64
	LMNP                      sql.NullTime
	RecentlyPregnant          sql.NullBool
	Pregnant22_42             sql.NullBool
	TetanusVaccination        sql.NullBool
	Occupation                sql.NullString
	SiteOfFirstEncounter      sql.NullString
	SiteOfFirstEncounterOther sql.NullString
	SuspectConfirmedCase      sql.NullString
	LymphPainful              sql.NullString
	LymphLocation             sql.NullString
	LymphOtherDetail          sql.NullString
	LymphPainLocation         sql.NullString
	LymphPainOtherDetail      sql.NullString
	CreatedAt                 time.Time
	UpdatedAt                 time.Time
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
	ID                                int
	DemographicsID                    int
	SymptomOnset                      sql.NullTime
	Fever                             sql.NullBool
	FeverOnsetDate                    sql.NullTime
	SoreThroat                        sql.NullBool
	SoreThroatOnsetDate               sql.NullTime
	Headache                          sql.NullBool
	HeadacheOnsetDate                 sql.NullTime
	MuscleAches                       sql.NullBool
	MuscleAchesOnsetDate              sql.NullTime
	Cough                             sql.NullBool
	CoughOnsetDate                    sql.NullTime
	Fatigue                           sql.NullBool
	FatigueOnsetDate                  sql.NullTime
	OralPain                          sql.NullBool
	OralPainOnsetDate                 sql.NullTime
	Nausea                            sql.NullBool
	NauseaOnsetDate                   sql.NullTime
	Vomiting                          sql.NullBool
	VomitingOnsetDate                 sql.NullTime
	Diarrhea                          sql.NullBool
	DiarrheaOnsetDate                 sql.NullTime
	RectalPain                        sql.NullBool
	RectalPainOnsetDate               sql.NullTime
	Lesions                           sql.NullBool
	LesionsOnsetDate                  sql.NullTime
	Lymphadenopathy                   sql.NullBool
	LymphadenopathyOnsetDate          sql.NullTime
	Pruritis                          sql.NullBool
	PruritisOnsetDate                 sql.NullTime
	PainSwallowing                    sql.NullBool
	PainSwallowingOnsetDate           sql.NullTime
	DifficultySwallowing              sql.NullBool
	DifficultySwallowingOnsetDate     sql.NullTime
	Urethritis                        sql.NullBool
	UrethritisOnsetDate               sql.NullTime
	ChestPain                         sql.NullBool
	ChestPainOnsetDate                sql.NullTime
	DecreasedUrine                    sql.NullBool
	DecreasedUrineOnsetDate           sql.NullTime
	Dizziness                         sql.NullBool
	DizzinessOnsetDate                sql.NullTime
	JointPain                         sql.NullBool
	JointPainOnsetDate                sql.NullTime
	PsychologicalDisturbance          sql.NullBool
	PsychologicalDisturbanceOnsetDate sql.NullTime
	Temperature                       sql.NullFloat64
	HeartRate                         sql.NullInt64
	RespiratoryRate                   sql.NullInt64
	BpSystolic                        sql.NullInt64
	BpDiastolic                       sql.NullInt64
	Dehydration                       sql.NullBool
	AVPU                              sql.NullString
	HeightCm                          sql.NullFloat64
	WeightKg                          sql.NullFloat64
	CreatedAt                         time.Time
	UpdatedAt                         time.Time
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
	ID                    int
	DemographicsID        int
	ALT                   sql.NullFloat64
	AST                   sql.NullFloat64
	Creatinine            sql.NullFloat64
	Potassium             sql.NullFloat64
	Urea                  sql.NullFloat64
	CreatineKinase        sql.NullFloat64
	Calcium               sql.NullFloat64
	Sodium                sql.NullFloat64
	CRP                   sql.NullFloat64
	Glucose               sql.NullFloat64
	Lactate               sql.NullFloat64
	Haemoglobin           sql.NullFloat64
	TotalBilirubin        sql.NullFloat64
	WBCCount              sql.NullFloat64
	Platelets             sql.NullFloat64
	ProthrombinTime       sql.NullFloat64
	APTT                  sql.NullFloat64
	TotalProtein          sql.NullFloat64
	Albumin               sql.NullFloat64
	BilirubinD            sql.NullFloat64
	Lymphocytes           sql.NullFloat64
	Monocytes             sql.NullFloat64
	Eosinophils           sql.NullFloat64
	Basophils             sql.NullFloat64
	Neutrophils           sql.NullFloat64
	HGB                   sql.NullFloat64
	HCT                   sql.NullFloat64
	MCV                   sql.NullFloat64
	MCH                   sql.NullFloat64
	MCHC                  sql.NullFloat64
	RDW                   sql.NullFloat64
	RDWSD                 sql.NullFloat64
	RDWCV                 sql.NullFloat64
	MPV                   sql.NullFloat64
	PDW                   sql.NullString
	PCT                   sql.NullFloat64
	LabOther              sql.NullString
	LabALTNotDone         sql.NullBool
	LabASTNotDone         sql.NullBool
	LabCreatinineNotDone  sql.NullBool
	LabPotassiumNotDone   sql.NullBool
	TotalProteinNotDone   sql.NullBool
	AlbuminNotDone        sql.NullBool
	LabUreaNotDone        sql.NullBool
	LabCKNotDone          sql.NullBool
	LabCalciumNotDone     sql.NullBool
	LabSodiumNotDone      sql.NullBool
	LabLymphocytesNotDone sql.NullBool
	LabMonocytesNotDone   sql.NullBool
	LabEosinophilsNotDone sql.NullBool
	LabBasophilsNotDone   sql.NullBool
	LabCRPNotDone         sql.NullBool
	LabNeutrophilsNotDone sql.NullBool
	LabHGBNotDone         sql.NullBool
	LabHCTNotDone         sql.NullBool
	LabMCVNotDone         sql.NullBool
	LabMCHNotDone         sql.NullBool
	LabMCHCNotDone        sql.NullBool
	LabRDWNotDone         sql.NullBool
	LabRDWSDNotDone       sql.NullBool
	LabRDWCVNotDone       sql.NullBool
	LabMPVNotDone         sql.NullBool
	LabPDWNotDone         sql.NullBool
	LabPCTNotDone         sql.NullBool
	LabOtherNotDone       sql.NullBool
	LabGlucoseNotDone     sql.NullBool
	LabLactateNotDone     sql.NullBool
	LabHaemoglobinNotDone sql.NullBool
	LabBilirubinNotDone   sql.NullBool
	LabBilirubinDNotDone  sql.NullBool
	LabWBCNotDone         sql.NullBool
	LabPlateletsNotDone   sql.NullBool
	LabProthrombinNotDone sql.NullBool
	LabAPTTNotDone        sql.NullBool
	OtherMalaria          sql.NullString
	OtherHIV              sql.NullString
	OtherSyphilis         sql.NullString
	OtherMpox             sql.NullString
	HepatitisB            sql.NullString
	HepatitisC            sql.NullString
	DataEntrantName       sql.NullString
	CreatedAt             time.Time
	UpdatedAt             time.Time
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
			site_of_first_encounter_other, suspect_confirmed_case,
			lymph_painful, lymph_location, lymph_other_detail,
			lymph_pain_location, lymph_pain_other_detail,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15,
			$16, $17, $18, $19, $20, $21, $22, $23, $24, $25,
			CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
		) RETURNING id`
	return db.QueryRow(query,
		d.ClientID, d.Sex, d.DateOfBirth, d.AgeYears, d.AgeMonths, d.AgeDays,
		d.HealthCareWorker, d.LaboratoryWorker, d.PPEStatus, d.Tribe, d.Pregnant,
		d.GestationalWeeks, d.LMNP, d.RecentlyPregnant, d.Pregnant22_42,
		d.TetanusVaccination, d.Occupation, d.SiteOfFirstEncounter,
		d.SiteOfFirstEncounterOther, d.SuspectConfirmedCase,
		d.LymphPainful, d.LymphLocation, d.LymphOtherDetail,
		d.LymphPainLocation, d.LymphPainOtherDetail).Scan(&d.ID)
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
			demographics_id, symptom_onset, fever, fever_onset_date, sore_throat, sore_throat_onset_date,
			headache, headache_onset_date, muscle_aches, muscle_aches_onset_date, cough, cough_onset_date,
			fatigue, fatigue_onset_date, oral_pain, oral_pain_onset_date, nausea, nausea_onset_date,
			vomiting, vomiting_onset_date, diarrhea, diarrhea_onset_date, rectal_pain, rectal_pain_onset_date,
			lesions, lesions_onset_date, lymphadenopathy, lymphadenopathy_onset_date, pruritis, pruritis_onset_date,
			pain_swallowing, pain_swallowing_onset_date, difficulty_swallowing, difficulty_swallowing_onset_date,
			urethritis, urethritis_onset_date, chest_pain, chest_pain_onset_date, decreased_urine, decreased_urine_onset_date,
			dizziness, dizziness_onset_date, joint_pain, joint_pain_onset_date, psychological_disturbance, psychological_disturbance_onset_date,
			temperature, heart_rate, respiratory_rate, bp_systolic, bp_diastolic, dehydration, avpu, height_cm, weight_kg
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16,
			$17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28, $29, $30,
			$31, $32, $33, $34, $35, $36, $37, $38, $39, $40, $41, $42, $43, $44,
			$45, $46, $47, $48, $49, $50, $51, $52, $53, $54, $55
		) RETURNING id`
	return db.QueryRow(query,
		v.DemographicsID, v.SymptomOnset, v.Fever, v.FeverOnsetDate, v.SoreThroat, v.SoreThroatOnsetDate,
		v.Headache, v.HeadacheOnsetDate, v.MuscleAches, v.MuscleAchesOnsetDate, v.Cough, v.CoughOnsetDate,
		v.Fatigue, v.FatigueOnsetDate, v.OralPain, v.OralPainOnsetDate, v.Nausea, v.NauseaOnsetDate,
		v.Vomiting, v.VomitingOnsetDate, v.Diarrhea, v.DiarrheaOnsetDate, v.RectalPain, v.RectalPainOnsetDate,
		v.Lesions, v.LesionsOnsetDate, v.Lymphadenopathy, v.LymphadenopathyOnsetDate, v.Pruritis, v.PruritisOnsetDate,
		v.PainSwallowing, v.PainSwallowingOnsetDate, v.DifficultySwallowing, v.DifficultySwallowingOnsetDate,
		v.Urethritis, v.UrethritisOnsetDate, v.ChestPain, v.ChestPainOnsetDate, v.DecreasedUrine, v.DecreasedUrineOnsetDate,
		v.Dizziness, v.DizzinessOnsetDate, v.JointPain, v.JointPainOnsetDate, v.PsychologicalDisturbance, v.PsychologicalDisturbanceOnsetDate,
		v.Temperature, v.HeartRate, v.RespiratoryRate, v.BpSystolic, v.BpDiastolic, v.Dehydration, v.AVPU, v.HeightCm, v.WeightKg).Scan(&v.ID)
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
			prothrombin_time, aptt, total_protein, albumin, bilirubin_d,
			lymphocytes, monocytes, eosinophils, basophils, neutrophils,
			hgb, hct, mcv, mch, mchc, rdw, rdw_sd, rdw_cv, mpv, pdw, pct,
			lab_other, lab_alt_notdone, lab_ast_notdone, lab_creatinine_notdone,
			lab_potassium_notdone, total_protein_notdone, albumin_notdone,
			lab_urea_notdone, lab_ck_notdone, lab_calcium_notdone, lab_sodium_notdone,
			lab_lymphocytes_notdone, lab_monocytes_notdone, lab_eosinophils_notdone,
			lab_basophils_notdone, lab_crp_notdone, lab_neutrophils_notdone,
			lab_hgb_notdone, lab_hct_notdone, lab_mcv_notdone, lab_mch_notdone,
			lab_mchc_notdone, lab_rdw_notdone, lab_rdw_sd_notdone, lab_rdw_cv_notdone,
			lab_mpv_notdone, lab_pdw_notdone, lab_pct_notdone, lab_other_notdone,
			lab_glucose_notdone, lab_lactate_notdone, lab_haemoglobin_notdone,
			lab_bilirubin_notdone, lab_bilirubin_d_notdone, lab_wbc_notdone,
			lab_platelets_notdone, lab_prothrombin_notdone, lab_aptt_notdone,
			other_malaria, other_hiv, other_syphilis, other_mpox, hepatitis_b, hepatitis_c,
			data_entrant_name, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15,
			$16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28, $29,
			$30, $31, $32, $33, $34, $35, $36, $37, $38, $39, $40, $41, $42, $43,
			$44, $45, $46, $47, $48, $49, $50, $51, $52, $53, $54, $55, $56, $57,
			$58, $59, $60, $61, $62, $63, $64, $65, $66, $67, $68, $69, $70, $71,
			$72, $73, $74, $75, $76, $77, $78, $79, $80, $81, $82, $83, $84, $85,
			CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
		) RETURNING id`
	return db.QueryRow(query,
		l.DemographicsID, l.ALT, l.AST, l.Creatinine, l.Potassium, l.Urea,
		l.CreatineKinase, l.Calcium, l.Sodium, l.CRP, l.Glucose, l.Lactate,
		l.Haemoglobin, l.TotalBilirubin, l.WBCCount, l.Platelets,
		l.ProthrombinTime, l.APTT, l.TotalProtein, l.Albumin, l.BilirubinD,
		l.Lymphocytes, l.Monocytes, l.Eosinophils, l.Basophils, l.Neutrophils,
		l.HGB, l.HCT, l.MCV, l.MCH, l.MCHC, l.RDW, l.RDWSD, l.RDWCV, l.MPV, l.PDW, l.PCT,
		l.LabOther, l.LabALTNotDone, l.LabASTNotDone, l.LabCreatinineNotDone,
		l.LabPotassiumNotDone, l.TotalProteinNotDone, l.AlbuminNotDone,
		l.LabUreaNotDone, l.LabCKNotDone, l.LabCalciumNotDone, l.LabSodiumNotDone,
		l.LabLymphocytesNotDone, l.LabMonocytesNotDone, l.LabEosinophilsNotDone,
		l.LabBasophilsNotDone, l.LabCRPNotDone, l.LabNeutrophilsNotDone,
		l.LabHGBNotDone, l.LabHCTNotDone, l.LabMCVNotDone, l.LabMCHNotDone,
		l.LabMCHCNotDone, l.LabRDWNotDone, l.LabRDWSDNotDone, l.LabRDWCVNotDone,
		l.LabMPVNotDone, l.LabPDWNotDone, l.LabPCTNotDone, l.LabOtherNotDone,
		l.LabGlucoseNotDone, l.LabLactateNotDone, l.LabHaemoglobinNotDone,
		l.LabBilirubinNotDone, l.LabBilirubinDNotDone, l.LabWBCNotDone,
		l.LabPlateletsNotDone, l.LabProthrombinNotDone, l.LabAPTTNotDone,
		l.OtherMalaria, l.OtherHIV, l.OtherSyphilis, l.OtherMpox, l.HepatitisB, l.HepatitisC,
		l.DataEntrantName).Scan(&l.ID)
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
