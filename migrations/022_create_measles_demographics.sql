CREATE TABLE measles_demographics (
    id SERIAL PRIMARY KEY,
    patient_id UUID REFERENCES measles_patients(patient_id) ON DELETE CASCADE,
    onset_district VARCHAR(100),
    reporting_unit VARCHAR(100),
    age_months INT,
    head_of_household VARCHAR(100),
    guardian_occupation VARCHAR(100),
    home_district VARCHAR(100),
    subcounty VARCHAR(100),
    parish VARCHAR(100),
    lc1_zone VARCHAR(100),
    lc1_chairman VARCHAR(100),
    lc1_tel VARCHAR(20)
); 