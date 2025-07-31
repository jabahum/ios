package models

import (
	"database/sql"
	"time"
)

// PolioCaseInvestigation represents the main case investigation record
type PolioCaseInvestigation struct {
	ID             int64
	CaseID         string
	EpidNumber     string
	Country        string
	RegionProvince string
	District       string
	YearOnset      int
	CaseNumber     int
	ReceivedDate   time.Time
	CreatedAt      time.Time
}

// PolioIdentification represents patient identification information
type PolioIdentification struct {
	ID                    int64
	CaseID                string
	District              string
	RegionProvince        string
	Address               string
	Village               string
	City                  string
	NearestHealthFacility string
	Longitude             sql.NullFloat64
	Latitude              sql.NullFloat64
	PatientName           string
	FatherMother          string
	PhoneNumber           string
	DateOfBirth           sql.NullTime
	AgeYears              sql.NullInt32
	AgeMonths             sql.NullInt32
	Sex                   string
}

// PolioNotificationInvestigation represents notification and investigation details
type PolioNotificationInvestigation struct {
	ID                  int64
	CaseID              string
	NotifiedBy          string
	DateOfNotification  sql.NullTime
	DateOfInvestigation sql.NullTime
}

// PolioHospitalization represents hospitalization information
type PolioHospitalization struct {
	ID                   int64
	CaseID               string
	Hospitalized         bool
	DateOfAdmission      sql.NullTime
	HospitalRecordNumber string
	HospitalNameAddress  string
}

// PolioClinicalHistory represents clinical history and symptoms
type PolioClinicalHistory struct {
	ID                     int64
	CaseID                 string
	FeverAtOnset           sql.NullString
	DateOnsetOfFever       sql.NullTime
	ProgressiveParalysis   sql.NullString
	DateOnsetOfParalysis   sql.NullTime
	FlaccidAcuteParalysis  sql.NullString
	SensationLoss          sql.NullString
	SuddenOnset            bool
	Asymmetric             bool
	LeftArmParalysis       bool
	RightArmParalysis      bool
	LeftLegParalysis       bool
	RightLegParalysis      bool
	DiminishedReflexes     bool
	DiminishedMuscleTone   bool
	MuscleWasting          bool
	MuscleWeakness         bool
	RespiratoryMuscles     bool
	Face                   bool
	StiffNeck              bool
	Convulsions            bool
	Headache               bool
	Vomiting               bool
	Diarrhoea              bool
	OtherSites             string
	RecentInjection        sql.NullString
	TotalInjections        sql.NullInt32
	InjectionDates         []sql.NullTime
	InjectionType          string
	InjectionSiteTable     string // JSON string for injection site table
	ParalyzedLimbSensitive sql.NullString
	InjectionFacilityName  string
	ProvisionalDiagnosis   string
	TrueAFPCase            sql.NullString
}

// PolioImmunizationHistory represents immunization history
type PolioImmunizationHistory struct {
	ID                     int64
	CaseID                 string
	TotalPolioDoses        sql.NullInt32
	ExcludeDoseAtBirth     bool
	OPVDoseAtBirth         sql.NullTime
	OPVDose1               sql.NullTime
	OPVDose2               sql.NullTime
	OPVDose3               sql.NullTime
	OPVDose4               sql.NullTime
	OPVDoseMoreThan4       sql.NullTime
	LastOPVDose            sql.NullTime
	TotalOPVSIA            sql.NullInt32
	LastOPVSIA             sql.NullTime
	TotalOPVRI             sql.NullInt32
	TotalIPVSIA            sql.NullInt32
	TotalIPVRI             sql.NullInt32
	LastIPVSIA             sql.NullTime
	SourceOfRIVaccination  string // "Card" or "Recall"
	UnknownZeroDoseReasons string
}

// PolioStoolSpecimenCollection represents stool specimen collection details
type PolioStoolSpecimenCollection struct {
	ID                           int64
	CaseID                       string
	DateFirstSpecimen            sql.NullTime
	DateSecondSpecimen           sql.NullTime
	DateSpecimenSentNational     sql.NullTime
	DateSpecimenReceivedNational sql.NullTime
	DateSpecimenSentLab          sql.NullTime
}

// PolioStoolSpecimenResults represents stool specimen results
type PolioStoolSpecimenResults struct {
	ID                            int64
	CaseID                        string
	DateReceivedAtLab             sql.NullTime
	SpecimenStatusAtReception     sql.NullString // "Adequate" or "Not adequate"
	DateCombinedCellCulture       sql.NullTime
	DateResultsSentToEPI          sql.NullTime
	DateResultsReceivedAtEPI      sql.NullTime
	FinalCellCultureResults       sql.NullString // "Suspected poliovirus", "Negative", "NPENT", "Suspect poliovirus + NPENT"
	W1                            sql.NullString
	W2                            sql.NullString
	W3                            sql.NullString
	DiscordantSabin               sql.NullString
	SL1                           sql.NullString
	SL2                           sql.NullString
	SL3                           sql.NullString
	RNPENT                        sql.NullString
	NEV                           sql.NullString
	DateSentToRegionalLab         sql.NullTime
	DateITDifferentiationSent     sql.NullTime
	DateITDifferentiationReceived sql.NullTime
	DateIsolateSentSequencing     sql.NullTime
	DateSeqResultsSentProgram     sql.NullTime
}

// PolioFollowUpExamination represents follow-up examination details
type PolioFollowUpExamination struct {
	ID                      int64
	CaseID                  string
	DateOfFollowUp          sql.NullTime
	ResidualParalysisLA     bool
	ResidualParalysisRA     bool
	ResidualParalysisLL     bool
	ResidualParalysisRL     bool
	ResultsOfExam           sql.NullString // "Residual Flaccid Paralysis", "No residual paralysis", "Lost follow-up", "Died before follow-up", "Residual Spastic Paralysis"
	ImmunocompromisedStatus sql.NullString
	FinalClassification     sql.NullString // "Confirmed Polio", "Compatible", "Discarded", "Not an AFP case"
	CVDPV                   bool
	AVDPV                   bool
	IVDPV                   bool
	Serotype                sql.NullString // "1", "2", "3"
}

// PolioPatientHistory represents patient visit history
type PolioPatientHistory struct {
	ID              int64
	CaseID          string
	Place1          string
	Duration1Months sql.NullInt32
	Duration1Days   sql.NullInt32
	Place2          string
	Duration2Months sql.NullInt32
	Duration2Days   sql.NullInt32
	Place3          string
	Duration3Months sql.NullInt32
	Duration3Days   sql.NullInt32
	Place4          string
	Duration4Months sql.NullInt32
	Duration4Days   sql.NullInt32
}

// PolioInvestigator represents investigator information
type PolioInvestigator struct {
	ID                int64
	CaseID            string
	InvestigatorName  string
	InvestigatorTitle string
	Unit              string
	Address           string
	Telephone         string
}

// Insert methods for all models
func (p *PolioCaseInvestigation) Insert(db *sql.DB) error {
	query := `INSERT INTO polio_case_investigation (case_id, epid_number, country, region_province, district, year_onset, case_number, received_date, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	_, err := db.Exec(query, p.CaseID, p.EpidNumber, p.Country, p.RegionProvince, p.District, p.YearOnset, p.CaseNumber, p.ReceivedDate, p.CreatedAt)
	return err
}

func (p *PolioIdentification) Insert(db *sql.DB) error {
	query := `INSERT INTO polio_identification (case_id, district, region_province, address, village, city, nearest_health_facility, longitude, latitude, patient_name, father_mother, phone_number, date_of_birth, age_years, age_months, sex) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)`
	_, err := db.Exec(query, p.CaseID, p.District, p.RegionProvince, p.Address, p.Village, p.City, p.NearestHealthFacility, p.Longitude, p.Latitude, p.PatientName, p.FatherMother, p.PhoneNumber, p.DateOfBirth, p.AgeYears, p.AgeMonths, p.Sex)
	return err
}

func (p *PolioNotificationInvestigation) Insert(db *sql.DB) error {
	query := `INSERT INTO polio_notification_investigation (case_id, notified_by, date_of_notification, date_of_investigation) VALUES ($1, $2, $3, $4)`
	_, err := db.Exec(query, p.CaseID, p.NotifiedBy, p.DateOfNotification, p.DateOfInvestigation)
	return err
}

func (p *PolioHospitalization) Insert(db *sql.DB) error {
	query := `INSERT INTO polio_hospitalization (case_id, hospitalized, date_of_admission, hospital_record_number, hospital_name_address) VALUES ($1, $2, $3, $4, $5)`
	_, err := db.Exec(query, p.CaseID, p.Hospitalized, p.DateOfAdmission, p.HospitalRecordNumber, p.HospitalNameAddress)
	return err
}

func (p *PolioClinicalHistory) Insert(db *sql.DB) error {
	query := `INSERT INTO polio_clinical_history (case_id, fever_at_onset, date_onset_of_fever, progressive_paralysis, date_onset_of_paralysis, flaccid_acute_paralysis, sensation_loss, sudden_onset, asymmetric, left_arm_paralysis, right_arm_paralysis, left_leg_paralysis, right_leg_paralysis, diminished_reflexes, diminished_muscle_tone, muscle_wasting, muscle_weakness, respiratory_muscles, face, stiff_neck, convulsions, headache, vomiting, diarrhoea, other_sites, recent_injection, total_injections, injection_type, paralyzed_limb_sensitive, injection_facility_name, provisional_diagnosis, true_afp_case) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28, $29, $30, $31, $32)`
	_, err := db.Exec(query, p.CaseID, p.FeverAtOnset, p.DateOnsetOfFever, p.ProgressiveParalysis, p.DateOnsetOfParalysis, p.FlaccidAcuteParalysis, p.SensationLoss, p.SuddenOnset, p.Asymmetric, p.LeftArmParalysis, p.RightArmParalysis, p.LeftLegParalysis, p.RightLegParalysis, p.DiminishedReflexes, p.DiminishedMuscleTone, p.MuscleWasting, p.MuscleWeakness, p.RespiratoryMuscles, p.Face, p.StiffNeck, p.Convulsions, p.Headache, p.Vomiting, p.Diarrhoea, p.OtherSites, p.RecentInjection, p.TotalInjections, p.InjectionType, p.ParalyzedLimbSensitive, p.InjectionFacilityName, p.ProvisionalDiagnosis, p.TrueAFPCase)
	return err
}

func (p *PolioImmunizationHistory) Insert(db *sql.DB) error {
	query := `INSERT INTO polio_immunization_history (case_id, total_polio_doses, exclude_dose_at_birth, opv_dose_at_birth, opv_dose1, opv_dose2, opv_dose3, opv_dose4, opv_dose_more_than4, last_opv_dose, total_opv_sia, last_opv_sia, total_opv_ri, total_ipv_sia, total_ipv_ri, last_ipv_sia, source_of_ri_vaccination, unknown_zero_dose_reasons) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)`
	_, err := db.Exec(query, p.CaseID, p.TotalPolioDoses, p.ExcludeDoseAtBirth, p.OPVDoseAtBirth, p.OPVDose1, p.OPVDose2, p.OPVDose3, p.OPVDose4, p.OPVDoseMoreThan4, p.LastOPVDose, p.TotalOPVSIA, p.LastOPVSIA, p.TotalOPVRI, p.TotalIPVSIA, p.TotalIPVRI, p.LastIPVSIA, p.SourceOfRIVaccination, p.UnknownZeroDoseReasons)
	return err
}

func (p *PolioStoolSpecimenCollection) Insert(db *sql.DB) error {
	query := `INSERT INTO polio_stool_specimen_collection (case_id, date_first_specimen, date_second_specimen, date_specimen_sent_national, date_specimen_received_national, date_specimen_sent_lab) VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := db.Exec(query, p.CaseID, p.DateFirstSpecimen, p.DateSecondSpecimen, p.DateSpecimenSentNational, p.DateSpecimenReceivedNational, p.DateSpecimenSentLab)
	return err
}

func (p *PolioStoolSpecimenResults) Insert(db *sql.DB) error {
	query := `INSERT INTO polio_stool_specimen_results (case_id, date_received_at_lab, specimen_status_at_reception, date_combined_cell_culture, date_results_sent_to_epi, date_results_received_at_epi, final_cell_culture_results, w1, w2, w3, discordant_sabin, sl1, sl2, sl3, r_npent, nev, date_sent_to_regional_lab, date_it_differentiation_sent, date_it_differentiation_received, date_isolate_sent_sequencing, date_seq_results_sent_program) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21)`
	_, err := db.Exec(query, p.CaseID, p.DateReceivedAtLab, p.SpecimenStatusAtReception, p.DateCombinedCellCulture, p.DateResultsSentToEPI, p.DateResultsReceivedAtEPI, p.FinalCellCultureResults, p.W1, p.W2, p.W3, p.DiscordantSabin, p.SL1, p.SL2, p.SL3, p.RNPENT, p.NEV, p.DateSentToRegionalLab, p.DateITDifferentiationSent, p.DateITDifferentiationReceived, p.DateIsolateSentSequencing, p.DateSeqResultsSentProgram)
	return err
}

func (p *PolioFollowUpExamination) Insert(db *sql.DB) error {
	query := `INSERT INTO polio_follow_up_examination (case_id, date_of_follow_up, residual_paralysis_la, residual_paralysis_ra, residual_paralysis_ll, residual_paralysis_rl, results_of_exam, immunocompromised_status, final_classification, cvdpv, avdpv, ivdpv, serotype) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`
	_, err := db.Exec(query, p.CaseID, p.DateOfFollowUp, p.ResidualParalysisLA, p.ResidualParalysisRA, p.ResidualParalysisLL, p.ResidualParalysisRL, p.ResultsOfExam, p.ImmunocompromisedStatus, p.FinalClassification, p.CVDPV, p.AVDPV, p.IVDPV, p.Serotype)
	return err
}

func (p *PolioPatientHistory) Insert(db *sql.DB) error {
	query := `INSERT INTO polio_patient_history (case_id, place1, duration1_months, duration1_days, place2, duration2_months, duration2_days, place3, duration3_months, duration3_days, place4, duration4_months, duration4_days) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`
	_, err := db.Exec(query, p.CaseID, p.Place1, p.Duration1Months, p.Duration1Days, p.Place2, p.Duration2Months, p.Duration2Days, p.Place3, p.Duration3Months, p.Duration3Days, p.Place4, p.Duration4Months, p.Duration4Days)
	return err
}

func (p *PolioInvestigator) Insert(db *sql.DB) error {
	query := `INSERT INTO polio_investigator (case_id, investigator_name, investigator_title, unit, address, telephone) VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := db.Exec(query, p.CaseID, p.InvestigatorName, p.InvestigatorTitle, p.Unit, p.Address, p.Telephone)
	return err
}
