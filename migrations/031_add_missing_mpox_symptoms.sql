-- Migration to add missing symptoms to mpox_onset_vitals table
-- These symptoms were in the original template but missing from the database schema

ALTER TABLE mpox_onset_vitals 
ADD COLUMN IF NOT EXISTS pruritis BOOLEAN,
ADD COLUMN IF NOT EXISTS pain_swallowing BOOLEAN,
ADD COLUMN IF NOT EXISTS difficulty_swallowing BOOLEAN,
ADD COLUMN IF NOT EXISTS urethritis BOOLEAN,
ADD COLUMN IF NOT EXISTS chest_pain BOOLEAN,
ADD COLUMN IF NOT EXISTS decreased_urine BOOLEAN,
ADD COLUMN IF NOT EXISTS dizziness BOOLEAN,
ADD COLUMN IF NOT EXISTS joint_pain BOOLEAN,
ADD COLUMN IF NOT EXISTS psychological_disturbance BOOLEAN;

-- Add comments to document the new symptom fields
COMMENT ON COLUMN mpox_onset_vitals.pruritis IS 'Pruritis (itching) symptom';
COMMENT ON COLUMN mpox_onset_vitals.pain_swallowing IS 'Pain with swallowing symptom';
COMMENT ON COLUMN mpox_onset_vitals.difficulty_swallowing IS 'Difficulty swallowing symptom';
COMMENT ON COLUMN mpox_onset_vitals.urethritis IS 'Urethritis symptom';
COMMENT ON COLUMN mpox_onset_vitals.chest_pain IS 'Chest pain symptom';
COMMENT ON COLUMN mpox_onset_vitals.decreased_urine IS 'Decreased urine output symptom';
COMMENT ON COLUMN mpox_onset_vitals.dizziness IS 'Dizziness symptom';
COMMENT ON COLUMN mpox_onset_vitals.joint_pain IS 'Joint pain (arthralgia) symptom';
COMMENT ON COLUMN mpox_onset_vitals.psychological_disturbance IS 'Psychological disturbance symptom'; 