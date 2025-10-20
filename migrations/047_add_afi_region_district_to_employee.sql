-- Migration: Add afi_region and afi_district columns to employee table
-- This allows employees to be assigned to specific regions or districts for access control

-- Add afi_region column to employee table
ALTER TABLE employee ADD COLUMN IF NOT EXISTS afi_region TEXT;

-- Add afi_district column to employee table  
ALTER TABLE employee ADD COLUMN IF NOT EXISTS afi_district TEXT;

-- Add comments for documentation
COMMENT ON COLUMN employee.afi_region IS 'AFI region assignment for access control - allows regional-level access to VHF cases';
COMMENT ON COLUMN employee.afi_district IS 'AFI district assignment for access control - allows district-level access to VHF cases';

-- Create index for better query performance
CREATE INDEX IF NOT EXISTS idx_employee_afi_region ON employee(afi_region);
CREATE INDEX IF NOT EXISTS idx_employee_afi_district ON employee(afi_district);
