-- Migration: Add Reports Role and Permissions
-- This migration adds a dedicated reports role with appropriate permissions

-- Add the reports role
INSERT INTO roles (name, description) VALUES 
('reports', 'Reports and analytics specialist with comprehensive reporting access')
ON CONFLICT (name) DO NOTHING;

-- Add additional permissions for reports functionality
INSERT INTO permissions (name, description, resource, action) VALUES 
-- Reports permissions (enhanced)
('View All Reports', 'View all system reports and analytics', 'reports', 'read'),
('Export All Reports', 'Export all report data in various formats', 'reports', 'export'),
('Create Custom Reports', 'Create custom reports and dashboards', 'reports', 'create'),
('Manage Report Templates', 'Manage report templates and layouts', 'reports', 'manage'),

-- CIF-specific report permissions
('View VHF CIF Reports', 'View VHF Case Investigation Form reports', 'vhf_cif_reports', 'read'),
('View Measles CIF Reports', 'View Measles Case Investigation Form reports', 'measles_cif_reports', 'read'),
('View Polio CIF Reports', 'View Polio Case Investigation Form reports', 'polio_cif_reports', 'read'),
('View Mpox CIF Reports', 'View Mpox Case Investigation Form reports', 'mpox_cif_reports', 'read'),

-- Analytics permissions
('View Analytics Dashboard', 'View analytics dashboard and metrics', 'analytics', 'read'),
('Export Analytics Data', 'Export analytics data and charts', 'analytics', 'export'),
('Create Analytics Reports', 'Create custom analytics reports', 'analytics', 'create'),

-- Data access permissions for reports
('View Patient Data for Reports', 'View patient data for reporting purposes', 'patients', 'read'),
('View Laboratory Data for Reports', 'View laboratory data for reporting purposes', 'laboratory', 'read'),
('View Treatment Data for Reports', 'View treatment data for reporting purposes', 'treatment', 'read'),
('View Surveillance Data for Reports', 'View surveillance data for reporting purposes', 'surveillance', 'read'),

-- Geographic and demographic permissions
('View Geographic Reports', 'View geographic distribution reports', 'geographic_reports', 'read'),
('View Demographic Reports', 'View demographic analysis reports', 'demographic_reports', 'read'),
('Export Geographic Data', 'Export geographic data for mapping', 'geographic_data', 'export'),
('Export Demographic Data', 'Export demographic data for analysis', 'demographic_data', 'export')
ON CONFLICT (resource, action) DO NOTHING;

-- Assign permissions to the reports role
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.name = 'reports' AND p.resource IN ('reports', 'analytics', 'vhf_cif_reports', 'measles_cif_reports', 'polio_cif_reports', 'mpox_cif_reports', 'geographic_reports', 'demographic_reports', 'geographic_data', 'demographic_data')
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- Also give reports role read access to core data for reporting purposes
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.name = 'reports' AND p.resource IN ('patients', 'laboratory', 'treatment', 'surveillance') AND p.action = 'read'
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- Add VHF Lab Technician role if it doesn't exist
INSERT INTO roles (name, description) VALUES 
('vhf_lab_technician', 'VHF Laboratory Technician with district-specific CIF access')
ON CONFLICT (name) DO NOTHING;

-- Add permissions for VHF Lab Technician
INSERT INTO permissions (name, description, resource, action) VALUES 
('View District VHF CIF Reports', 'View VHF CIF reports for assigned district only', 'district_vhf_cif_reports', 'read'),
('Export District VHF CIF Data', 'Export VHF CIF data for assigned district only', 'district_vhf_cif_data', 'export')
ON CONFLICT (resource, action) DO NOTHING;

-- Assign district-specific permissions to VHF Lab Technician
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.name = 'vhf_lab_technician' AND p.resource IN ('district_vhf_cif_reports', 'district_vhf_cif_data')
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- Add Case Manager role if it doesn't exist
INSERT INTO roles (name, description) VALUES 
('case_manager', 'Case Manager with case tracking and management access')
ON CONFLICT (name) DO NOTHING;

-- Add permissions for Case Manager
INSERT INTO permissions (name, description, resource, action) VALUES 
('View Case Management Reports', 'View case management and tracking reports', 'case_management_reports', 'read'),
('Export Case Management Data', 'Export case management data', 'case_management_data', 'export'),
('Update Case Status', 'Update case status and progress', 'cases', 'update')
ON CONFLICT (resource, action) DO NOTHING;

-- Assign case management permissions to Case Manager
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.name = 'case_manager' AND p.resource IN ('case_management_reports', 'case_management_data', 'cases')
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- Add Data Analyst role if it doesn't exist
INSERT INTO roles (name, description) VALUES 
('data_analyst', 'Data Analyst with advanced analytics and export capabilities')
ON CONFLICT (name) DO NOTHING;

-- Add permissions for Data Analyst
INSERT INTO permissions (name, description, resource, action) VALUES 
('View Advanced Analytics', 'View advanced analytics and trend analysis', 'advanced_analytics', 'read'),
('Create Advanced Reports', 'Create advanced analytical reports', 'advanced_reports', 'create'),
('Export Advanced Analytics', 'Export advanced analytics data', 'advanced_analytics', 'export'),
('Manage Analytics Dashboards', 'Manage and configure analytics dashboards', 'analytics_dashboards', 'manage')
ON CONFLICT (resource, action) DO NOTHING;

-- Assign advanced analytics permissions to Data Analyst
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.name = 'data_analyst' AND p.resource IN ('advanced_analytics', 'advanced_reports', 'analytics_dashboards')
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- Also give data analyst access to all basic reports
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.name = 'data_analyst' AND p.resource = 'reports'
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- Create a function to check if user has reports access
CREATE OR REPLACE FUNCTION user_has_reports_access(user_id_param INTEGER)
RETURNS BOOLEAN AS $$
DECLARE
    has_access BOOLEAN := FALSE;
BEGIN
    SELECT EXISTS(
        SELECT 1 
        FROM user_roles ur
        JOIN roles r ON ur.role_id = r.id
        WHERE ur.user_id = user_id_param 
        AND r.name IN ('reports', 'data_analyst', 'admin', 'super_admin', 'vhf_lab_technician', 'case_manager')
        AND r.is_active = true
    ) INTO has_access;
    
    RETURN has_access;
END;
$$ LANGUAGE plpgsql;

-- Create a function to get user's report access level
CREATE OR REPLACE FUNCTION get_user_report_access_level(user_id_param INTEGER)
RETURNS TEXT AS $$
DECLARE
    access_level TEXT := 'none';
    role_name TEXT;
BEGIN
    -- Check for highest level access first
    SELECT r.name INTO role_name
    FROM user_roles ur
    JOIN roles r ON ur.role_id = r.id
    WHERE ur.user_id = user_id_param 
    AND r.name IN ('super_admin', 'admin', 'data_analyst', 'reports', 'case_manager', 'vhf_lab_technician')
    AND r.is_active = true
    ORDER BY 
        CASE r.name 
            WHEN 'super_admin' THEN 1
            WHEN 'admin' THEN 2
            WHEN 'data_analyst' THEN 3
            WHEN 'reports' THEN 4
            WHEN 'case_manager' THEN 5
            WHEN 'vhf_lab_technician' THEN 6
            ELSE 7
        END
    LIMIT 1;
    
    CASE role_name
        WHEN 'super_admin', 'admin' THEN access_level := 'full';
        WHEN 'data_analyst' THEN access_level := 'analytics';
        WHEN 'reports' THEN access_level := 'reports';
        WHEN 'case_manager' THEN access_level := 'case_management';
        WHEN 'vhf_lab_technician' THEN access_level := 'district_vhf';
        ELSE access_level := 'none';
    END CASE;
    
    RETURN access_level;
END;
$$ LANGUAGE plpgsql;

-- Create indexes for better performance on role lookups
CREATE INDEX IF NOT EXISTS idx_user_roles_user_id_active ON user_roles(user_id);
CREATE INDEX IF NOT EXISTS idx_roles_name_active ON roles(name, is_active);

-- Add comments for documentation
COMMENT ON FUNCTION user_has_reports_access(INTEGER) IS 'Check if a user has access to reports functionality';
COMMENT ON FUNCTION get_user_report_access_level(INTEGER) IS 'Get the report access level for a user (full, analytics, reports, case_management, district_vhf, none)';
COMMENT ON TABLE roles IS 'Roles table with enhanced reporting roles';
COMMENT ON TABLE permissions IS 'Permissions table with comprehensive reporting permissions'; 