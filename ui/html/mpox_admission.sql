-- Mpox Admission Form PostgreSQL Schemas

-- 1a. Demographics
CREATE TABLE mpox_demographics (
  id SERIAL PRIMARY KEY,
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
  site_of_first_encounter VARCHAR(50)
);

-- 1b. Exposure and Social History
CREATE TABLE mpox_exposure_history (
  id SERIAL PRIMARY KEY,
  demographics_id INT REFERENCES mpox_demographics(id),
  known_link BOOLEAN,
  sexually_active BOOLEAN,
  sex_of_partners VARCHAR(20),
  recent_travel BOOLEAN,
  travel_high_risk BOOLEAN,
  travel_details TEXT
);

-- 1c. Date of Onset and Vital Signs
CREATE TABLE mpox_onset_vitals (
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
  weight_kg NUMERIC
);

-- 1d. Co-morbidities
CREATE TABLE mpox_comorbidities (
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
  viral_load VARCHAR(50)
);

-- 1e. Rash Evaluation
CREATE TABLE mpox_rash_evaluation (
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
  pain_score INT
);

-- 1g. Laboratory Investigations
CREATE TABLE mpox_laboratory_investigations (
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
  mpox_result VARCHAR(20)
);

-- Data Entrant
CREATE TABLE mpox_data_entrant (
  id SERIAL PRIMARY KEY,
  demographics_id INT REFERENCES mpox_demographics(id),
  name VARCHAR(100)
); 