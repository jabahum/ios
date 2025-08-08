-- Migration to add individual symptom onset date fields
-- This adds date fields for each symptom to track when each symptom started

ALTER TABLE mpox_onset_vitals 
ADD COLUMN IF NOT EXISTS fever_onset_date DATE,
ADD COLUMN IF NOT EXISTS sore_throat_onset_date DATE,
ADD COLUMN IF NOT EXISTS headache_onset_date DATE,
ADD COLUMN IF NOT EXISTS muscle_aches_onset_date DATE,
ADD COLUMN IF NOT EXISTS cough_onset_date DATE,
ADD COLUMN IF NOT EXISTS fatigue_onset_date DATE,
ADD COLUMN IF NOT EXISTS oral_pain_onset_date DATE,
ADD COLUMN IF NOT EXISTS nausea_onset_date DATE,
ADD COLUMN IF NOT EXISTS vomiting_onset_date DATE,
ADD COLUMN IF NOT EXISTS diarrhea_onset_date DATE,
ADD COLUMN IF NOT EXISTS rectal_pain_onset_date DATE,
ADD COLUMN IF NOT EXISTS lesions_onset_date DATE,
ADD COLUMN IF NOT EXISTS lymphadenopathy_onset_date DATE,
ADD COLUMN IF NOT EXISTS pruritis_onset_date DATE,
ADD COLUMN IF NOT EXISTS pain_swallowing_onset_date DATE,
ADD COLUMN IF NOT EXISTS difficulty_swallowing_onset_date DATE,
ADD COLUMN IF NOT EXISTS urethritis_onset_date DATE,
ADD COLUMN IF NOT EXISTS chest_pain_onset_date DATE,
ADD COLUMN IF NOT EXISTS decreased_urine_onset_date DATE,
ADD COLUMN IF NOT EXISTS dizziness_onset_date DATE,
ADD COLUMN IF NOT EXISTS joint_pain_onset_date DATE,
ADD COLUMN IF NOT EXISTS psychological_disturbance_onset_date DATE;

-- Add comments to document the new fields
COMMENT ON COLUMN mpox_onset_vitals.fever_onset_date IS 'Date when fever symptoms first appeared';
COMMENT ON COLUMN mpox_onset_vitals.sore_throat_onset_date IS 'Date when sore throat symptoms first appeared';
COMMENT ON COLUMN mpox_onset_vitals.headache_onset_date IS 'Date when headache symptoms first appeared';
COMMENT ON COLUMN mpox_onset_vitals.muscle_aches_onset_date IS 'Date when muscle aches symptoms first appeared';
COMMENT ON COLUMN mpox_onset_vitals.cough_onset_date IS 'Date when cough symptoms first appeared';
COMMENT ON COLUMN mpox_onset_vitals.fatigue_onset_date IS 'Date when fatigue symptoms first appeared';
COMMENT ON COLUMN mpox_onset_vitals.oral_pain_onset_date IS 'Date when oral pain symptoms first appeared';
COMMENT ON COLUMN mpox_onset_vitals.nausea_onset_date IS 'Date when nausea symptoms first appeared';
COMMENT ON COLUMN mpox_onset_vitals.vomiting_onset_date IS 'Date when vomiting symptoms first appeared';
COMMENT ON COLUMN mpox_onset_vitals.diarrhea_onset_date IS 'Date when diarrhea symptoms first appeared';
COMMENT ON COLUMN mpox_onset_vitals.rectal_pain_onset_date IS 'Date when rectal pain symptoms first appeared';
COMMENT ON COLUMN mpox_onset_vitals.lesions_onset_date IS 'Date when lesions first appeared';
COMMENT ON COLUMN mpox_onset_vitals.lymphadenopathy_onset_date IS 'Date when lymphadenopathy symptoms first appeared';
COMMENT ON COLUMN mpox_onset_vitals.pruritis_onset_date IS 'Date when pruritis (itching) symptoms first appeared';
COMMENT ON COLUMN mpox_onset_vitals.pain_swallowing_onset_date IS 'Date when pain with swallowing symptoms first appeared';
COMMENT ON COLUMN mpox_onset_vitals.difficulty_swallowing_onset_date IS 'Date when difficulty swallowing symptoms first appeared';
COMMENT ON COLUMN mpox_onset_vitals.urethritis_onset_date IS 'Date when urethritis symptoms first appeared';
COMMENT ON COLUMN mpox_onset_vitals.chest_pain_onset_date IS 'Date when chest pain symptoms first appeared';
COMMENT ON COLUMN mpox_onset_vitals.decreased_urine_onset_date IS 'Date when decreased urine output symptoms first appeared';
COMMENT ON COLUMN mpox_onset_vitals.dizziness_onset_date IS 'Date when dizziness symptoms first appeared';
COMMENT ON COLUMN mpox_onset_vitals.joint_pain_onset_date IS 'Date when joint pain (arthralgia) symptoms first appeared';
COMMENT ON COLUMN mpox_onset_vitals.psychological_disturbance_onset_date IS 'Date when psychological disturbance symptoms first appeared'; 