-- Migration to add missing columns to mpox_onset_vitals table
-- Add columns that the model expects but are missing from the database

-- Add missing symptom columns
ALTER TABLE mpox_onset_vitals ADD COLUMN IF NOT EXISTS pruritis BOOLEAN;
ALTER TABLE mpox_onset_vitals ADD COLUMN IF NOT EXISTS rectal_pain BOOLEAN;
ALTER TABLE mpox_onset_vitals ADD COLUMN IF NOT EXISTS rectal_pain_onset_date DATE;
ALTER TABLE mpox_onset_vitals ADD COLUMN IF NOT EXISTS lesions BOOLEAN;
ALTER TABLE mpox_onset_vitals ADD COLUMN IF NOT EXISTS lesions_onset_date DATE;
ALTER TABLE mpox_onset_vitals ADD COLUMN IF NOT EXISTS lymphadenopathy BOOLEAN;
ALTER TABLE mpox_onset_vitals ADD COLUMN IF NOT EXISTS lymphadenopathy_onset_date DATE;
ALTER TABLE mpox_onset_vitals ADD COLUMN IF NOT EXISTS pruritis_onset_date DATE;
ALTER TABLE mpox_onset_vitals ADD COLUMN IF NOT EXISTS pain_swallowing BOOLEAN;
ALTER TABLE mpox_onset_vitals ADD COLUMN IF NOT EXISTS pain_swallowing_onset_date DATE;
ALTER TABLE mpox_onset_vitals ADD COLUMN IF NOT EXISTS difficulty_swallowing BOOLEAN;
ALTER TABLE mpox_onset_vitals ADD COLUMN IF NOT EXISTS difficulty_swallowing_onset_date DATE;
ALTER TABLE mpox_onset_vitals ADD COLUMN IF NOT EXISTS urethritis BOOLEAN;
ALTER TABLE mpox_onset_vitals ADD COLUMN IF NOT EXISTS urethritis_onset_date DATE;
ALTER TABLE mpox_onset_vitals ADD COLUMN IF NOT EXISTS chest_pain BOOLEAN;
ALTER TABLE mpox_onset_vitals ADD COLUMN IF NOT EXISTS chest_pain_onset_date DATE;
ALTER TABLE mpox_onset_vitals ADD COLUMN IF NOT EXISTS decreased_urine BOOLEAN;
ALTER TABLE mpox_onset_vitals ADD COLUMN IF NOT EXISTS decreased_urine_onset_date DATE;
ALTER TABLE mpox_onset_vitals ADD COLUMN IF NOT EXISTS dizziness BOOLEAN;
ALTER TABLE mpox_onset_vitals ADD COLUMN IF NOT EXISTS dizziness_onset_date DATE;
ALTER TABLE mpox_onset_vitals ADD COLUMN IF NOT EXISTS joint_pain BOOLEAN;
ALTER TABLE mpox_onset_vitals ADD COLUMN IF NOT EXISTS joint_pain_onset_date DATE;
ALTER TABLE mpox_onset_vitals ADD COLUMN IF NOT EXISTS psychological_disturbance BOOLEAN;
ALTER TABLE mpox_onset_vitals ADD COLUMN IF NOT EXISTS psychological_disturbance_onset_date DATE; 