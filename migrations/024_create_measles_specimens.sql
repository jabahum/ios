CREATE TABLE measles_specimens (
    id SERIAL PRIMARY KEY,
    patient_id UUID REFERENCES measles_patients(patient_id) ON DELETE CASCADE,
    blood_collection_date DATE,
    blood_sent_date DATE,
    blood_received_date DATE,
    blood_condition TEXT,
    urine_collection_date DATE,
    urine_sent_date DATE,
    urine_received_date DATE,
    urine_condition TEXT,
    form_sent_date DATE,
    form_received_date DATE
); 