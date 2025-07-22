CREATE TABLE measles_investigators (
    id SERIAL PRIMARY KEY,
    patient_id UUID REFERENCES measles_patients(patient_id) ON DELETE CASCADE,
    investigator_name VARCHAR(100),
    investigator_title VARCHAR(100),
    investigator_date DATE
); 