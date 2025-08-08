-- Migration to add new fields to mpox_demographics table
-- Add new fields for enhanced mpox admission form

-- Add lymphadenopathy pain field
ALTER TABLE mpox_demographics ADD COLUMN IF NOT EXISTS lymph_painful VARCHAR(10);

-- Add lymphadenopathy pain location fields (these might already exist, but adding IF NOT EXISTS for safety)
ALTER TABLE mpox_demographics ADD COLUMN IF NOT EXISTS lymph_pain_location TEXT;
ALTER TABLE mpox_demographics ADD COLUMN IF NOT EXISTS lymph_pain_other_detail TEXT;

-- Add other detail fields for various dropdowns
ALTER TABLE mpox_demographics ADD COLUMN IF NOT EXISTS lymph_other_detail TEXT;
ALTER TABLE mpox_demographics ADD COLUMN IF NOT EXISTS lymph_location TEXT;

-- Add site of first encounter other detail
ALTER TABLE mpox_demographics ADD COLUMN IF NOT EXISTS site_of_first_encounter_other TEXT;

-- Add sex of partners other detail
ALTER TABLE mpox_demographics ADD COLUMN IF NOT EXISTS sex_of_partners_other TEXT;

-- Add recent travel details
ALTER TABLE mpox_demographics ADD COLUMN IF NOT EXISTS recent_travel_details TEXT;

-- Add exposure notes
ALTER TABLE mpox_demographics ADD COLUMN IF NOT EXISTS exposure_notes TEXT;

-- Add progression other detail
ALTER TABLE mpox_demographics ADD COLUMN IF NOT EXISTS progression_other TEXT;

-- Add lesion other details
ALTER TABLE mpox_demographics ADD COLUMN IF NOT EXISTS lesion_other TEXT;
ALTER TABLE mpox_demographics ADD COLUMN IF NOT EXISTS lesion_specify_where TEXT;

-- Add lesion type other detail
ALTER TABLE mpox_demographics ADD COLUMN IF NOT EXISTS type_other TEXT;

-- Add pain description
ALTER TABLE mpox_demographics ADD COLUMN IF NOT EXISTS pain_description TEXT;

-- Add comorbidity specification fields
ALTER TABLE mpox_demographics ADD COLUMN IF NOT EXISTS comorb_neoplasm_specify TEXT;
ALTER TABLE mpox_demographics ADD COLUMN IF NOT EXISTS comorb_immunosuppressants_specify TEXT;
ALTER TABLE mpox_demographics ADD COLUMN IF NOT EXISTS comorb_immunosuppressive_condition_specify TEXT;

-- Add STI other detail
ALTER TABLE mpox_demographics ADD COLUMN IF NOT EXISTS sti_other_detail TEXT;

-- Add lab other detail
ALTER TABLE mpox_demographics ADD COLUMN IF NOT EXISTS lab_other TEXT;

-- Add data entrant name
ALTER TABLE mpox_demographics ADD COLUMN IF NOT EXISTS data_entrant_name TEXT;

-- Update suspect_confirmed_case to be more specific
-- Note: This field already exists, but we're ensuring it can handle the new dropdown values
COMMENT ON COLUMN mpox_demographics.suspect_confirmed_case IS 'Values: Suspect, Confirmed, or empty'; 