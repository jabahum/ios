-- Migration: Add Outbreak Assignment and Patient Management Roles
-- This migration adds outbreak assignment functionality and patient management roles

-- Create user_outbreaks junction table for outbreak assignment
CREATE TABLE IF NOT EXISTS user_outbreaks (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    outbreak_id INTEGER NOT NULL, -- Will reference your outbreak table
    assigned_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    assigned_by INTEGER REFERENCES users(user_id),
    is_active BOOLEAN DEFAULT true,
    UNIQUE(user_id, outbreak_id)
);

-- Create patient management roles table
CREATE TABLE IF NOT EXISTS patient_management_roles (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    role_type VARCHAR(50) NOT NULL, -- 'registration', 'admission', 'discharge', 'care'
    outbreak_id INTEGER, -- NULL means all outbreaks
    facility_id INTEGER, -- NULL means all facilities
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_by INTEGER REFERENCES users(user_id),
    UNIQUE(user_id, role_type, outbreak_id, facility_id)
);

-- Create password change requests table
CREATE TABLE IF NOT EXISTS password_change_requests (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    request_token VARCHAR(255) NOT NULL UNIQUE,
    current_password_hash VARCHAR(255) NOT NULL,
    new_password_hash VARCHAR(255) NOT NULL,
    new_password_salt VARCHAR(255) NOT NULL,
    is_approved BOOLEAN DEFAULT false,
    approved_by INTEGER REFERENCES users(user_id),
    approved_at TIMESTAMP WITH TIME ZONE,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    ip_address VARCHAR(45),
    user_agent TEXT
);

-- Enhance employee table with additional fields
ALTER TABLE employee ADD COLUMN IF NOT EXISTS employee_code VARCHAR(20) UNIQUE;
ALTER TABLE employee ADD COLUMN IF NOT EXISTS employee_title VARCHAR(100);
ALTER TABLE employee ADD COLUMN IF NOT EXISTS employee_department VARCHAR(100);
ALTER TABLE employee ADD COLUMN IF NOT EXISTS employee_supervisor INTEGER REFERENCES employee(employee_id);
ALTER TABLE employee ADD COLUMN IF NOT EXISTS employee_start_date DATE;
ALTER TABLE employee ADD COLUMN IF NOT EXISTS employee_end_date DATE;
ALTER TABLE employee ADD COLUMN IF NOT EXISTS employee_status VARCHAR(20) DEFAULT 'active';
ALTER TABLE employee ADD COLUMN IF NOT EXISTS employee_photo_url VARCHAR(255);
ALTER TABLE employee ADD COLUMN IF NOT EXISTS employee_notes TEXT;
ALTER TABLE employee ADD COLUMN IF NOT EXISTS created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE employee ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE employee ADD COLUMN IF NOT EXISTS created_by INTEGER REFERENCES users(user_id);
ALTER TABLE employee ADD COLUMN IF NOT EXISTS updated_by INTEGER REFERENCES users(user_id);

-- Create employee_departments table
CREATE TABLE IF NOT EXISTS employee_departments (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create employee_titles table
CREATE TABLE IF NOT EXISTS employee_titles (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create employee_assignments table for flexible assignments
CREATE TABLE IF NOT EXISTS employee_assignments (
    id SERIAL PRIMARY KEY,
    employee_id INTEGER NOT NULL REFERENCES employee(employee_id) ON DELETE CASCADE,
    assignment_type VARCHAR(50) NOT NULL, -- 'facility', 'outbreak', 'department', 'project'
    assignment_id INTEGER NOT NULL, -- ID of the facility, outbreak, department, etc.
    assignment_name VARCHAR(255) NOT NULL, -- Name for display purposes
    start_date DATE NOT NULL,
    end_date DATE,
    is_primary BOOLEAN DEFAULT false,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_by INTEGER REFERENCES users(user_id)
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_user_outbreaks_user_id ON user_outbreaks(user_id);
CREATE INDEX IF NOT EXISTS idx_user_outbreaks_outbreak_id ON user_outbreaks(outbreak_id);
CREATE INDEX IF NOT EXISTS idx_user_outbreaks_active ON user_outbreaks(is_active);

CREATE INDEX IF NOT EXISTS idx_patient_management_roles_user_id ON patient_management_roles(user_id);
CREATE INDEX IF NOT EXISTS idx_patient_management_roles_role_type ON patient_management_roles(role_type);
CREATE INDEX IF NOT EXISTS idx_patient_management_roles_outbreak_id ON patient_management_roles(outbreak_id);
CREATE INDEX IF NOT EXISTS idx_patient_management_roles_facility_id ON patient_management_roles(facility_id);

CREATE INDEX IF NOT EXISTS idx_password_change_requests_user_id ON password_change_requests(user_id);
CREATE INDEX IF NOT EXISTS idx_password_change_requests_token ON password_change_requests(request_token);
CREATE INDEX IF NOT EXISTS idx_password_change_requests_expires_at ON password_change_requests(expires_at);

CREATE INDEX IF NOT EXISTS idx_employee_code ON employee(employee_code);
CREATE INDEX IF NOT EXISTS idx_employee_status ON employee(employee_status);
CREATE INDEX IF NOT EXISTS idx_employee_department ON employee(employee_department);
CREATE INDEX IF NOT EXISTS idx_employee_supervisor ON employee(employee_supervisor);

CREATE INDEX IF NOT EXISTS idx_employee_assignments_employee_id ON employee_assignments(employee_id);
CREATE INDEX IF NOT EXISTS idx_employee_assignments_type_id ON employee_assignments(assignment_type, assignment_id);
CREATE INDEX IF NOT EXISTS idx_employee_assignments_active ON employee_assignments(is_active);

-- Insert default employee departments
INSERT INTO employee_departments (name, description) VALUES 
('Clinical Services', 'Clinical care and patient management'),
('Laboratory Services', 'Laboratory testing and diagnostics'),
('Surveillance', 'Disease surveillance and monitoring'),
('Administration', 'Administrative and management services'),
('Support Services', 'Support and auxiliary services'),
('Research', 'Research and development activities')
ON CONFLICT (name) DO NOTHING;

-- Insert default employee titles
INSERT INTO employee_titles (name, description) VALUES 
('Medical Officer', 'Medical doctor providing clinical care'),
('Nurse', 'Registered nurse providing patient care'),
('Laboratory Technician', 'Laboratory testing and analysis'),
('Surveillance Officer', 'Disease surveillance and monitoring'),
('Data Entry Clerk', 'Data entry and record management'),
('Administrator', 'Administrative and management role'),
('Support Staff', 'Support and auxiliary services'),
('Research Officer', 'Research and development activities')
ON CONFLICT (name) DO NOTHING;

-- Add new permissions for patient management
INSERT INTO permissions (name, description, resource, action) VALUES 
-- Patient Registration permissions
('Register Patients', 'Register new patients in the system', 'patient_registration', 'create'),
('View Patient Registrations', 'View patient registration records', 'patient_registration', 'read'),
('Update Patient Registrations', 'Update patient registration information', 'patient_registration', 'update'),
('Delete Patient Registrations', 'Delete patient registration records', 'patient_registration', 'delete'),

-- Patient Admission permissions
('Admit Patients', 'Admit patients into care', 'patient_admission', 'create'),
('View Patient Admissions', 'View patient admission records', 'patient_admission', 'read'),
('Update Patient Admissions', 'Update patient admission information', 'patient_admission', 'update'),
('Delete Patient Admissions', 'Delete patient admission records', 'patient_admission', 'delete'),

-- Patient Discharge permissions
('Discharge Patients', 'Discharge patients from care', 'patient_discharge', 'create'),
('View Patient Discharges', 'View patient discharge records', 'patient_discharge', 'read'),
('Update Patient Discharges', 'Update patient discharge information', 'patient_discharge', 'update'),
('Delete Patient Discharges', 'Delete patient discharge records', 'patient_discharge', 'delete'),

-- Patient Care permissions
('Manage Patient Care', 'Manage ongoing patient care', 'patient_care', 'create'),
('View Patient Care', 'View patient care records', 'patient_care', 'read'),
('Update Patient Care', 'Update patient care information', 'patient_care', 'update'),
('Delete Patient Care', 'Delete patient care records', 'patient_care', 'delete'),

-- Outbreak Assignment permissions
('Assign Users to Outbreaks', 'Assign users to specific outbreaks', 'outbreak_assignment', 'create'),
('View Outbreak Assignments', 'View outbreak assignments', 'outbreak_assignment', 'read'),
('Update Outbreak Assignments', 'Update outbreak assignments', 'outbreak_assignment', 'update'),
('Delete Outbreak Assignments', 'Remove outbreak assignments', 'outbreak_assignment', 'delete'),

-- Employee Management permissions
('Create Employees', 'Create new employee records', 'employees', 'create'),
('View Employees', 'View employee records', 'employees', 'read'),
('Update Employees', 'Update employee records', 'employees', 'update'),
('Delete Employees', 'Delete employee records', 'employees', 'delete'),
('Manage Employee Assignments', 'Manage employee assignments', 'employees', 'assign'),

-- Password Management permissions
('Change Own Password', 'Change own password', 'password', 'change_own'),
('Reset User Passwords', 'Reset passwords for other users', 'password', 'reset_others'),
('Approve Password Changes', 'Approve password change requests', 'password', 'approve')
ON CONFLICT (resource, action) DO NOTHING;

-- Create new roles for patient management
INSERT INTO roles (name, description) VALUES 
('patient_registrar', 'Patient registration specialist'),
('patient_admission_officer', 'Patient admission and intake specialist'),
('patient_discharge_officer', 'Patient discharge and follow-up specialist'),
('patient_care_coordinator', 'Patient care coordination and management'),
('outbreak_manager', 'Outbreak management and assignment coordinator'),
('employee_manager', 'Employee management and assignment coordinator')
ON CONFLICT (name) DO NOTHING;

-- Assign permissions to new roles
-- Patient Registrar gets registration permissions
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.name = 'patient_registrar' AND p.resource = 'patient_registration'
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- Patient Admission Officer gets admission permissions
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.name = 'patient_admission_officer' AND p.resource = 'patient_admission'
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- Patient Discharge Officer gets discharge permissions
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.name = 'patient_discharge_officer' AND p.resource = 'patient_discharge'
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- Patient Care Coordinator gets care permissions
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.name = 'patient_care_coordinator' AND p.resource = 'patient_care'
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- Outbreak Manager gets outbreak assignment permissions
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.name = 'outbreak_manager' AND p.resource = 'outbreak_assignment'
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- Employee Manager gets employee management permissions
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.name = 'employee_manager' AND p.resource = 'employees'
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- Add password change permission to all roles
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE p.resource = 'password' AND p.action = 'change_own'
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- Add password reset permission to admin roles
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.name IN ('super_admin', 'admin') AND p.resource = 'password' AND p.action IN ('reset_others', 'approve')
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- Create function to update updated_at timestamp for employee table
CREATE OR REPLACE FUNCTION update_employee_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create trigger for employee table
CREATE TRIGGER update_employee_updated_at BEFORE UPDATE ON employee
    FOR EACH ROW EXECUTE FUNCTION update_employee_updated_at();

-- Create function to generate employee codes
CREATE OR REPLACE FUNCTION generate_employee_code()
RETURNS TRIGGER AS $$
DECLARE
    new_code VARCHAR(20);
    counter INTEGER := 1;
BEGIN
    -- Generate code based on first letter of first name and last name
    new_code := UPPER(LEFT(NEW.employee_fname, 1) || LEFT(NEW.employee_lname, 1) || 
                     TO_CHAR(CURRENT_DATE, 'YY') || LPAD(counter::TEXT, 3, '0'));
    
    -- Check if code exists and increment counter
    WHILE EXISTS (SELECT 1 FROM employee WHERE employee_code = new_code) LOOP
        counter := counter + 1;
        new_code := UPPER(LEFT(NEW.employee_fname, 1) || LEFT(NEW.employee_lname, 1) || 
                         TO_CHAR(CURRENT_DATE, 'YY') || LPAD(counter::TEXT, 3, '0'));
    END LOOP;
    
    NEW.employee_code := new_code;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create trigger to auto-generate employee codes
CREATE TRIGGER generate_employee_code_trigger BEFORE INSERT ON employee
    FOR EACH ROW EXECUTE FUNCTION generate_employee_code();

-- Create view for employee information with assignments (FIXED: using correct table and column names)
CREATE OR REPLACE VIEW employee_info AS
SELECT 
    e.employee_id,
    e.employee_fname,
    e.employee_lname,
    e.employee_sex,
    e.employee_email,
    e.employee_phone,
    e.employee_cadre,
    e.employee_code,
    e.employee_title,
    e.employee_department,
    e.employee_status,
    e.employee_start_date,
    e.employee_end_date,
    e.facility,
    f.facility_name as facility_name,
    sup.employee_fname || ' ' || sup.employee_lname as supervisor_name,
    e.created_at,
    e.updated_at,
    -- Get current assignments
    array_agg(DISTINCT ea.assignment_name) FILTER (WHERE ea.is_active = true) as current_assignments
FROM employee e
LEFT JOIN facility f ON e.facility = f.facility_id
LEFT JOIN employee sup ON e.employee_supervisor = sup.employee_id
LEFT JOIN employee_assignments ea ON e.employee_id = ea.employee_id
GROUP BY e.employee_id, e.employee_fname, e.employee_lname, e.employee_sex, 
         e.employee_email, e.employee_phone, e.employee_cadre, e.employee_code,
         e.employee_title, e.employee_department, e.employee_status,
         e.employee_start_date, e.employee_end_date, e.facility, f.facility_name,
         sup.employee_fname, sup.employee_lname, e.created_at, e.updated_at;

-- Create view for user outbreak assignments (FIXED: using correct column names)
CREATE OR REPLACE VIEW user_outbreak_info AS
SELECT 
    uo.id,
    uo.user_id,
    u.user_name,
    uo.outbreak_id,
    o.name as outbreak_name,
    o.description as outbreak_description,
    uo.assigned_at,
    uo.assigned_by,
    ass.user_name as assigned_by_name,
    uo.is_active
FROM user_outbreaks uo
JOIN users u ON uo.user_id = u.user_id
LEFT JOIN outbreaks o ON uo.outbreak_id = o.id
LEFT JOIN users ass ON uo.assigned_by = ass.user_id;

-- Create view for patient management roles (FIXED: using correct column names and table names)
CREATE OR REPLACE VIEW patient_management_info AS
SELECT 
    pmr.id,
    pmr.user_id,
    u.user_name,
    pmr.role_type,
    pmr.outbreak_id,
    o.name as outbreak_name,
    pmr.facility_id,
    f.facility_name as facility_name,
    pmr.is_active,
    pmr.created_at,
    pmr.created_by,
    c.user_name as created_by_name
FROM patient_management_roles pmr
JOIN users u ON pmr.user_id = u.user_id
LEFT JOIN outbreaks o ON pmr.outbreak_id = o.id
LEFT JOIN facility f ON pmr.facility_id = f.facility_id
LEFT JOIN users c ON pmr.created_by = c.user_id; 