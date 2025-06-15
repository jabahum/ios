-- Mpox Admission Form PostgreSQL Schemas

-- 1a. Demographics
CREATE TABLE IF NOT EXISTS mpox_demographics (
    id SERIAL PRIMARY KEY,
    client_id INTEGER REFERENCES clients(id),
    sex VARCHAR(10),
    date_of_birth DATE,
    age_years INT,
    age_months INT,
    age_days INT,
    health_care_worker VARCHAR(10),
    laboratory_worker VARCHAR(10),
    ppe_status VARCHAR(50),
    tribe VARCHAR(50),
    pregnant BOOLEAN,
    gestational_weeks INT,
    lmnp DATE,
    recently_pregnant BOOLEAN,
    pregnant_22_42 BOOLEAN,
    tetanus_vaccination BOOLEAN,
    occupation VARCHAR(100),
    site_of_first_encounter VARCHAR(50),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 1b. Exposure and Social History
CREATE TABLE IF NOT EXISTS mpox_exposure_history (
    id SERIAL PRIMARY KEY,
    demographics_id INT REFERENCES mpox_demographics(id),
    known_link BOOLEAN,
    sexually_active BOOLEAN,
    sex_of_partners VARCHAR(20),
    recent_travel BOOLEAN,
    travel_high_risk BOOLEAN,
    travel_details TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 1c. Date of Onset and Vital Signs
CREATE TABLE IF NOT EXISTS mpox_onset_vitals (
    id SERIAL PRIMARY KEY,
    demographics_id INT REFERENCES mpox_demographics(id),
    symptom_onset DATE,
    fever BOOLEAN,
    sore_throat BOOLEAN,
    headache BOOLEAN,
    muscle_aches BOOLEAN,
    cough BOOLEAN,
    fatigue BOOLEAN,
    oral_pain BOOLEAN,
    nausea BOOLEAN,
    vomiting BOOLEAN,
    diarrhea BOOLEAN,
    rectal_pain BOOLEAN,
    lesions BOOLEAN,
    lymphadenopathy BOOLEAN,
    temperature NUMERIC,
    heart_rate INT,
    respiratory_rate INT,
    bp_systolic INT,
    bp_diastolic INT,
    dehydration BOOLEAN,
    avpu VARCHAR(20),
    height_cm NUMERIC,
    weight_kg NUMERIC,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 1d. Co-morbidities
CREATE TABLE IF NOT EXISTS mpox_comorbidities (
    id SERIAL PRIMARY KEY,
    demographics_id INT REFERENCES mpox_demographics(id),
    cardiac_disease BOOLEAN,
    hypertension BOOLEAN,
    pulmonary_disease BOOLEAN,
    asthma BOOLEAN,
    kidney_disease BOOLEAN,
    liver_disease BOOLEAN,
    neurological_disorder BOOLEAN,
    diabetes BOOLEAN,
    tuberculosis_active BOOLEAN,
    tuberculosis_previous BOOLEAN,
    asplenia BOOLEAN,
    neoplasm BOOLEAN,
    alcohol_use_disorder BOOLEAN,
    immunosuppressants BOOLEAN,
    sti BOOLEAN,
    hiv_status VARCHAR(20),
    art_regimen VARCHAR(100),
    cd4 INT,
    viral_load VARCHAR(50),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 1e. Rash Evaluation
CREATE TABLE IF NOT EXISTS mpox_rash_evaluation (
    id SERIAL PRIMARY KEY,
    demographics_id INT REFERENCES mpox_demographics(id),
    severity VARCHAR(20),
    face BOOLEAN,
    nares BOOLEAN,
    mouth BOOLEAN,
    chest BOOLEAN,
    abdomen BOOLEAN,
    back BOOLEAN,
    perianal BOOLEAN,
    genitals BOOLEAN,
    palms BOOLEAN,
    arms BOOLEAN,
    forearms BOOLEAN,
    thighs BOOLEAN,
    legs BOOLEAN,
    soles BOOLEAN,
    macule BOOLEAN,
    papule BOOLEAN,
    early_vesicle BOOLEAN,
    small_pustule BOOLEAN,
    umbilicated_pustule BOOLEAN,
    ulcerated_lesion BOOLEAN,
    crusting BOOLEAN,
    partially_removed_scab BOOLEAN,
    pain_at_lesion BOOLEAN,
    pain_score INT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 1g. Laboratory Investigations
CREATE TABLE IF NOT EXISTS mpox_laboratory_investigations (
    id SERIAL PRIMARY KEY,
    demographics_id INT REFERENCES mpox_demographics(id),
    alt NUMERIC,
    ast NUMERIC,
    creatinine NUMERIC,
    potassium NUMERIC,
    urea NUMERIC,
    creatine_kinase NUMERIC,
    calcium NUMERIC,
    sodium NUMERIC,
    crp NUMERIC,
    glucose NUMERIC,
    lactate NUMERIC,
    haemoglobin NUMERIC,
    total_bilirubin NUMERIC,
    wbc_count NUMERIC,
    platelets NUMERIC,
    prothrombin_time NUMERIC,
    aptt NUMERIC,
    malaria_result VARCHAR(20),
    syphilis_result VARCHAR(20),
    mpox_result VARCHAR(20),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Data Entrant
CREATE TABLE IF NOT EXISTS mpox_data_entrant (
    id SERIAL PRIMARY KEY,
    demographics_id INT REFERENCES mpox_demographics(id),
    name VARCHAR(100),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Daily Follow-up Table
CREATE TABLE IF NOT EXISTS mpox_daily_follow_up (
    id SERIAL PRIMARY KEY,
    client_id INTEGER REFERENCES clients(id),
    encounter_date TIMESTAMP WITH TIME ZONE,
    temperature NUMERIC,
    heart_rate INTEGER,
    respiratory_rate INTEGER,
    bp_systolic INTEGER,
    bp_diastolic INTEGER,
    symptoms TEXT,
    rash_present BOOLEAN,
    rash_description TEXT,
    lesion_count INTEGER,
    pain_score INTEGER,
    lab_results TEXT,
    other_notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Add indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_mpox_demographics_client_id ON mpox_demographics(client_id);
CREATE INDEX IF NOT EXISTS idx_mpox_exposure_history_demographics_id ON mpox_exposure_history(demographics_id);
CREATE INDEX IF NOT EXISTS idx_mpox_onset_vitals_demographics_id ON mpox_onset_vitals(demographics_id);
CREATE INDEX IF NOT EXISTS idx_mpox_comorbidities_demographics_id ON mpox_comorbidities(demographics_id);
CREATE INDEX IF NOT EXISTS idx_mpox_rash_evaluation_demographics_id ON mpox_rash_evaluation(demographics_id);
CREATE INDEX IF NOT EXISTS idx_mpox_laboratory_investigations_demographics_id ON mpox_laboratory_investigations(demographics_id);
CREATE INDEX IF NOT EXISTS idx_mpox_data_entrant_demographics_id ON mpox_data_entrant(demographics_id);
CREATE INDEX IF NOT EXISTS idx_mpox_daily_follow_up_client_id ON mpox_daily_follow_up(client_id);
CREATE INDEX IF NOT EXISTS idx_mpox_daily_follow_up_encounter_date ON mpox_daily_follow_up(encounter_date);

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

-- Create case exposure history table
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
    fever VARCHAR(10),
    fever_temperature VARCHAR(20),
    lymphadenopathy VARCHAR(10),
    symptoms TEXT[],
    symptom_other_specify TEXT,
    nausea_vomiting VARCHAR(10),
    pregnant VARCHAR(10),
    pregnant_trimester VARCHAR(20),
    vaccinated VARCHAR(10),
    vaccination_date VARCHAR(20),
    rash VARCHAR(10),
    rash_onset_date DATE,
    rash_distribution TEXT[],
    rash_type TEXT[],
    underlying_illness VARCHAR(10),
    underlying_illness_details TEXT
);

-- Create travel history table
CREATE TABLE IF NOT EXISTS mpox_travel_history (
    id SERIAL PRIMARY KEY,
    case_id VARCHAR(50) NOT NULL REFERENCES mpox_case_investigation(case_id),
    travel_outside_uganda VARCHAR(10),
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
    sample_other_specify TEXT,
    test_requested TEXT[],
    test_other_specify TEXT,
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

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_mpox_case_investigation_case_id ON mpox_case_investigation(case_id);
CREATE INDEX IF NOT EXISTS idx_mpox_patient_demographics_case_id ON mpox_patient_demographics(case_id);
CREATE INDEX IF NOT EXISTS idx_mpox_clinician_info_case_id ON mpox_clinician_info(case_id);
CREATE INDEX IF NOT EXISTS idx_mpox_case_exposure_history_case_id ON mpox_case_exposure_history(case_id);
CREATE INDEX IF NOT EXISTS idx_mpox_clinical_manifestations_case_id ON mpox_clinical_manifestations(case_id);
CREATE INDEX IF NOT EXISTS idx_mpox_travel_history_case_id ON mpox_travel_history(case_id);
CREATE INDEX IF NOT EXISTS idx_mpox_lab_investigation_case_id ON mpox_lab_investigation(case_id); 