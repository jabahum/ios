CREATE TABLE measles_results (
    id SERIAL PRIMARY KEY,
    patient_id UUID REFERENCES measles_patients(patient_id) ON DELETE CASCADE,
    serology_igm VARCHAR(50),
    serology_date DATE,
    serology_epi_sent_date DATE,
    virus_isolation_urine VARCHAR(50),
    virus_isolation_date DATE,
    final_classification INT,
    results_sent_date DATE
); 