-- Migration: Create RBAC (Role-Based Access Control) System
-- This migration creates a modern user management and access control system

-- Create departments table
CREATE TABLE IF NOT EXISTS departments (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create enhanced users table
CREATE TABLE IF NOT EXISTS enhanced_users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(100) NOT NULL UNIQUE,
    email VARCHAR(255) UNIQUE,
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    password_hash VARCHAR(255) NOT NULL,
    password_salt VARCHAR(255) NOT NULL,
    is_active BOOLEAN DEFAULT true,
    is_locked BOOLEAN DEFAULT false,
    failed_login_attempts INTEGER DEFAULT 0,
    last_login_at TIMESTAMP WITH TIME ZONE,
    password_changed_at TIMESTAMP WITH TIME ZONE,
    password_expires_at TIMESTAMP WITH TIME ZONE,
    department_id INTEGER REFERENCES departments(id),
    employee_id INTEGER,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_by INTEGER,
    updated_by INTEGER
);

-- Create roles table
CREATE TABLE IF NOT EXISTS roles (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create permissions table
CREATE TABLE IF NOT EXISTS permissions (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    resource VARCHAR(100) NOT NULL,  -- e.g., 'vhf_patients', 'users', 'reports'
    action VARCHAR(50) NOT NULL,     -- e.g., 'create', 'read', 'update', 'delete'
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(resource, action)
);

-- Create user_roles junction table
CREATE TABLE IF NOT EXISTS user_roles (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES enhanced_users(id) ON DELETE CASCADE,
    role_id INTEGER NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, role_id)
);

-- Create role_permissions junction table
CREATE TABLE IF NOT EXISTS role_permissions (
    id SERIAL PRIMARY KEY,
    role_id INTEGER NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    permission_id INTEGER NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(role_id, permission_id)
);

-- Create user_sessions table for session management
CREATE TABLE IF NOT EXISTS user_sessions (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES enhanced_users(id) ON DELETE CASCADE,
    session_token VARCHAR(255) NOT NULL UNIQUE,
    ip_address VARCHAR(45),
    user_agent TEXT,
    is_active BOOLEAN DEFAULT true,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create audit_logs table for comprehensive audit trail
CREATE TABLE IF NOT EXISTS audit_logs (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES enhanced_users(id),
    action VARCHAR(100) NOT NULL,
    resource VARCHAR(100),
    resource_id INTEGER,
    details JSONB,
    ip_address VARCHAR(45),
    user_agent TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_enhanced_users_username ON enhanced_users(username);
CREATE INDEX IF NOT EXISTS idx_enhanced_users_email ON enhanced_users(email);
CREATE INDEX IF NOT EXISTS idx_enhanced_users_department ON enhanced_users(department_id);
CREATE INDEX IF NOT EXISTS idx_enhanced_users_employee ON enhanced_users(employee_id);
CREATE INDEX IF NOT EXISTS idx_enhanced_users_active ON enhanced_users(is_active);

CREATE INDEX IF NOT EXISTS idx_roles_name ON roles(name);
CREATE INDEX IF NOT EXISTS idx_roles_active ON roles(is_active);

CREATE INDEX IF NOT EXISTS idx_permissions_resource_action ON permissions(resource, action);
CREATE INDEX IF NOT EXISTS idx_permissions_active ON permissions(is_active);

CREATE INDEX IF NOT EXISTS idx_user_roles_user_id ON user_roles(user_id);
CREATE INDEX IF NOT EXISTS idx_user_roles_role_id ON user_roles(role_id);

CREATE INDEX IF NOT EXISTS idx_role_permissions_role_id ON role_permissions(role_id);
CREATE INDEX IF NOT EXISTS idx_role_permissions_permission_id ON role_permissions(permission_id);

CREATE INDEX IF NOT EXISTS idx_user_sessions_token ON user_sessions(session_token);
CREATE INDEX IF NOT EXISTS idx_user_sessions_user_id ON user_sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_user_sessions_expires_at ON user_sessions(expires_at);

CREATE INDEX IF NOT EXISTS idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_action ON audit_logs(action);
CREATE INDEX IF NOT EXISTS idx_audit_logs_resource ON audit_logs(resource);
CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at ON audit_logs(created_at);

-- Insert default departments
INSERT INTO departments (name, description) VALUES 
('Administration', 'System administration and management'),
('Surveillance', 'Disease surveillance and monitoring'),
('Laboratory', 'Laboratory services and testing'),
('Clinical', 'Clinical services and patient care'),
('Data Management', 'Data entry and management'),
('Reports', 'Reporting and analytics')
ON CONFLICT (name) DO NOTHING;

-- Insert default roles
INSERT INTO roles (name, description) VALUES 
('super_admin', 'Super Administrator with full system access'),
('admin', 'Administrator with management access'),
('manager', 'Department manager with oversight access'),
('data_entry', 'Data entry personnel'),
('viewer', 'Read-only access to reports and data'),
('lab_technician', 'Laboratory technician with lab-specific access'),
('surveillance_officer', 'Surveillance officer with monitoring access')
ON CONFLICT (name) DO NOTHING;

-- Insert default permissions for VHF module
INSERT INTO permissions (name, description, resource, action) VALUES 
-- VHF Patients permissions
('Create VHF Patients', 'Create new VHF patient records', 'vhf_patients', 'create'),
('View VHF Patients', 'View VHF patient records', 'vhf_patients', 'read'),
('Update VHF Patients', 'Update VHF patient records', 'vhf_patients', 'update'),
('Delete VHF Patients', 'Delete VHF patient records', 'vhf_patients', 'delete'),
('Export VHF Patients', 'Export VHF patient data', 'vhf_patients', 'export'),

-- Users permissions
('Create Users', 'Create new user accounts', 'users', 'create'),
('View Users', 'View user accounts', 'users', 'read'),
('Update Users', 'Update user accounts', 'users', 'update'),
('Delete Users', 'Delete user accounts', 'users', 'delete'),

-- Reports permissions
('View Reports', 'View system reports', 'reports', 'read'),
('Export Reports', 'Export report data', 'reports', 'export'),

-- Outbreaks permissions
('Create Outbreaks', 'Create new outbreak records', 'outbreaks', 'create'),
('View Outbreaks', 'View outbreak records', 'outbreaks', 'read'),
('Update Outbreaks', 'Update outbreak records', 'outbreaks', 'update'),
('Delete Outbreaks', 'Delete outbreak records', 'outbreaks', 'delete'),

-- Employees permissions
('Create Employees', 'Create new employee records', 'employees', 'create'),
('View Employees', 'View employee records', 'employees', 'read'),
('Update Employees', 'Update employee records', 'employees', 'update'),
('Delete Employees', 'Delete employee records', 'employees', 'delete'),

-- Facilities permissions
('Create Facilities', 'Create new facility records', 'facilities', 'create'),
('View Facilities', 'View facility records', 'facilities', 'read'),
('Update Facilities', 'Update facility records', 'facilities', 'update'),
('Delete Facilities', 'Delete facility records', 'facilities', 'delete'),

-- Laboratory permissions
('Create Lab Results', 'Create laboratory test results', 'laboratory', 'create'),
('View Lab Results', 'View laboratory test results', 'laboratory', 'read'),
('Update Lab Results', 'Update laboratory test results', 'laboratory', 'update'),
('Delete Lab Results', 'Delete laboratory test results', 'laboratory', 'delete'),

-- Surveillance permissions
('Create Surveillance Data', 'Create surveillance data', 'surveillance', 'create'),
('View Surveillance Data', 'View surveillance data', 'surveillance', 'read'),
('Update Surveillance Data', 'Update surveillance data', 'surveillance', 'update'),
('Delete Surveillance Data', 'Delete surveillance data', 'surveillance', 'delete')
ON CONFLICT (resource, action) DO NOTHING;

-- Assign permissions to roles
-- Super Admin gets all permissions
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.name = 'super_admin'
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- Admin gets most permissions except super admin functions
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.name = 'admin' AND p.resource != 'users'
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- Manager gets read and update permissions for most resources
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.name = 'manager' AND p.action IN ('read', 'update', 'export')
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- Data Entry gets create and read permissions
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.name = 'data_entry' AND p.action IN ('create', 'read')
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- Viewer gets only read permissions
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.name = 'viewer' AND p.action = 'read'
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- Lab Technician gets lab-specific permissions
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.name = 'lab_technician' AND p.resource = 'laboratory'
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- Surveillance Officer gets surveillance permissions
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.name = 'surveillance_officer' AND p.resource = 'surveillance'
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- Create a default super admin user (password: admin123)
-- Note: In production, use proper password hashing
INSERT INTO enhanced_users (username, email, first_name, last_name, password_hash, password_salt, is_active)
VALUES ('admin', 'admin@system.local', 'System', 'Administrator', 
        'hashed_password_here', 'salt_here', true)
ON CONFLICT (username) DO NOTHING;

-- Assign super admin role to default admin user
INSERT INTO user_roles (user_id, role_id)
SELECT u.id, r.id FROM enhanced_users u, roles r
WHERE u.username = 'admin' AND r.name = 'super_admin'
ON CONFLICT (user_id, role_id) DO NOTHING;

-- Create function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create triggers to automatically update updated_at
CREATE TRIGGER update_enhanced_users_updated_at BEFORE UPDATE ON enhanced_users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_roles_updated_at BEFORE UPDATE ON roles
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_permissions_updated_at BEFORE UPDATE ON permissions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_departments_updated_at BEFORE UPDATE ON departments
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column(); 