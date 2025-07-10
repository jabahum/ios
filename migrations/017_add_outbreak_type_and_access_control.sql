-- Migration: Add Outbreak Type and Enhanced Access Control
-- This migration adds outbreak type classification and enhanced access control

-- Add outbreak_type column to outbreaks table
ALTER TABLE outbreaks ADD COLUMN IF NOT EXISTS outbreak_type VARCHAR(50) DEFAULT 'general';
ALTER TABLE outbreaks ADD COLUMN IF NOT EXISTS outbreak_category VARCHAR(50) DEFAULT 'general';

-- Create index for outbreak type queries
CREATE INDEX IF NOT EXISTS idx_outbreaks_type ON outbreaks(outbreak_type);
CREATE INDEX IF NOT EXISTS idx_outbreaks_category ON outbreaks(outbreak_category);

-- Update existing outbreaks to have proper types
UPDATE outbreaks SET outbreak_type = 'vhf', outbreak_category = 'vhf' WHERE name ILIKE '%ebola%' OR name ILIKE '%vhf%';
UPDATE outbreaks SET outbreak_type = 'mpox', outbreak_category = 'mpox' WHERE name ILIKE '%mpox%' OR name ILIKE '%monkeypox%';

-- Create outbreak-specific roles
INSERT INTO roles (name, description) VALUES 
('vhf_lab_technician', 'VHF laboratory technician - can access VHF list and capture lab requests'),
('vhf_data_entry', 'VHF data entry specialist - can access VHF cases and enter data'),
('mpox_case_manager', 'MPOX case manager - can manage MPOX outbreak cases'),
('mpox_data_entry', 'MPOX data entry specialist - can enter MPOX case data'),
('outbreak_viewer', 'Outbreak viewer - can view assigned outbreaks but not edit/close'),
('outbreak_manager', 'Outbreak manager - can manage assigned outbreaks including edit/close')
ON CONFLICT (name) DO NOTHING;

-- Create outbreak-specific permissions
INSERT INTO permissions (name, description, resource, action) VALUES 
-- VHF-specific permissions
('Access VHF List', 'Access VHF case list', 'vhf_cases', 'read'),
('Capture VHF Lab Requests', 'Capture VHF laboratory requests', 'vhf_lab', 'create'),
('View VHF Lab Results', 'View VHF laboratory results', 'vhf_lab', 'read'),
('Update VHF Lab Results', 'Update VHF laboratory results', 'vhf_lab', 'update'),

-- MPOX-specific permissions
('Access MPOX Outbreak', 'Access MPOX outbreak data', 'mpox_outbreak', 'read'),
('Create MPOX Cases', 'Create new MPOX cases', 'mpox_cases', 'create'),
('View MPOX Cases', 'View MPOX case data', 'mpox_cases', 'read'),
('Update MPOX Cases', 'Update MPOX case data', 'mpox_cases', 'update'),
('Manage MPOX Follow-up', 'Manage MPOX daily follow-up', 'mpox_followup', 'manage'),

-- Outbreak management permissions
('View Assigned Outbreaks', 'View outbreaks assigned to user', 'outbreaks', 'read_assigned'),
('Manage Assigned Outbreaks', 'Manage outbreaks assigned to user', 'outbreaks', 'manage_assigned'),
('Edit Outbreaks', 'Edit outbreak details', 'outbreaks', 'update'),
('Close Outbreaks', 'Close outbreaks', 'outbreaks', 'close'),
('Assign Users to Outbreaks', 'Assign users to specific outbreaks', 'outbreak_assignments', 'create')
ON CONFLICT (resource, action) DO NOTHING;

-- Assign permissions to VHF roles
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.name = 'vhf_lab_technician' AND p.resource IN ('vhf_cases', 'vhf_lab')
ON CONFLICT (role_id, permission_id) DO NOTHING;

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.name = 'vhf_data_entry' AND p.resource IN ('vhf_cases', 'vhf_lab')
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- Assign permissions to MPOX roles
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.name = 'mpox_case_manager' AND p.resource IN ('mpox_outbreak', 'mpox_cases', 'mpox_followup')
ON CONFLICT (role_id, permission_id) DO NOTHING;

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.name = 'mpox_data_entry' AND p.resource IN ('mpox_cases', 'mpox_followup')
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- Assign permissions to outbreak management roles
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.name = 'outbreak_viewer' AND p.resource = 'outbreaks' AND p.action = 'read_assigned'
ON CONFLICT (role_id, permission_id) DO NOTHING;

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.name = 'outbreak_manager' AND p.resource IN ('outbreaks', 'outbreak_assignments')
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- Add password change permission to all new roles
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.name IN ('vhf_lab_technician', 'vhf_data_entry', 'mpox_case_manager', 'mpox_data_entry', 'outbreak_viewer', 'outbreak_manager') 
AND p.resource = 'password' AND p.action = 'change_own'
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- Create function to check user outbreak access
CREATE OR REPLACE FUNCTION check_user_outbreak_access(user_id INTEGER, outbreak_id INTEGER)
RETURNS BOOLEAN AS $$
DECLARE
    has_access BOOLEAN := FALSE;
    user_role VARCHAR(50);
    outbreak_type VARCHAR(50);
BEGIN
    -- Check if user is assigned to this outbreak
    SELECT EXISTS(SELECT 1 FROM user_outbreaks uo WHERE uo.user_id = $1 AND uo.outbreak_id = $2 AND uo.is_active = true)
    INTO has_access;
    
    IF has_access THEN
        RETURN TRUE;
    END IF;
    
    -- Check role-based access
    SELECT r.name, o.outbreak_type 
    INTO user_role, outbreak_type
    FROM users u
    JOIN user_roles ur ON u.user_id = ur.user_id
    JOIN roles r ON ur.role_id = r.id
    CROSS JOIN outbreaks o
    WHERE u.user_id = $1 AND o.id = $2;
    
    -- VHF roles can access VHF outbreaks
    IF user_role IN ('vhf_lab_technician', 'vhf_data_entry') AND outbreak_type = 'vhf' THEN
        RETURN TRUE;
    END IF;
    
    -- MPOX roles can access MPOX outbreaks
    IF user_role IN ('mpox_case_manager', 'mpox_data_entry') AND outbreak_type = 'mpox' THEN
        RETURN TRUE;
    END IF;
    
    -- Admin roles can access all outbreaks
    IF user_role IN ('super_admin', 'admin') THEN
        RETURN TRUE;
    END IF;
    
    RETURN FALSE;
END;
$$ LANGUAGE plpgsql;

-- Create function to get user accessible outbreaks
CREATE OR REPLACE FUNCTION get_user_accessible_outbreaks(user_id INTEGER)
RETURNS TABLE (
    id INTEGER,
    name VARCHAR(255),
    description TEXT,
    outbreak_type VARCHAR(50),
    outbreak_category VARCHAR(50),
    start_date TIMESTAMP,
    end_date TIMESTAMP,
    status VARCHAR(50)
) AS $$
BEGIN
    RETURN QUERY
    SELECT DISTINCT o.id, o.name, o.description, o.outbreak_type, o.outbreak_category, 
           o.start_date, o.end_date, o.status
    FROM outbreaks o
    WHERE (
        -- User is assigned to this outbreak
        EXISTS(SELECT 1 FROM user_outbreaks uo WHERE uo.user_id = $1 AND uo.outbreak_id = o.id AND uo.is_active = true)
        OR
        -- User has role-based access
        EXISTS(
            SELECT 1 FROM users u
            JOIN user_roles ur ON u.user_id = ur.user_id
            JOIN roles r ON ur.role_id = r.id
            WHERE u.user_id = $1 AND (
                (r.name IN ('vhf_lab_technician', 'vhf_data_entry') AND o.outbreak_type = 'vhf')
                OR (r.name IN ('mpox_case_manager', 'mpox_data_entry') AND o.outbreak_type = 'mpox')
                OR r.name IN ('super_admin', 'admin')
            )
        )
    )
    AND o.status != 'closed'
    ORDER BY o.start_date DESC;
END;
$$ LANGUAGE plpgsql;

-- Create function to check if user can edit/close outbreak
CREATE OR REPLACE FUNCTION can_user_manage_outbreak(user_id INTEGER, outbreak_id INTEGER)
RETURNS BOOLEAN AS $$
DECLARE
    user_role VARCHAR(50);
BEGIN
    -- Get user's highest privilege role
    SELECT r.name INTO user_role
    FROM users u
    JOIN user_roles ur ON u.user_id = ur.user_id
    JOIN roles r ON ur.role_id = r.id
    WHERE u.user_id = $1
    ORDER BY CASE r.name 
        WHEN 'super_admin' THEN 1
        WHEN 'admin' THEN 2
        WHEN 'outbreak_manager' THEN 3
        ELSE 4
    END
    LIMIT 1;
    
    -- Super admin and admin can manage all outbreaks
    IF user_role IN ('super_admin', 'admin') THEN
        RETURN TRUE;
    END IF;
    
    -- Outbreak managers can manage outbreaks they're assigned to
    IF user_role = 'outbreak_manager' THEN
        RETURN EXISTS(SELECT 1 FROM user_outbreaks uo WHERE uo.user_id = $1 AND uo.outbreak_id = $2 AND uo.is_active = true);
    END IF;
    
    RETURN FALSE;
END;
$$ LANGUAGE plpgsql;

-- Insert sample outbreaks for testing
INSERT INTO outbreaks (name, description, start_date, status, outbreak_type, outbreak_category) VALUES 
('VHF Ebola 2025', 'Ebola outbreak in 2025', CURRENT_TIMESTAMP, 'active', 'vhf', 'vhf'),
('MPOX Outbreak 2024', 'Monkeypox outbreak in 2024', CURRENT_TIMESTAMP, 'active', 'mpox', 'mpox'); 