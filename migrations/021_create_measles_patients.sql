CREATE TABLE measles_patients (
    patient_id UUID PRIMARY KEY,
    measles_code VARCHAR(32) UNIQUE,
    patient_name VARCHAR(100),
    sex VARCHAR(10),
    dob DATE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
); 