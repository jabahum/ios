package models

import (
	"database/sql"
	"time"

	"github.com/lib/pq"
)

// MpoxCaseInvestigation represents the main case investigation record
type MpoxCaseInvestigation struct {
	ID                 int64
	CaseID             string
	Date               time.Time
	CaseStatus         sql.NullString
	CaseClassification sql.NullString
}

// MpoxPatientDemographics represents patient demographic information
type MpoxPatientDemographics struct {
	ID                   int64
	CaseID               string
	HealthFacilityCaseID sql.NullString
	Surname              string
	OtherNames           sql.NullString
	Sex                  string
	DateOfBirth          time.Time
	Age                  int
	Parish               sql.NullString
	SubCounty            sql.NullString
	PhysicalAddress      sql.NullString
	ContactTelephone     sql.NullString
	Occupation           sql.NullString
	Nationality          sql.NullString
	VaccinationStatus    sql.NullString
	DateOfVaccination    sql.NullTime
	NextOfKin            sql.NullString
	NextOfKinContact     sql.NullString
	MaritalStatus        sql.NullString
	IfDeadDateOfDeath    sql.NullTime
	AdmissionDate        sql.NullTime
	OnsetDate            sql.NullTime
	RashOnsetDate        sql.NullTime
}

// MpoxClinicianInfo represents clinician information
type MpoxClinicianInfo struct {
	ID               int64
	CaseID           string
	ClinicianName    sql.NullString
	ClinicianContact sql.NullString
	FacilityName     sql.NullString
	ClinicianEmail   sql.NullString
	FacilityDistrict sql.NullString
	PDPIDNumber      sql.NullString
	AdmissionDate    sql.NullTime
	Ward             sql.NullString
}

// MpoxCaseExposureHistory represents exposure history information
type MpoxCaseExposureHistory struct {
	ID                          int64
	CaseID                      string
	TraveledCountryReportedMpox sql.NullString
	CloseContactMpox            sql.NullString
	IntlTravel                  sql.NullString
	ContactAnimals              sql.NullString
	DomesticWildAnimals         sql.NullString
	SexualExposure              sql.NullString
}

// MpoxClinicalManifestations represents clinical symptoms and manifestations
type MpoxClinicalManifestations struct {
	ID                       int64
	CaseID                   string
	OnsetDate                sql.NullTime
	Fever                    sql.NullString
	FeverTemperature         sql.NullString
	Lymphadenopathy          sql.NullString
	Symptoms                 pq.StringArray
	SymptomOtherSpecify      sql.NullString
	NauseaVomiting           sql.NullString
	Pregnant                 sql.NullString
	PregnantTrimester        sql.NullString
	Vaccinated               sql.NullString
	VaccinationDate          sql.NullString
	Rash                     sql.NullString
	RashOnsetDate            sql.NullTime
	RashDistribution         pq.StringArray
	RashType                 pq.StringArray
	UnderlyingIllness        sql.NullString
	UnderlyingIllnessDetails sql.NullString
}

// MpoxTravelHistory represents travel-related data
type MpoxTravelHistory struct {
	ID                  int64
	CaseID              string
	TravelOutsideUganda sql.NullString
	CountryVisited      pq.StringArray
	LocationVisited     pq.StringArray
	DateArrival         pq.StringArray
	DateDeparture       pq.StringArray
	ActivitiesLocation  pq.StringArray
}

// MpoxLabInvestigation represents lab investigation details
type MpoxLabInvestigation struct {
	ID                      int64
	CaseID                  string
	LabID                   sql.NullString
	SampleCollected         pq.StringArray
	SampleOtherSpecify      sql.NullString
	TestRequested           pq.StringArray
	TestOtherSpecify        sql.NullString
	DateSampleCollection    sql.NullTime
	TimeSampleCollection    sql.NullTime
	DateSampleDispatch      sql.NullTime
	SampleCollectorName     sql.NullString
	SampleCollectorPhone    sql.NullString
	DateSampleReception     sql.NullTime
	TimeSampleReception     sql.NullTime
	SampleRecipientName     sql.NullString
	SampleRecipientPhone    sql.NullString
	GenomicCharacterization sql.NullString
	Clade                   pq.StringArray
	AccessionNumber         sql.NullString
}

// Insert methods for each model
func (m *MpoxCaseInvestigation) Insert(tx *sql.Tx) error {
	query := `
		INSERT INTO mpox_case_investigation (case_id, date, case_status, case_classification)
		VALUES ($1, $2, $3, $4)
		RETURNING id`
	return tx.QueryRow(query, m.CaseID, m.Date, m.CaseStatus, m.CaseClassification).Scan(&m.ID)
}

func (m *MpoxPatientDemographics) Insert(tx *sql.Tx) error {
	query := `
		INSERT INTO mpox_patient_demographics (
			case_id, health_facility_case_id, surname, other_names, sex, date_of_birth,
			age, parish, sub_county, physical_address, contact_telephone, occupation,
			nationality, vaccination_status, date_of_vaccination, next_of_kin,
			next_of_kin_contact, marital_status, if_dead_date_of_death,
			admission_date, onset_date, rash_onset_date
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16,
			$17, $18, $19, $20, $21, $22
		)
		RETURNING id`
	return tx.QueryRow(query,
		m.CaseID, m.HealthFacilityCaseID, m.Surname, m.OtherNames, m.Sex, m.DateOfBirth,
		m.Age, m.Parish, m.SubCounty, m.PhysicalAddress, m.ContactTelephone, m.Occupation,
		m.Nationality, m.VaccinationStatus, m.DateOfVaccination, m.NextOfKin,
		m.NextOfKinContact, m.MaritalStatus, m.IfDeadDateOfDeath,
		m.AdmissionDate, m.OnsetDate, m.RashOnsetDate,
	).Scan(&m.ID)
}

func (m *MpoxClinicianInfo) Insert(tx *sql.Tx) error {
	query := `
		INSERT INTO mpox_clinician_info (
			case_id, clinician_name, clinician_contact, facility_name,
			clinician_email, facility_district, pdpid_number, admission_date, ward
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id`
	return tx.QueryRow(query,
		m.CaseID, m.ClinicianName, m.ClinicianContact, m.FacilityName,
		m.ClinicianEmail, m.FacilityDistrict, m.PDPIDNumber, m.AdmissionDate, m.Ward,
	).Scan(&m.ID)
}

func (m *MpoxCaseExposureHistory) Insert(tx *sql.Tx) error {
	query := `
		INSERT INTO mpox_case_exposure_history (
			case_id, traveled_country_reported_mpox, close_contact_mpox,
			intl_travel, contact_animals, domestic_wild_animals, sexual_exposure
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id`
	return tx.QueryRow(query,
		m.CaseID, m.TraveledCountryReportedMpox, m.CloseContactMpox,
		m.IntlTravel, m.ContactAnimals, m.DomesticWildAnimals, m.SexualExposure,
	).Scan(&m.ID)
}

func (m *MpoxClinicalManifestations) Insert(tx *sql.Tx) error {
	query := `
		INSERT INTO mpox_clinical_manifestations (
			case_id, onset_date, fever, fever_temperature, lymphadenopathy,
			symptoms, symptom_other_specify, nausea_vomiting, pregnant,
			pregnant_trimester, vaccinated, vaccination_date, rash,
			rash_onset_date, rash_distribution, rash_type, underlying_illness,
			underlying_illness_details
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18
		)
		RETURNING id`
	return tx.QueryRow(query,
		m.CaseID, m.OnsetDate, m.Fever, m.FeverTemperature, m.Lymphadenopathy,
		m.Symptoms, m.SymptomOtherSpecify, m.NauseaVomiting, m.Pregnant,
		m.PregnantTrimester, m.Vaccinated, m.VaccinationDate, m.Rash,
		m.RashOnsetDate, m.RashDistribution, m.RashType, m.UnderlyingIllness,
		m.UnderlyingIllnessDetails,
	).Scan(&m.ID)
}

func (m *MpoxTravelHistory) Insert(tx *sql.Tx) error {
	query := `
		INSERT INTO mpox_travel_history (
			case_id, travel_outside_uganda, country_visited, location_visited,
			date_arrival, date_departure, activities_location
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id`
	return tx.QueryRow(query,
		m.CaseID, m.TravelOutsideUganda, m.CountryVisited, m.LocationVisited,
		m.DateArrival, m.DateDeparture, m.ActivitiesLocation,
	).Scan(&m.ID)
}

func (m *MpoxLabInvestigation) Insert(tx *sql.Tx) error {
	query := `
		INSERT INTO mpox_lab_investigation (
			case_id, lab_id, sample_collected, sample_other_specify,
			test_requested, test_other_specify, date_sample_collection,
			time_sample_collection, date_sample_dispatch, sample_collector_name,
			sample_collector_phone, date_sample_reception, time_sample_reception,
			sample_recipient_name, sample_recipient_phone, genomic_characterization,
			clade, accession_number
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18
		)
		RETURNING id`
	return tx.QueryRow(query,
		m.CaseID, m.LabID, m.SampleCollected, m.SampleOtherSpecify,
		m.TestRequested, m.TestOtherSpecify, m.DateSampleCollection,
		m.TimeSampleCollection, m.DateSampleDispatch, m.SampleCollectorName,
		m.SampleCollectorPhone, m.DateSampleReception, m.TimeSampleReception,
		m.SampleRecipientName, m.SampleRecipientPhone, m.GenomicCharacterization,
		m.Clade, m.AccessionNumber,
	).Scan(&m.ID)
}
