package models

import (
	"database/sql"
	"time"
)

// VHFPatient represents the patient information
type VHFPatient struct {
	ID                          int64           `json:"id"`
	Surname                     string          `json:"surname"`
	OtherNames                  string          `json:"other_names"`
	DateOfBirth                 sql.NullTime    `json:"date_of_birth"`
	AgeYears                    sql.NullInt32   `json:"age_years"`
	AgeMonths                   sql.NullInt32   `json:"age_months"`
	Gender                      string          `json:"gender"`
	PatientPhone                string          `json:"patient_phone"`
	PhoneOwner                  string          `json:"phone_owner"`
	NextOfKin                   string          `json:"next_of_kin"`
	NextOfKinPhone              string          `json:"next_of_kin_phone"`
	DataCapturerName            sql.NullString  `json:"data_capturer_name"`
	DataCapturerPhone           string          `json:"data_capturer_phone"`
	ReportingHealthFacilityName string          `json:"reporting_health_facility_name"`
	CaseCode                    string          `json:"case_code"`
	Status                      string          `json:"status"`
	DateOfDeath                 sql.NullTime    `json:"date_of_death"`
	HeadOfHousehold             string          `json:"head_of_household"`
	VillageTown                 string          `json:"village_town"`
	Parish                      string          `json:"parish"`
	Subcounty                   string          `json:"subcounty"`
	District                    string          `json:"district"`
	CountryOfResidence          string          `json:"country_of_residence"`
	Occupation                  string          `json:"occupation"`
	IllVillageTown              string          `json:"ill_village_town"`
	IllSubcounty                string          `json:"ill_subcounty"`
	IllDistrict                 string          `json:"ill_district"`
	Latitude                    sql.NullFloat64 `json:"latitude"`
	Longitude                   sql.NullFloat64 `json:"longitude"`
	DateResidingFrom            sql.NullTime    `json:"date_residing_from"`
	DateResidingTo              sql.NullTime    `json:"date_residing_to"`
	CreatedAt                   time.Time       `json:"created_at"`
}

// VHFClinicalSigns represents the clinical signs and symptoms
type VHFClinicalSigns struct {
	ID                                    int64           `json:"id"`
	PatientID                             int64           `json:"patient_id"`
	DateInitialOnset                      sql.NullTime    `json:"date_initial_onset"`
	TempSource                            string          `json:"temp_source"`
	Temperature                           sql.NullFloat64 `json:"temperature"`
	Fever                                 sql.NullBool    `json:"fever"`
	DateFever                             sql.NullTime    `json:"date_fever"`
	DurationFever                         sql.NullInt32   `json:"duration_fever"`
	Vomiting                              sql.NullBool    `json:"vomiting"`
	DateVomiting                          sql.NullTime    `json:"date_vomiting"`
	DurationVomiting                      sql.NullInt32   `json:"duration_vomiting"`
	Nausea                                sql.NullBool    `json:"nausea"`
	DateNausea                            sql.NullTime    `json:"date_nausea"`
	DurationNausea                        sql.NullInt32   `json:"duration_nausea"`
	Diarrhea                              sql.NullBool    `json:"diarrhea"`
	DateDiarrhea                          sql.NullTime    `json:"date_diarrhea"`
	DurationDiarrhea                      sql.NullInt32   `json:"duration_diarrhea"`
	IntenseFatigueGeneralWeakness         sql.NullBool    `json:"intense_fatigue_general_weakness"`
	DateIntenseFatigueGeneralWeakness     sql.NullTime    `json:"date_intense_fatigue_general_weakness"`
	DurationIntenseFatigueGeneralWeakness sql.NullInt32   `json:"duration_intense_fatigue_general_weakness"`
	EpigastricPain                        sql.NullBool    `json:"epigastric_pain"`
	DateEpigastricPain                    sql.NullTime    `json:"date_epigastric_pain"`
	DurationEpigastricPain                sql.NullInt32   `json:"duration_epigastric_pain"`
	LowerAbdominalPain                    sql.NullBool    `json:"lower_abdominal_pain"`
	DateLowerAbdominalPain                sql.NullTime    `json:"date_lower_abdominal_pain"`
	DurationLowerAbdominalPain            sql.NullInt32   `json:"duration_lower_abdominal_pain"`
	ChestPain                             sql.NullBool    `json:"chest_pain"`
	DateChestPain                         sql.NullTime    `json:"date_chest_pain"`
	DurationChestPain                     sql.NullInt32   `json:"duration_chest_pain"`
	MusclePain                            sql.NullBool    `json:"muscle_pain"`
	DateMusclePain                        sql.NullTime    `json:"date_muscle_pain"`
	DurationMusclePain                    sql.NullInt32   `json:"duration_muscle_pain"`
	JointPain                             sql.NullBool    `json:"joint_pain"`
	DateJointPain                         sql.NullTime    `json:"date_joint_pain"`
	DurationJointPain                     sql.NullInt32   `json:"duration_joint_pain"`
	Headache                              sql.NullBool    `json:"headache"`
	DateHeadache                          sql.NullTime    `json:"date_headache"`
	DurationHeadache                      sql.NullInt32   `json:"duration_headache"`
	Cough                                 sql.NullBool    `json:"cough"`
	DateCough                             sql.NullTime    `json:"date_cough"`
	DurationCough                         sql.NullInt32   `json:"duration_cough"`
	DifficultyBreathing                   sql.NullBool    `json:"difficulty_breathing"`
	DateDifficultyBreathing               sql.NullTime    `json:"date_difficulty_breathing"`
	DurationDifficultyBreathing           sql.NullInt32   `json:"duration_difficulty_breathing"`
	DifficultySwallowing                  sql.NullBool    `json:"difficulty_swallowing"`
	DateDifficultySwallowing              sql.NullTime    `json:"date_difficulty_swallowing"`
	DurationDifficultySwallowing          sql.NullInt32   `json:"duration_difficulty_swallowing"`
	SoreThroat                            sql.NullBool    `json:"sore_throat"`
	DateSoreThroat                        sql.NullTime    `json:"date_sore_throat"`
	DurationSoreThroat                    sql.NullInt32   `json:"duration_sore_throat"`
	Jaundice                              sql.NullBool    `json:"jaundice"`
	DateJaundice                          sql.NullTime    `json:"date_jaundice"`
	DurationJaundice                      sql.NullInt32   `json:"duration_jaundice"`
	Conjunctivitis                        sql.NullBool    `json:"conjunctivitis"`
	DateConjunctivitis                    sql.NullTime    `json:"date_conjunctivitis"`
	DurationConjunctivitis                sql.NullInt32   `json:"duration_conjunctivitis"`
	SkinRash                              sql.NullBool    `json:"skin_rash"`
	DateSkinRash                          sql.NullTime    `json:"date_skin_rash"`
	DurationSkinRash                      sql.NullInt32   `json:"duration_skin_rash"`
	Hiccups                               sql.NullBool    `json:"hiccups"`
	DateHiccups                           sql.NullTime    `json:"date_hiccups"`
	DurationHiccups                       sql.NullInt32   `json:"duration_hiccups"`
	PainBehindEyes                        sql.NullBool    `json:"pain_behind_eyes"`
	DatePainBehindEyes                    sql.NullTime    `json:"date_pain_behind_eyes"`
	DurationPainBehindEyes                sql.NullInt32   `json:"duration_pain_behind_eyes"`
	SensitiveToLight                      sql.NullBool    `json:"sensitive_to_light"`
	DateSensitiveToLight                  sql.NullTime    `json:"date_sensitive_to_light"`
	DurationSensitiveToLight              sql.NullInt32   `json:"duration_sensitive_to_light"`
	ComaUnconscious                       sql.NullBool    `json:"coma_unconscious"`
	DateComaUnconscious                   sql.NullTime    `json:"date_coma_unconscious"`
	DurationComaUnconscious               sql.NullInt32   `json:"duration_coma_unconscious"`
	ConfusedOrDisoriented                 sql.NullBool    `json:"confused_or_disoriented"`
	DateConfusedOrDisoriented             sql.NullTime    `json:"date_confused_or_disoriented"`
	DurationConfusedOrDisoriented         sql.NullInt32   `json:"duration_confused_or_disoriented"`
	Convulsions                           sql.NullBool    `json:"convulsions"`
	DateConvulsions                       sql.NullTime    `json:"date_convulsions"`
	DurationConvulsions                   sql.NullInt32   `json:"duration_convulsions"`
	UnexplainedBleeding                   sql.NullBool    `json:"unexplained_bleeding"`
	DateUnexplainedBleeding               sql.NullTime    `json:"date_unexplained_bleeding"`
	DurationUnexplainedBleeding           sql.NullInt32   `json:"duration_unexplained_bleeding"`
	BleedingOfTheGums                     sql.NullBool    `json:"bleeding_of_the_gums"`
	DateBleedingOfTheGums                 sql.NullTime    `json:"date_bleeding_of_the_gums"`
	DurationBleedingOfTheGums             sql.NullInt32   `json:"duration_bleeding_of_the_gums"`
	BleedingFromInjectionSite             sql.NullBool    `json:"bleeding_from_injection_site"`
	DateBleedingFromInjectionSite         sql.NullTime    `json:"date_bleeding_from_injection_site"`
	DurationBleedingFromInjectionSite     sql.NullInt32   `json:"duration_bleeding_from_injection_site"`
	NoseBleedEpistaxis                    sql.NullBool    `json:"nose_bleed_epistaxis"`
	DateNoseBleedEpistaxis                sql.NullTime    `json:"date_nose_bleed_epistaxis"`
	DurationNoseBleedEpistaxis            sql.NullInt32   `json:"duration_nose_bleed_epistaxis"`
	BloodyStool                           sql.NullBool    `json:"bloody_stool"`
	DateBloodyStool                       sql.NullTime    `json:"date_bloody_stool"`
	DurationBloodyStool                   sql.NullInt32   `json:"duration_bloody_stool"`
	BloodInVomit                          sql.NullBool    `json:"blood_in_vomit"`
	DateBloodInVomit                      sql.NullTime    `json:"date_blood_in_vomit"`
	DurationBloodInVomit                  sql.NullInt32   `json:"duration_blood_in_vomit"`
	CoughingUpBloodHemoptysis             sql.NullBool    `json:"coughing_up_blood_hemoptysis"`
	DateCoughingUpBloodHemoptysis         sql.NullTime    `json:"date_coughing_up_blood_hemoptysis"`
	DurationCoughingUpBloodHemoptysis     sql.NullInt32   `json:"duration_coughing_up_blood_hemoptysis"`
	BleedingFromVagina                    sql.NullBool    `json:"bleeding_from_vagina"`
	DateBleedingFromVagina                sql.NullTime    `json:"date_bleeding_from_vagina"`
	DurationBleedingFromVagina            sql.NullInt32   `json:"duration_bleeding_from_vagina"`
	BruisingOfTheSkin                     sql.NullBool    `json:"bruising_of_the_skin"`
	DateBruisingOfTheSkin                 sql.NullTime    `json:"date_bruising_of_the_skin"`
	DurationBruisingOfTheSkin             sql.NullInt32   `json:"duration_bruising_of_the_skin"`
	BloodInUrine                          sql.NullBool    `json:"blood_in_urine"`
	DateBloodInUrine                      sql.NullTime    `json:"date_blood_in_urine"`
	DurationBloodInUrine                  sql.NullInt32   `json:"duration_blood_in_urine"`
	OtherHemorrhagicSymptoms              sql.NullBool    `json:"other_hemorrhagic_symptoms"`
	DateOtherHemorrhagicSymptoms          sql.NullTime    `json:"date_other_hemorrhagic_symptoms"`
	DurationOtherHemorrhagicSymptoms      sql.NullInt32   `json:"duration_other_hemorrhagic_symptoms"`
	CreatedAt                             time.Time       `json:"created_at"`
}

// VHFHospitalization represents current hospitalization information
type VHFHospitalization struct {
	ID                 int64        `json:"id"`
	PatientID          int64        `json:"patient_id"`
	Hospitalized       bool         `json:"hospitalized"`
	AdmissionDate      sql.NullTime `json:"admission_date"`
	HealthFacilityName string       `json:"health_facility_name"`
	InIsolation        bool         `json:"in_isolation"`
	IsolationDate      sql.NullTime `json:"isolation_date"`
	CreatedAt          time.Time    `json:"created_at"`
}

// VHFPreviousHospitalization represents previous hospitalization records
type VHFPreviousHospitalization struct {
	ID                   int64     `json:"id"`
	PatientID            int64     `json:"patient_id"`
	HospitalizationDates string    `json:"hospitalization_dates"`
	HealthFacilityName   string    `json:"health_facility_name"`
	Village              string    `json:"village"`
	District             string    `json:"district"`
	WasIsolated          bool      `json:"was_isolated"`
	CreatedAt            time.Time `json:"created_at"`
}

// VHFRiskFactors represents risk factors and exposures
type VHFRiskFactors struct {
	ID               int64        `json:"id"`
	PatientID        int64        `json:"patient_id"`
	ContactWithCase  sql.NullBool `json:"contact_with_case"`
	ContactName      string       `json:"contact_name"`
	ContactRelation  string       `json:"contact_relation"`
	ContactDates     string       `json:"contact_dates"`
	ContactVillage   string       `json:"contact_village"`
	ContactDistrict  string       `json:"contact_district"`
	ContactStatus    string       `json:"contact_status"`
	ContactDeathDate sql.NullTime `json:"contact_death_date"`
	ContactTypes     string       `json:"contact_types"`
	// ... Add all other risk factor fields
	CreatedAt time.Time `json:"created_at"`
}

// VHFLaboratory represents laboratory testing information
type VHFLaboratory struct {
	ID                   int64          `json:"id"`
	PatientID            int64          `json:"patient_id"`
	SampleCollectionDate sql.NullTime   `json:"sample_collection_date"`
	SampleCollectionTime sql.NullString `json:"sample_collection_time"`
	SampleType           string         `json:"sample_type"`
	OtherSampleType      string         `json:"other_sample_type"`
	RequestedTest        string         `json:"requested_test"`
	Serology             string         `json:"serology"`
	MalariaRDT           string         `json:"malaria_rdt"`
	HIVRDT               string         `json:"hiv_rdt"`
	CreatedAt            time.Time      `json:"created_at"`
}

// VHFInvestigator represents case investigator information
type VHFInvestigator struct {
	ID                int64     `json:"id"`
	PatientID         int64     `json:"patient_id"`
	InvestigatorName  string    `json:"investigator_name"`
	Phone             string    `json:"phone"`
	Email             string    `json:"email"`
	Position          string    `json:"position"`
	District          string    `json:"district"`
	HealthFacility    string    `json:"health_facility"`
	InformationSource string    `json:"information_source"`
	ProxyName         string    `json:"proxy_name"`
	ProxyRelation     string    `json:"proxy_relation"`
	CreatedAt         time.Time `json:"created_at"`
}

// SaveVHFPatient saves a new patient record
func SaveVHFPatient(db *sql.DB, patient *VHFPatient) error {
	query := `
		INSERT INTO vhf_patients (
			surname, other_names, date_of_birth, age_years, age_months,
			gender, patient_phone, phone_owner, next_of_kin, next_of_kin_phone,
			data_capturer_name, data_capturer_phone, reporting_health_facility_name, case_code, status, date_of_death,
			head_of_household, village_town, parish, subcounty, district,
			country_of_residence, occupation, ill_village_town, ill_subcounty, ill_district,
			latitude, longitude, date_residing_from, date_residing_to
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15,
			$16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28, $29, $30
		) RETURNING id, created_at`

	err := db.QueryRow(
		query,
		patient.Surname, patient.OtherNames, patient.DateOfBirth, patient.AgeYears,
		patient.AgeMonths, patient.Gender, patient.PatientPhone, patient.PhoneOwner,
		patient.NextOfKin, patient.NextOfKinPhone, patient.DataCapturerName,
		patient.DataCapturerPhone, patient.ReportingHealthFacilityName, patient.CaseCode, patient.Status, patient.DateOfDeath,
		patient.HeadOfHousehold, patient.VillageTown, patient.Parish, patient.Subcounty,
		patient.District, patient.CountryOfResidence, patient.Occupation,
		patient.IllVillageTown, patient.IllSubcounty, patient.IllDistrict,
		patient.Latitude, patient.Longitude, patient.DateResidingFrom, patient.DateResidingTo,
	).Scan(&patient.ID, &patient.CreatedAt)

	return err
}

// GetVHFPatient retrieves a patient record by ID
func GetVHFPatient(db *sql.DB, id int64) (*VHFPatient, error) {
	patient := &VHFPatient{}
	query := `
		SELECT id, surname, other_names, date_of_birth, age_years, age_months,
			gender, patient_phone, phone_owner, next_of_kin, next_of_kin_phone,
			data_capturer_name, data_capturer_phone, reporting_health_facility_name, case_code, status, date_of_death,
			head_of_household, village_town, parish, subcounty, district,
			country_of_residence, occupation, ill_village_town, ill_subcounty, ill_district,
			latitude, longitude, date_residing_from, date_residing_to, created_at
		FROM vhf_patients
		WHERE id = $1`

	err := db.QueryRow(query, id).Scan(
		&patient.ID, &patient.Surname, &patient.OtherNames, &patient.DateOfBirth,
		&patient.AgeYears, &patient.AgeMonths, &patient.Gender, &patient.PatientPhone,
		&patient.PhoneOwner, &patient.NextOfKin, &patient.NextOfKinPhone,
		&patient.DataCapturerName, &patient.DataCapturerPhone, &patient.ReportingHealthFacilityName, &patient.CaseCode,
		&patient.Status, &patient.DateOfDeath, &patient.HeadOfHousehold,
		&patient.VillageTown, &patient.Parish, &patient.Subcounty, &patient.District,
		&patient.CountryOfResidence, &patient.Occupation, &patient.IllVillageTown,
		&patient.IllSubcounty, &patient.IllDistrict, &patient.Latitude,
		&patient.Longitude, &patient.DateResidingFrom, &patient.DateResidingTo,
		&patient.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return patient, nil
}

// ListVHFPatients retrieves all patient records
func ListVHFPatients(db *sql.DB) ([]*VHFPatient, error) {
	query := `
		SELECT id, surname, other_names, date_of_birth, age_years, age_months,
			gender, patient_phone, phone_owner, next_of_kin, next_of_kin_phone,
			data_capturer_name, data_capturer_phone, reporting_health_facility_name, case_code, status, date_of_death,
			head_of_household, village_town, parish, subcounty, district,
			country_of_residence, occupation, ill_village_town, ill_subcounty, ill_district,
			latitude, longitude, date_residing_from, date_residing_to, created_at
		FROM vhf_patients
		ORDER BY created_at DESC`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var patients []*VHFPatient
	for rows.Next() {
		patient := &VHFPatient{}
		err := rows.Scan(
			&patient.ID, &patient.Surname, &patient.OtherNames, &patient.DateOfBirth,
			&patient.AgeYears, &patient.AgeMonths, &patient.Gender, &patient.PatientPhone,
			&patient.PhoneOwner, &patient.NextOfKin, &patient.NextOfKinPhone,
			&patient.DataCapturerName, &patient.DataCapturerPhone, &patient.ReportingHealthFacilityName, &patient.CaseCode,
			&patient.Status, &patient.DateOfDeath, &patient.HeadOfHousehold,
			&patient.VillageTown, &patient.Parish, &patient.Subcounty, &patient.District,
			&patient.CountryOfResidence, &patient.Occupation, &patient.IllVillageTown,
			&patient.IllSubcounty, &patient.IllDistrict, &patient.Latitude,
			&patient.Longitude, &patient.DateResidingFrom, &patient.DateResidingTo,
			&patient.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		patients = append(patients, patient)
	}

	return patients, nil
}

// Get VHF Patient by Case Code
func GetVHFPatientByCaseCode(db *sql.DB, caseCode string) (*VHFPatient, error) {
	patient := &VHFPatient{}
	query := `
		SELECT id, surname, other_names, date_of_birth, age_years, age_months,
			gender, patient_phone, phone_owner, next_of_kin, next_of_kin_phone,
			data_capturer_name, data_capturer_phone, reporting_health_facility_name, case_code, status, date_of_death,
			head_of_household, village_town, parish, subcounty, district,
			country_of_residence, occupation, ill_village_town, ill_subcounty, ill_district,
			latitude, longitude, date_residing_from, date_residing_to, created_at
		FROM vhf_patients
		WHERE case_code = $1`

	err := db.QueryRow(query, caseCode).Scan(
		&patient.ID, &patient.Surname, &patient.OtherNames, &patient.DateOfBirth,
		&patient.AgeYears, &patient.AgeMonths, &patient.Gender, &patient.PatientPhone,
		&patient.PhoneOwner, &patient.NextOfKin, &patient.NextOfKinPhone,
		&patient.DataCapturerName, &patient.DataCapturerPhone, &patient.ReportingHealthFacilityName, &patient.CaseCode,
		&patient.Status, &patient.DateOfDeath, &patient.HeadOfHousehold,
		&patient.VillageTown, &patient.Parish, &patient.Subcounty, &patient.District,
		&patient.CountryOfResidence, &patient.Occupation, &patient.IllVillageTown,
		&patient.IllSubcounty, &patient.IllDistrict, &patient.Latitude,
		&patient.Longitude, &patient.DateResidingFrom, &patient.DateResidingTo,
		&patient.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return patient, nil
}

// SaveVHFInvestigator saves a new investigator record
func SaveVHFInvestigator(db *sql.DB, investigator *VHFInvestigator) error {
	query := `
		INSERT INTO vhf_investigator (
			patient_id, investigator_name, phone, email, position,
			district, health_facility, information_source, proxy_name,
			proxy_relation, created_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
		) RETURNING id, created_at`

	now := time.Now()
	investigator.CreatedAt = now

	err := db.QueryRow(
		query,
		investigator.PatientID,
		investigator.InvestigatorName,
		investigator.Phone,
		investigator.Email,
		investigator.Position,
		investigator.District,
		investigator.HealthFacility,
		investigator.InformationSource,
		investigator.ProxyName,
		investigator.ProxyRelation,
		investigator.CreatedAt,
	).Scan(&investigator.ID, &investigator.CreatedAt)

	return err
}

// SaveVHFClinicalSigns saves a new clinical signs record
func SaveVHFClinicalSigns(db *sql.DB, signs *VHFClinicalSigns) error {
	query := `
		INSERT INTO vhf_clinical_signs (
			patient_id, date_initial_onset, temp_source, temperature, fever, vomiting, nausea, diarrhea,
			intense_fatigue_general_weakness, epigastric_pain, lower_abdominal_pain,
			chest_pain, muscle_pain, joint_pain, headache, cough, difficulty_breathing,
			difficulty_swallowing, sore_throat, jaundice, conjunctivitis, skin_rash,
			hiccups, pain_behind_eyes, sensitive_to_light, coma_unconscious,
			confused_or_disoriented, convulsions, unexplained_bleeding,
			bleeding_of_the_gums, bleeding_from_injection_site, nose_bleed_epistaxis,
			bloody_stool, blood_in_vomit, coughing_up_blood_hemoptysis,
			bleeding_from_vagina, bruising_of_the_skin, blood_in_urine,
			other_hemorrhagic_symptoms, created_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15,
			$16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28,
			$29, $30, $31, $32, $33, $34, $35, $36, $37, $38, $39, $40
		) RETURNING id, created_at`

	now := time.Now()
	signs.CreatedAt = now

	err := db.QueryRow(
		query,
		signs.PatientID, signs.DateInitialOnset, signs.TempSource, signs.Temperature,
		signs.Fever, signs.Vomiting, signs.Nausea, signs.Diarrhea,
		signs.IntenseFatigueGeneralWeakness, signs.EpigastricPain, signs.LowerAbdominalPain,
		signs.ChestPain, signs.MusclePain, signs.JointPain, signs.Headache, signs.Cough,
		signs.DifficultyBreathing, signs.DifficultySwallowing, signs.SoreThroat,
		signs.Jaundice, signs.Conjunctivitis, signs.SkinRash, signs.Hiccups,
		signs.PainBehindEyes, signs.SensitiveToLight, signs.ComaUnconscious,
		signs.ConfusedOrDisoriented, signs.Convulsions, signs.UnexplainedBleeding,
		signs.BleedingOfTheGums, signs.BleedingFromInjectionSite,
		signs.NoseBleedEpistaxis, signs.BloodyStool, signs.BloodInVomit,
		signs.CoughingUpBloodHemoptysis, signs.BleedingFromVagina,
		signs.BruisingOfTheSkin, signs.BloodInUrine, signs.OtherHemorrhagicSymptoms,
		signs.CreatedAt,
	).Scan(&signs.ID, &signs.CreatedAt)

	return err
}

// SaveVHFHospitalization saves a new hospitalization record
func SaveVHFHospitalization(db *sql.DB, hospitalization *VHFHospitalization) error {
	query := `
		INSERT INTO vhf_hospitalization (
			patient_id, hospitalized, admission_date, health_facility_name,
			in_isolation, isolation_date
		) VALUES (
			$1, $2, $3, $4, $5, $6
		) RETURNING id, created_at`

	err := db.QueryRow(
		query,
		hospitalization.PatientID, hospitalization.Hospitalized,
		hospitalization.AdmissionDate, hospitalization.HealthFacilityName,
		hospitalization.InIsolation, hospitalization.IsolationDate,
	).Scan(&hospitalization.ID, &hospitalization.CreatedAt)

	return err
}

// SaveVHFRiskFactors saves a new risk factors record
func SaveVHFRiskFactors(db *sql.DB, riskFactors *VHFRiskFactors) error {
	query := `
		INSERT INTO vhf_risk_factors (
			patient_id, contact_with_case, contact_name, contact_relation,
			contact_dates, contact_village, contact_district, contact_status,
			contact_death_date, contact_types
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10
		) RETURNING id, created_at`

	err := db.QueryRow(
		query,
		riskFactors.PatientID, riskFactors.ContactWithCase,
		riskFactors.ContactName, riskFactors.ContactRelation,
		riskFactors.ContactDates, riskFactors.ContactVillage,
		riskFactors.ContactDistrict, riskFactors.ContactStatus,
		riskFactors.ContactDeathDate, riskFactors.ContactTypes,
	).Scan(&riskFactors.ID, &riskFactors.CreatedAt)

	return err
}

// SaveVHFLaboratory saves a new laboratory record
func SaveVHFLaboratory(db *sql.DB, laboratory *VHFLaboratory) error {
	query := `
		INSERT INTO vhf_laboratory (
			patient_id, sample_collection_date, sample_collection_time,
			sample_type, other_sample_type, requested_test, serology,
			malaria_rdt, hiv_rdt
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9
		) RETURNING id, created_at`

	err := db.QueryRow(
		query,
		laboratory.PatientID, laboratory.SampleCollectionDate,
		laboratory.SampleCollectionTime, laboratory.SampleType,
		laboratory.OtherSampleType, laboratory.RequestedTest,
		laboratory.Serology, laboratory.MalariaRDT, laboratory.HIVRDT,
	).Scan(&laboratory.ID, &laboratory.CreatedAt)

	return err
}

// Similar Save, Get, and List functions for other models...
// I'll continue with the handlers in the next message
