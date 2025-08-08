-- Create the correct mpox_daily_followup table
-- This replaces the old mpox_daily_follow_up table with the proper structure

-- Drop the old table if it exists
DROP TABLE IF EXISTS mpox_daily_follow_up;

-- Create the new table with the correct structure
CREATE TABLE IF NOT EXISTS mpox_daily_followup (
    id SERIAL PRIMARY KEY,
    client_id INTEGER NOT NULL REFERENCES clients(id) ON DELETE CASCADE,
    followup_date DATE NOT NULL,
    encounter_type TEXT[],
    other_site TEXT,
    spo2 INTEGER,
    new_lesions BOOLEAN,
    disease_progression TEXT,
    progression_other TEXT,
    lesion_face TEXT,
    lesion_mouth TEXT,
    lesion_chest TEXT,
    lesion_abdomen TEXT,
    lesion_back TEXT,
    lesion_arms TEXT,
    lesion_palms TEXT,
    lesion_forearms TEXT,
    lesion_thighs TEXT,
    lesion_legs TEXT,
    lesion_soles TEXT,
    lesion_genitalia TEXT,
    lesion_perianal TEXT,
    lesion_other TEXT,
    lesion_specify_where TEXT,
    type_macule TEXT,
    type_papule TEXT,
    type_vesicle TEXT,
    type_pustule TEXT,
    type_umbilicated TEXT,
    type_ulcerated TEXT,
    type_crusting TEXT,
    type_partialscab TEXT,
    type_other TEXT,
    pain_present BOOLEAN,
    pain_score INTEGER,
    pain_description TEXT,
    temperature NUMERIC(4,1),
    heart_rate INTEGER,
    respiratory_rate INTEGER,
    bp_systolic INTEGER,
    bp_diastolic INTEGER,
    consciousness TEXT,
    data_entrant TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create index for better performance
CREATE INDEX IF NOT EXISTS idx_mpox_daily_followup_client_id ON mpox_daily_followup(client_id);
CREATE INDEX IF NOT EXISTS idx_mpox_daily_followup_followup_date ON mpox_daily_followup(followup_date); 