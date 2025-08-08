-- Migration to add missing fields to mpox_demographics table
-- Only add fields that don't already exist

-- Add lymphadenopathy pain field
ALTER TABLE mpox_demographics ADD COLUMN IF NOT EXISTS lymph_painful VARCHAR(10);

-- Add lymphadenopathy location fields
ALTER TABLE mpox_demographics ADD COLUMN IF NOT EXISTS lymph_location TEXT;
ALTER TABLE mpox_demographics ADD COLUMN IF NOT EXISTS lymph_other_detail TEXT;

-- Add lymphadenopathy pain location fields
ALTER TABLE mpox_demographics ADD COLUMN IF NOT EXISTS lymph_pain_location TEXT;
ALTER TABLE mpox_demographics ADD COLUMN IF NOT EXISTS lymph_pain_other_detail TEXT;

-- Add site of first encounter other detail
ALTER TABLE mpox_demographics ADD COLUMN IF NOT EXISTS site_of_first_encounter_other TEXT;

-- Add suspect confirmed case field (this might already exist with a different name)
ALTER TABLE mpox_demographics ADD COLUMN IF NOT EXISTS suspect_confirmed_case VARCHAR(20);

-- Check if suspect_confirmed_case column exists, if not, try to add it
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns 
                   WHERE table_name = 'mpox_demographics' 
                   AND column_name = 'suspect_confirmed_case') THEN
        ALTER TABLE mpox_demographics ADD COLUMN suspect_confirmed_case VARCHAR(20);
    END IF;
END $$; 