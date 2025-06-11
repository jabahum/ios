-- Add case_code column to vhf_patients table
ALTER TABLE vhf_patients ADD COLUMN IF NOT EXISTS case_code VARCHAR(50) UNIQUE;

-- Add other missing columns
ALTER TABLE vhf_patients ADD COLUMN IF NOT EXISTS data_capturer_name VARCHAR(100);
ALTER TABLE vhf_patients ADD COLUMN IF NOT EXISTS data_capturer_phone VARCHAR(20);
ALTER TABLE vhf_patients ADD COLUMN IF NOT EXISTS reporting_health_facility_name VARCHAR(200); 