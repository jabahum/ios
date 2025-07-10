-- Fix VARCHAR(8) constraint issues in vhf_patients table
-- This migration addresses the "value too long for type character varying(8)" error

-- Update gender column to allow longer values
ALTER TABLE vhf_patients ALTER COLUMN gender TYPE VARCHAR(50);

-- Update other potentially problematic columns
ALTER TABLE vhf_patients ALTER COLUMN status TYPE VARCHAR(50);

-- Ensure location columns have sufficient length
ALTER TABLE vhf_patients ALTER COLUMN village_town TYPE VARCHAR(255);
ALTER TABLE vhf_patients ALTER COLUMN parish TYPE VARCHAR(255);
ALTER TABLE vhf_patients ALTER COLUMN subcounty TYPE VARCHAR(255);
ALTER TABLE vhf_patients ALTER COLUMN district TYPE VARCHAR(255);
ALTER TABLE vhf_patients ALTER COLUMN country_of_residence TYPE VARCHAR(255);
ALTER TABLE vhf_patients ALTER COLUMN occupation TYPE VARCHAR(255);
ALTER TABLE vhf_patients ALTER COLUMN ill_village_town TYPE VARCHAR(255);
ALTER TABLE vhf_patients ALTER COLUMN ill_subcounty TYPE VARCHAR(255);
ALTER TABLE vhf_patients ALTER COLUMN ill_district TYPE VARCHAR(255);

-- Update phone number columns to allow longer values
ALTER TABLE vhf_patients ALTER COLUMN patient_phone TYPE VARCHAR(50);
ALTER TABLE vhf_patients ALTER COLUMN next_of_kin_phone TYPE VARCHAR(50);
ALTER TABLE vhf_patients ALTER COLUMN data_capturer_phone TYPE VARCHAR(50);

-- Update name columns to allow longer values
ALTER TABLE vhf_patients ALTER COLUMN surname TYPE VARCHAR(255);
ALTER TABLE vhf_patients ALTER COLUMN other_names TYPE VARCHAR(255);
ALTER TABLE vhf_patients ALTER COLUMN phone_owner TYPE VARCHAR(255);
ALTER TABLE vhf_patients ALTER COLUMN next_of_kin TYPE VARCHAR(255);
ALTER TABLE vhf_patients ALTER COLUMN data_capturer_name TYPE VARCHAR(255);
ALTER TABLE vhf_patients ALTER COLUMN head_of_household TYPE VARCHAR(255);
ALTER TABLE vhf_patients ALTER COLUMN reporting_health_facility_name TYPE VARCHAR(255); 