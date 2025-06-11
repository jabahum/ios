-- VHF Patient Information
CREATE TABLE IF NOT EXISTS vhf_patients (
    id SERIAL PRIMARY KEY,
    surname VARCHAR(100) NOT NULL,
    other_names VARCHAR(100),
    date_of_birth DATE,
    age_years INTEGER,
    age_months INTEGER,
    gender VARCHAR(10) NOT NULL,
    patient_phone VARCHAR(20),
    phone_owner VARCHAR(100),
    next_of_kin VARCHAR(100),
    next_of_kin_phone VARCHAR(20),
    data_capturer_name VARCHAR(100),
    data_capturer_phone VARCHAR(20),
    reporting_health_facility_name VARCHAR(200),
    case_code VARCHAR(50) UNIQUE,
    status VARCHAR(20) NOT NULL,
    date_of_death DATE,
    head_of_household VARCHAR(100),
    village_town VARCHAR(100),
    parish VARCHAR(100),
    subcounty VARCHAR(100),
    district VARCHAR(100),
    country_of_residence VARCHAR(100),
    occupation VARCHAR(100),
    ill_village_town VARCHAR(100),
    ill_subcounty VARCHAR(100),
    ill_district VARCHAR(100),
    latitude DECIMAL(10, 8),
    longitude DECIMAL(11, 8),
    date_residing_from DATE,
    date_residing_to DATE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- VHF Clinical Signs
CREATE TABLE IF NOT EXISTS vhf_clinical_signs (
    id SERIAL PRIMARY KEY,
    patient_id INTEGER REFERENCES vhf_patients(id) ON DELETE CASCADE,
    date_initial_onset DATE,
    fever BOOLEAN,
    date_fever DATE,
    duration_fever INTEGER,
    temp_source VARCHAR(50),
    temperature DECIMAL(4, 1),
    vomiting BOOLEAN,
    date_vomiting DATE,
    duration_vomiting INTEGER,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- VHF Current Hospitalization
CREATE TABLE IF NOT EXISTS vhf_hospitalization (
    id SERIAL PRIMARY KEY,
    patient_id INTEGER REFERENCES vhf_patients(id) ON DELETE CASCADE,
    hospitalized BOOLEAN NOT NULL,
    admission_date DATE,
    health_facility_name VARCHAR(200),
    in_isolation BOOLEAN,
    isolation_date DATE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- VHF Previous Hospitalization
CREATE TABLE IF NOT EXISTS vhf_previous_hospitalization (
    id SERIAL PRIMARY KEY,
    patient_id INTEGER REFERENCES vhf_patients(id) ON DELETE CASCADE,
    hospitalization_dates VARCHAR(200),
    health_facility_name VARCHAR(200),
    village VARCHAR(100),
    district VARCHAR(100),
    was_isolated BOOLEAN,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- VHF Risk Factors
CREATE TABLE IF NOT EXISTS vhf_risk_factors (
    id SERIAL PRIMARY KEY,
    patient_id INTEGER REFERENCES vhf_patients(id) ON DELETE CASCADE,
    contact_with_case BOOLEAN,
    contact_name VARCHAR(100),
    contact_relation VARCHAR(100),
    contact_dates VARCHAR(200),
    contact_village VARCHAR(100),
    contact_district VARCHAR(100),
    contact_status VARCHAR(50),
    contact_death_date DATE,
    contact_types VARCHAR(200),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- VHF Laboratory
CREATE TABLE IF NOT EXISTS vhf_laboratory (
    id SERIAL PRIMARY KEY,
    patient_id INTEGER REFERENCES vhf_patients(id) ON DELETE CASCADE,
    sample_collection_date DATE,
    sample_collection_time TIME,
    sample_type VARCHAR(100),
    other_sample_type VARCHAR(100),
    requested_test VARCHAR(200),
    serology VARCHAR(100),
    malaria_rdt VARCHAR(50),
    hiv_rdt VARCHAR(50),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- VHF Investigator
CREATE TABLE IF NOT EXISTS vhf_investigator (
    id SERIAL PRIMARY KEY,
    patient_id INTEGER REFERENCES vhf_patients(id) ON DELETE CASCADE,
    investigator_name VARCHAR(100) NOT NULL,
    phone VARCHAR(20),
    email VARCHAR(100),
    position VARCHAR(100),
    district VARCHAR(100),
    health_facility VARCHAR(200),
    information_source VARCHAR(200),
    proxy_name VARCHAR(100),
    proxy_relation VARCHAR(100),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
); 