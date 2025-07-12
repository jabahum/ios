-- Migration: Migrate facilities data and add VHF filtering support
-- This migration migrates data from afi_facilities to facility table and adds missing columns

-- Add missing columns to facility table first
ALTER TABLE facility ADD COLUMN IF NOT EXISTS facility_code VARCHAR(50);
ALTER TABLE facility ADD COLUMN IF NOT EXISTS facility_type VARCHAR(100);
ALTER TABLE facility ADD COLUMN IF NOT EXISTS district VARCHAR(100);
ALTER TABLE facility ADD COLUMN IF NOT EXISTS subcounty VARCHAR(100);
ALTER TABLE facility ADD COLUMN IF NOT EXISTS parish VARCHAR(100);
ALTER TABLE facility ADD COLUMN IF NOT EXISTS village VARCHAR(100);
ALTER TABLE facility ADD COLUMN IF NOT EXISTS latitude DECIMAL(10, 8);
ALTER TABLE facility ADD COLUMN IF NOT EXISTS longitude DECIMAL(11, 8);
ALTER TABLE facility ADD COLUMN IF NOT EXISTS contact_person VARCHAR(100);
ALTER TABLE facility ADD COLUMN IF NOT EXISTS contact_phone VARCHAR(20);
ALTER TABLE facility ADD COLUMN IF NOT EXISTS contact_email VARCHAR(100);
ALTER TABLE facility ADD COLUMN IF NOT EXISTS ownership VARCHAR(50);
ALTER TABLE facility ADD COLUMN IF NOT EXISTS hsd VARCHAR(100);
ALTER TABLE facility ADD COLUMN IF NOT EXISTS is_active BOOLEAN DEFAULT true;
ALTER TABLE facility ADD COLUMN IF NOT EXISTS created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE facility ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP;

-- First, let's check if afi_facilities table exists and has data
DO $$
BEGIN
    IF EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'afi_facilities') THEN
        -- Migrate data from afi_facilities to facility table
        -- Use a temporary table approach to avoid conflicts
        CREATE TEMP TABLE temp_facilities AS
        SELECT DISTINCT
            COALESCE(facility_name, 'Unknown Facility') as facility_name,
            CASE 
                WHEN level = 'HC I' THEN 1
                WHEN level = 'HC II' THEN 2
                WHEN level = 'HC III' THEN 3
                WHEN level = 'HC IV' THEN 4
                WHEN level = 'Hospital' THEN 5
                WHEN level = 'Regional Referral Hospital' THEN 6
                WHEN level = 'National Referral Hospital' THEN 7
                WHEN level = 'Clinic' THEN 1
                WHEN level = 'Health Centre' THEN 2
                WHEN level = 'Medical Centre' THEN 3
                ELSE 1
            END as facility_level,
            hsd,
            subcounty_town_council_division,
            ownership,
            phone_number
        FROM afi_facilities
        WHERE facility_name IS NOT NULL AND facility_name != '';
        
        -- Insert only facilities that don't already exist
        INSERT INTO facility (facility_name, facility_level)
        SELECT tf.facility_name, tf.facility_level
        FROM temp_facilities tf
        WHERE NOT EXISTS (
            SELECT 1 FROM facility f 
            WHERE LOWER(TRIM(f.facility_name)) = LOWER(TRIM(tf.facility_name))
        );
        
        -- Update existing facilities with additional data
        UPDATE facility 
        SET 
            district = tf.hsd,
            subcounty = tf.subcounty_town_council_division,
            ownership = tf.ownership,
            contact_phone = tf.phone_number,
            hsd = tf.hsd
        FROM temp_facilities tf
        WHERE LOWER(TRIM(facility.facility_name)) = LOWER(TRIM(tf.facility_name))
          AND (facility.district IS NULL OR facility.subcounty IS NULL);
        
        DROP TABLE temp_facilities;
        
        RAISE NOTICE 'Migrated facilities from afi_facilities to facility table';
    ELSE
        RAISE NOTICE 'afi_facilities table does not exist, skipping migration';
    END IF;
END $$;

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_facility_district ON facility(district);
CREATE INDEX IF NOT EXISTS idx_facility_is_active ON facility(is_active);
CREATE INDEX IF NOT EXISTS idx_facility_name ON facility(facility_name);
CREATE INDEX IF NOT EXISTS idx_facility_hsd ON facility(hsd);

-- Add facility_id column to vhf_patients table if it doesn't exist
ALTER TABLE vhf_patients ADD COLUMN IF NOT EXISTS facility_id INTEGER REFERENCES facility(facility_id);

-- Create index for facility filtering
CREATE INDEX IF NOT EXISTS idx_vhf_patients_facility_id ON vhf_patients(facility_id);

-- Add facility_id column to employee table if it doesn't exist (it should already exist based on previous migrations)
-- This is just to ensure it exists
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'employee' AND column_name = 'facility') THEN
        ALTER TABLE employee ADD COLUMN facility INTEGER REFERENCES facility(facility_id);
    END IF;
END $$;

-- Create index for employee facility lookups
CREATE INDEX IF NOT EXISTS idx_employee_facility ON employee(facility);

-- Update VHF patients to link to facilities based on reporting_health_facility_name
UPDATE vhf_patients 
SET facility_id = f.facility_id
FROM facility f
WHERE vhf_patients.reporting_health_facility_name IS NOT NULL 
  AND vhf_patients.reporting_health_facility_name != ''
  AND LOWER(TRIM(vhf_patients.reporting_health_facility_name)) = LOWER(TRIM(f.facility_name))
  AND vhf_patients.facility_id IS NULL;

-- Create a function to get user's facility ID
CREATE OR REPLACE FUNCTION get_user_facility_id(user_id_param INTEGER)
RETURNS INTEGER AS $$
DECLARE
    facility_id_result INTEGER;
BEGIN
    SELECT e.facility INTO facility_id_result
    FROM employee e
    JOIN users u ON e.employee_email = u.email
    WHERE u.user_id = user_id_param
    LIMIT 1;
    
    RETURN facility_id_result;
END;
$$ LANGUAGE plpgsql;

-- Create a function to check if user has facility-based access
CREATE OR REPLACE FUNCTION user_has_facility_access(user_id_param INTEGER, target_facility_id INTEGER)
RETURNS BOOLEAN AS $$
DECLARE
    user_facility_id INTEGER;
    user_role_id INTEGER;
BEGIN
    -- Get user's facility ID
    user_facility_id := get_user_facility_id(user_id_param);
    
    -- If user has no facility assigned, they have access to all facilities
    IF user_facility_id IS NULL THEN
        RETURN TRUE;
    END IF;
    
    -- If user has a facility assigned, check if it matches the target facility
    RETURN user_facility_id = target_facility_id;
END;
$$ LANGUAGE plpgsql; 