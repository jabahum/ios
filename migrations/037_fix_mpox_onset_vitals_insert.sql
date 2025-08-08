-- Migration to ensure all required columns exist in mpox_onset_vitals table
-- This will add any missing columns that the INSERT statement expects

-- Add any missing columns that the model expects
ALTER TABLE mpox_onset_vitals ADD COLUMN IF NOT EXISTS fever_onset_date DATE;
ALTER TABLE mpox_onset_vitals ADD COLUMN IF NOT EXISTS sore_throat_onset_date DATE;
ALTER TABLE mpox_onset_vitals ADD COLUMN IF NOT EXISTS headache_onset_date DATE;
ALTER TABLE mpox_onset_vitals ADD COLUMN IF NOT EXISTS muscle_aches_onset_date DATE;
ALTER TABLE mpox_onset_vitals ADD COLUMN IF NOT EXISTS cough_onset_date DATE;
ALTER TABLE mpox_onset_vitals ADD COLUMN IF NOT EXISTS fatigue_onset_date DATE;
ALTER TABLE mpox_onset_vitals ADD COLUMN IF NOT EXISTS oral_pain_onset_date DATE;
ALTER TABLE mpox_onset_vitals ADD COLUMN IF NOT EXISTS nausea_onset_date DATE;
ALTER TABLE mpox_onset_vitals ADD COLUMN IF NOT EXISTS vomiting_onset_date DATE;
ALTER TABLE mpox_onset_vitals ADD COLUMN IF NOT EXISTS diarrhea_onset_date DATE;
ALTER TABLE mpox_onset_vitals ADD COLUMN IF NOT EXISTS lesions_onset_date DATE;
ALTER TABLE mpox_onset_vitals ADD COLUMN IF NOT EXISTS temperature FLOAT;
ALTER TABLE mpox_onset_vitals ADD COLUMN IF NOT EXISTS heart_rate INTEGER;
ALTER TABLE mpox_onset_vitals ADD COLUMN IF NOT EXISTS respiratory_rate INTEGER;
ALTER TABLE mpox_onset_vitals ADD COLUMN IF NOT EXISTS bp_systolic INTEGER;
ALTER TABLE mpox_onset_vitals ADD COLUMN IF NOT EXISTS bp_diastolic INTEGER;
ALTER TABLE mpox_onset_vitals ADD COLUMN IF NOT EXISTS dehydration BOOLEAN;
ALTER TABLE mpox_onset_vitals ADD COLUMN IF NOT EXISTS avpu VARCHAR(50);
ALTER TABLE mpox_onset_vitals ADD COLUMN IF NOT EXISTS height_cm FLOAT;
ALTER TABLE mpox_onset_vitals ADD COLUMN IF NOT EXISTS weight_kg FLOAT; 