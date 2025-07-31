-- Migration: Create Polio CIF tables
-- Date: 2024-01-XX

-- Polio Case Investigation table
CREATE TABLE IF NOT EXISTS polio_case_investigation (
    id BIGSERIAL PRIMARY KEY,
    case_id VARCHAR(50) UNIQUE NOT NULL,
    epid_number VARCHAR(50),
    country VARCHAR(100),
    region_province VARCHAR(100),
    district VARCHAR(100),
    year_onset INTEGER,
    case_number INTEGER,
    received_date TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Polio Identification table
CREATE TABLE IF NOT EXISTS polio_identification (
    id BIGSERIAL PRIMARY KEY,
    case_id VARCHAR(50) REFERENCES polio_case_investigation(case_id) ON DELETE CASCADE,
    district VARCHAR(100),
    region_province VARCHAR(100),
    address TEXT,
    village VARCHAR(100),
    city VARCHAR(100),
    nearest_health_facility VARCHAR(200),
    longitude DECIMAL(10, 8),
    latitude DECIMAL(10, 8),
    patient_name VARCHAR(200) NOT NULL,
    father_mother VARCHAR(200),
    phone_number VARCHAR(50),
    date_of_birth DATE,
    age_years INTEGER,
    age_months INTEGER,
    sex VARCHAR(10),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Polio Notification/Investigation table
CREATE TABLE IF NOT EXISTS polio_notification_investigation (
    id BIGSERIAL PRIMARY KEY,
    case_id VARCHAR(50) REFERENCES polio_case_investigation(case_id) ON DELETE CASCADE,
    notified_by VARCHAR(200),
    date_of_notification DATE,
    date_of_investigation DATE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Polio Hospitalization table
CREATE TABLE IF NOT EXISTS polio_hospitalization (
    id BIGSERIAL PRIMARY KEY,
    case_id VARCHAR(50) REFERENCES polio_case_investigation(case_id) ON DELETE CASCADE,
    hospitalized BOOLEAN DEFAULT FALSE,
    date_of_admission DATE,
    hospital_record_number VARCHAR(100),
    hospital_name_address TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Polio Clinical History table
CREATE TABLE IF NOT EXISTS polio_clinical_history (
    id BIGSERIAL PRIMARY KEY,
    case_id VARCHAR(50) REFERENCES polio_case_investigation(case_id) ON DELETE CASCADE,
    fever_at_onset VARCHAR(10),
    date_onset_of_fever DATE,
    progressive_paralysis VARCHAR(10),
    date_onset_of_paralysis DATE,
    flaccid_acute_paralysis VARCHAR(10),
    sensation_loss VARCHAR(10),
    sudden_onset BOOLEAN DEFAULT FALSE,
    "asymmetric" BOOLEAN DEFAULT FALSE,
    left_arm_paralysis BOOLEAN DEFAULT FALSE,
    right_arm_paralysis BOOLEAN DEFAULT FALSE,
    left_leg_paralysis BOOLEAN DEFAULT FALSE,
    right_leg_paralysis BOOLEAN DEFAULT FALSE,
    diminished_reflexes BOOLEAN DEFAULT FALSE,
    diminished_muscle_tone BOOLEAN DEFAULT FALSE,
    muscle_wasting BOOLEAN DEFAULT FALSE,
    muscle_weakness BOOLEAN DEFAULT FALSE,
    respiratory_muscles BOOLEAN DEFAULT FALSE,
    face BOOLEAN DEFAULT FALSE,
    stiff_neck BOOLEAN DEFAULT FALSE,
    convulsions BOOLEAN DEFAULT FALSE,
    headache BOOLEAN DEFAULT FALSE,
    vomiting BOOLEAN DEFAULT FALSE,
    diarrhoea BOOLEAN DEFAULT FALSE,
    other_sites TEXT,
    recent_injection VARCHAR(10),
    total_injections INTEGER,
    injection_type TEXT,
    paralyzed_limb_sensitive VARCHAR(10),
    injection_facility_name TEXT,
    provisional_diagnosis TEXT,
    true_afp_case VARCHAR(10),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Polio Immunization History table
CREATE TABLE IF NOT EXISTS polio_immunization_history (
    id BIGSERIAL PRIMARY KEY,
    case_id VARCHAR(50) REFERENCES polio_case_investigation(case_id) ON DELETE CASCADE,
    total_polio_doses INTEGER,
    exclude_dose_at_birth BOOLEAN DEFAULT FALSE,
    opv_dose_at_birth DATE,
    opv_dose1 DATE,
    opv_dose2 DATE,
    opv_dose3 DATE,
    opv_dose4 DATE,
    opv_dose_more_than4 DATE,
    last_opv_dose DATE,
    total_opv_sia INTEGER,
    last_opv_sia DATE,
    total_opv_ri INTEGER,
    total_ipv_sia INTEGER,
    total_ipv_ri INTEGER,
    last_ipv_sia DATE,
    source_of_ri_vaccination VARCHAR(20), -- 'Card' or 'Recall'
    unknown_zero_dose_reasons TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Polio Stool Specimen Collection table
CREATE TABLE IF NOT EXISTS polio_stool_specimen_collection (
    id BIGSERIAL PRIMARY KEY,
    case_id VARCHAR(50) REFERENCES polio_case_investigation(case_id) ON DELETE CASCADE,
    date_first_specimen DATE,
    date_second_specimen DATE,
    date_specimen_sent_national DATE,
    date_specimen_received_national DATE,
    date_specimen_sent_lab DATE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Polio Stool Specimen Results table
CREATE TABLE IF NOT EXISTS polio_stool_specimen_results (
    id BIGSERIAL PRIMARY KEY,
    case_id VARCHAR(50) REFERENCES polio_case_investigation(case_id) ON DELETE CASCADE,
    date_received_at_lab DATE,
    specimen_status_at_reception VARCHAR(20), -- 'Adequate' or 'Not adequate'
    date_combined_cell_culture DATE,
    date_results_sent_to_epi DATE,
    date_results_received_at_epi DATE,
    final_cell_culture_results VARCHAR(50), -- 'Suspected poliovirus', 'Negative', 'NPENT', 'Suspect poliovirus + NPENT'
    w1 VARCHAR(10),
    w2 VARCHAR(10),
    w3 VARCHAR(10),
    discordant_sabin VARCHAR(10),
    sl1 VARCHAR(10),
    sl2 VARCHAR(10),
    sl3 VARCHAR(10),
    r_npent VARCHAR(10),
    nev VARCHAR(10),
    date_sent_to_regional_lab DATE,
    date_it_differentiation_sent DATE,
    date_it_differentiation_received DATE,
    date_isolate_sent_sequencing DATE,
    date_seq_results_sent_program DATE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Polio Follow-up Examination table
CREATE TABLE IF NOT EXISTS polio_follow_up_examination (
    id BIGSERIAL PRIMARY KEY,
    case_id VARCHAR(50) REFERENCES polio_case_investigation(case_id) ON DELETE CASCADE,
    date_of_follow_up DATE,
    residual_paralysis_la BOOLEAN DEFAULT FALSE,
    residual_paralysis_ra BOOLEAN DEFAULT FALSE,
    residual_paralysis_ll BOOLEAN DEFAULT FALSE,
    residual_paralysis_rl BOOLEAN DEFAULT FALSE,
    results_of_exam VARCHAR(50), -- 'Residual Flaccid Paralysis', 'No residual paralysis', 'Lost follow-up', 'Died before follow-up', 'Residual Spastic Paralysis'
    immunocompromised_status VARCHAR(10),
    final_classification VARCHAR(50), -- 'Confirmed Polio', 'Compatible', 'Discarded', 'Not an AFP case'
    cvdpv BOOLEAN DEFAULT FALSE,
    avdpv BOOLEAN DEFAULT FALSE,
    ivdpv BOOLEAN DEFAULT FALSE,
    serotype VARCHAR(10), -- '1', '2', '3'
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Polio Patient History table
CREATE TABLE IF NOT EXISTS polio_patient_history (
    id BIGSERIAL PRIMARY KEY,
    case_id VARCHAR(50) REFERENCES polio_case_investigation(case_id) ON DELETE CASCADE,
    place1 VARCHAR(200),
    duration1_months INTEGER,
    duration1_days INTEGER,
    place2 VARCHAR(200),
    duration2_months INTEGER,
    duration2_days INTEGER,
    place3 VARCHAR(200),
    duration3_months INTEGER,
    duration3_days INTEGER,
    place4 VARCHAR(200),
    duration4_months INTEGER,
    duration4_days INTEGER,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Polio Investigator table
CREATE TABLE IF NOT EXISTS polio_investigator (
    id BIGSERIAL PRIMARY KEY,
    case_id VARCHAR(50) REFERENCES polio_case_investigation(case_id) ON DELETE CASCADE,
    investigator_name VARCHAR(200),
    investigator_title VARCHAR(100),
    unit VARCHAR(100),
    address TEXT,
    telephone VARCHAR(50),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_polio_case_investigation_case_id ON polio_case_investigation(case_id);
CREATE INDEX IF NOT EXISTS idx_polio_identification_case_id ON polio_identification(case_id);
CREATE INDEX IF NOT EXISTS idx_polio_notification_case_id ON polio_notification_investigation(case_id);
CREATE INDEX IF NOT EXISTS idx_polio_hospitalization_case_id ON polio_hospitalization(case_id);
CREATE INDEX IF NOT EXISTS idx_polio_clinical_history_case_id ON polio_clinical_history(case_id);
CREATE INDEX IF NOT EXISTS idx_polio_immunization_history_case_id ON polio_immunization_history(case_id);
CREATE INDEX IF NOT EXISTS idx_polio_stool_specimen_collection_case_id ON polio_stool_specimen_collection(case_id);
CREATE INDEX IF NOT EXISTS idx_polio_stool_specimen_results_case_id ON polio_stool_specimen_results(case_id);
CREATE INDEX IF NOT EXISTS idx_polio_follow_up_examination_case_id ON polio_follow_up_examination(case_id);
CREATE INDEX IF NOT EXISTS idx_polio_patient_history_case_id ON polio_patient_history(case_id);
CREATE INDEX IF NOT EXISTS idx_polio_investigator_case_id ON polio_investigator(case_id);

-- Add comments for documentation
COMMENT ON TABLE polio_case_investigation IS 'Main table for Polio Case Investigation Form';
COMMENT ON TABLE polio_identification IS 'Patient identification information for Polio cases';
COMMENT ON TABLE polio_notification_investigation IS 'Notification and investigation details for Polio cases';
COMMENT ON TABLE polio_hospitalization IS 'Hospitalization information for Polio cases';
COMMENT ON TABLE polio_clinical_history IS 'Clinical history and symptoms for Polio cases';
COMMENT ON TABLE polio_immunization_history IS 'Immunization history for Polio cases';
COMMENT ON TABLE polio_stool_specimen_collection IS 'Stool specimen collection details for Polio cases';
COMMENT ON TABLE polio_stool_specimen_results IS 'Stool specimen results for Polio cases';
COMMENT ON TABLE polio_follow_up_examination IS 'Follow-up examination details for Polio cases';
COMMENT ON TABLE polio_patient_history IS 'Patient visit history for Polio cases';
COMMENT ON TABLE polio_investigator IS 'Investigator information for Polio cases'; 