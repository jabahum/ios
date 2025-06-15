-- Create main case investigation table
CREATE TABLE IF NOT EXISTS mpox_case_investigation (
    id SERIAL PRIMARY KEY,
    case_id VARCHAR(50) NOT NULL UNIQUE,
    date TIMESTAMP NOT NULL,
    case_status VARCHAR(50),
    case_classification VARCHAR(50)
);

-- Create patient demographics table
CREATE TABLE IF NOT EXISTS mpox_patient_demographics (
    id SERIAL PRIMARY KEY,
    case_id VARCHAR(50) NOT NULL REFERENCES mpox_case_investigation(case_id),
    health_facility_case_id VARCHAR(50),
    surname VARCHAR(100) NOT NULL,
    other_names VARCHAR(100),
    sex VARCHAR(10) NOT NULL,
    date_of_birth DATE NOT NULL,
    age INTEGER NOT NULL,
    parish VARCHAR(100),
    sub_county VARCHAR(100),
    physical_address TEXT,
    contact_telephone VARCHAR(20),
    occupation VARCHAR(100),
    nationality VARCHAR(50),
    vaccination_status VARCHAR(50),
    date_of_vaccination DATE,
    next_of_kin VARCHAR(100),
    next_of_kin_contact VARCHAR(20),
    marital_status VARCHAR(20),
    if_dead_date_of_death DATE,
    admission_date DATE,
    onset_date DATE,
    rash_onset_date DATE
);

-- Create clinician info table
CREATE TABLE IF NOT EXISTS mpox_clinician_info (
    id SERIAL PRIMARY KEY,
    case_id VARCHAR(50) NOT NULL REFERENCES mpox_case_investigation(case_id),
    clinician_name VARCHAR(100),
    clinician_contact VARCHAR(20),
    facility_name VARCHAR(100),
    clinician_email VARCHAR(100),
    facility_district VARCHAR(100),
    pdpid_number VARCHAR(50),
    admission_date DATE,
    ward VARCHAR(50)
);

-- Create exposure history table
CREATE TABLE IF NOT EXISTS mpox_case_exposure_history (
    id SERIAL PRIMARY KEY,
    case_id VARCHAR(50) NOT NULL REFERENCES mpox_case_investigation(case_id),
    traveled_country_reported_mpox VARCHAR(50),
    close_contact_mpox VARCHAR(50),
    intl_travel VARCHAR(50),
    contact_animals VARCHAR(50),
    domestic_wild_animals VARCHAR(50),
    sexual_exposure VARCHAR(50)
);

-- Create clinical manifestations table
CREATE TABLE IF NOT EXISTS mpox_clinical_manifestations (
    id SERIAL PRIMARY KEY,
    case_id VARCHAR(50) NOT NULL REFERENCES mpox_case_investigation(case_id),
    onset_date DATE,
    fever VARCHAR(50),
    fever_temperature VARCHAR(50),
    lymphadenopathy VARCHAR(50),
    symptoms TEXT[],
    symptom_other_specify VARCHAR(100),
    nausea_vomiting VARCHAR(50),
    pregnant VARCHAR(50),
    pregnant_trimester VARCHAR(50),
    vaccinated VARCHAR(50),
    vaccination_date VARCHAR(50),
    rash VARCHAR(50),
    rash_onset_date DATE,
    rash_distribution TEXT[],
    rash_type TEXT[],
    underlying_illness VARCHAR(50),
    underlying_illness_details TEXT
);

-- Create travel history table
CREATE TABLE IF NOT EXISTS mpox_travel_history (
    id SERIAL PRIMARY KEY,
    case_id VARCHAR(50) NOT NULL REFERENCES mpox_case_investigation(case_id),
    travel_outside_uganda VARCHAR(50),
    country_visited TEXT[],
    location_visited TEXT[],
    date_arrival TEXT[],
    date_departure TEXT[],
    activities_location TEXT[]
);

-- Create lab investigation table
CREATE TABLE IF NOT EXISTS mpox_lab_investigation (
    id SERIAL PRIMARY KEY,
    case_id VARCHAR(50) NOT NULL REFERENCES mpox_case_investigation(case_id),
    lab_id VARCHAR(50),
    sample_collected TEXT[],
    sample_other_specify VARCHAR(100),
    test_requested TEXT[],
    test_other_specify VARCHAR(100),
    date_sample_collection DATE,
    time_sample_collection TIME,
    date_sample_dispatch DATE,
    sample_collector_name VARCHAR(100),
    sample_collector_phone VARCHAR(20),
    date_sample_reception DATE,
    time_sample_reception TIME,
    sample_recipient_name VARCHAR(100),
    sample_recipient_phone VARCHAR(20),
    genomic_characterization VARCHAR(50),
    clade TEXT[],
    accession_number VARCHAR(50)
);

-- Create indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_mpox_case_investigation_case_id ON mpox_case_investigation(case_id);
CREATE INDEX IF NOT EXISTS idx_mpox_patient_demographics_case_id ON mpox_patient_demographics(case_id);
CREATE INDEX IF NOT EXISTS idx_mpox_clinician_info_case_id ON mpox_clinician_info(case_id);
CREATE INDEX IF NOT EXISTS idx_mpox_case_exposure_history_case_id ON mpox_case_exposure_history(case_id);
CREATE INDEX IF NOT EXISTS idx_mpox_clinical_manifestations_case_id ON mpox_clinical_manifestations(case_id);
CREATE INDEX IF NOT EXISTS idx_mpox_travel_history_case_id ON mpox_travel_history(case_id);
CREATE INDEX IF NOT EXISTS idx_mpox_lab_investigation_case_id ON mpox_lab_investigation(case_id); 