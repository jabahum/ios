-- Migration: Merge enhanced_users into users table
-- This migration merges the enhanced_users table functionality into the existing users table
-- while preserving all existing data and relationships

-- First, add the missing RBAC columns to the existing users table
ALTER TABLE users ADD COLUMN IF NOT EXISTS email VARCHAR(255);
ALTER TABLE users ADD COLUMN IF NOT EXISTS first_name VARCHAR(100);
ALTER TABLE users ADD COLUMN IF NOT EXISTS last_name VARCHAR(100);
ALTER TABLE users ADD COLUMN IF NOT EXISTS is_active BOOLEAN DEFAULT true;
ALTER TABLE users ADD COLUMN IF NOT EXISTS is_locked BOOLEAN DEFAULT false;
ALTER TABLE users ADD COLUMN IF NOT EXISTS failed_login_attempts INTEGER DEFAULT 0;
ALTER TABLE users ADD COLUMN IF NOT EXISTS last_login_at TIMESTAMP WITH TIME ZONE;
ALTER TABLE users ADD COLUMN IF NOT EXISTS password_changed_at TIMESTAMP WITH TIME ZONE;
ALTER TABLE users ADD COLUMN IF NOT EXISTS password_expires_at TIMESTAMP WITH TIME ZONE;
ALTER TABLE users ADD COLUMN IF NOT EXISTS department_id INTEGER;
ALTER TABLE users ADD COLUMN IF NOT EXISTS created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE users ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE users ADD COLUMN IF NOT EXISTS created_by INTEGER;
ALTER TABLE users ADD COLUMN IF NOT EXISTS updated_by INTEGER;
ALTER TABLE users ADD COLUMN IF NOT EXISTS password_salt VARCHAR(255);

-- Add password_hash column to users table (will replace user_pass)
ALTER TABLE users ADD COLUMN IF NOT EXISTS password_hash VARCHAR(255);

-- Update user_roles table to reference users table instead of enhanced_users
-- First, drop the existing foreign key constraint if it exists
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.table_constraints 
        WHERE constraint_name = 'user_roles_user_id_fkey'
    ) THEN
        ALTER TABLE user_roles DROP CONSTRAINT user_roles_user_id_fkey;
    END IF;
END $$;

-- Update the foreign key to reference users table
ALTER TABLE user_roles ADD CONSTRAINT user_roles_user_id_fkey 
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE;

-- Update user_sessions table to reference users table
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.table_constraints 
        WHERE constraint_name = 'user_sessions_user_id_fkey'
    ) THEN
        ALTER TABLE user_sessions DROP CONSTRAINT user_sessions_user_id_fkey;
    END IF;
END $$;

ALTER TABLE user_sessions ADD CONSTRAINT user_sessions_user_id_fkey 
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE;

-- Update audit_logs table to reference users table
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.table_constraints 
        WHERE constraint_name = 'audit_logs_user_id_fkey'
    ) THEN
        ALTER TABLE audit_logs DROP CONSTRAINT audit_logs_user_id_fkey;
    END IF;
END $$;

-- Add foreign key constraint for department_id if it doesn't exist
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints 
        WHERE constraint_name = 'users_department_id_fkey'
    ) THEN
        ALTER TABLE users ADD CONSTRAINT users_department_id_fkey 
        FOREIGN KEY (department_id) REFERENCES departments(id);
    END IF;
END $$;

-- Migrate data from enhanced_users to users table
-- First, update existing users with data from enhanced_users where user_name matches username
UPDATE users 
SET 
    email = eu.email,
    first_name = eu.first_name,
    last_name = eu.last_name,
    is_active = eu.is_active,
    is_locked = eu.is_locked,
    failed_login_attempts = eu.failed_login_attempts,
    last_login_at = eu.last_login_at,
    password_changed_at = eu.password_changed_at,
    password_expires_at = eu.password_expires_at,
    department_id = eu.department_id,
    created_at = COALESCE(eu.created_at, users.created_at),
    updated_at = eu.updated_at,
    created_by = eu.created_by,
    updated_by = eu.updated_by,
    password_salt = eu.password_salt,
    password_hash = eu.password_hash
FROM enhanced_users eu
WHERE users.user_name = eu.username;

-- Insert any enhanced_users that don't exist in users table
INSERT INTO users (user_name, user_pass, user_employee, email, first_name, last_name, 
                   password_hash, password_salt, is_active, is_locked, failed_login_attempts,
                   last_login_at, password_changed_at, password_expires_at, department_id,
                   created_at, updated_at, created_by, updated_by)
SELECT 
    eu.username, 
    eu.password_hash, -- Use password_hash as user_pass for now
    eu.employee_id,
    eu.email,
    eu.first_name,
    eu.last_name,
    eu.password_hash,
    eu.password_salt,
    eu.is_active,
    eu.is_locked,
    eu.failed_login_attempts,
    eu.last_login_at,
    eu.password_changed_at,
    eu.password_expires_at,
    eu.department_id,
    eu.created_at,
    eu.updated_at,
    eu.created_by,
    eu.updated_by
FROM enhanced_users eu
WHERE NOT EXISTS (
    SELECT 1 FROM users u WHERE u.user_name = eu.username
);

-- Update user_roles to use the correct user_id from users table
UPDATE user_roles 
SET user_id = u.user_id
FROM users u, enhanced_users eu
WHERE user_roles.user_id = eu.id AND u.user_name = eu.username;

-- Update user_sessions to use the correct user_id from users table
UPDATE user_sessions 
SET user_id = u.user_id
FROM users u, enhanced_users eu
WHERE user_sessions.user_id = eu.id AND u.user_name = eu.username;

-- Update audit_logs to use the correct user_id from users table
UPDATE audit_logs 
SET user_id = u.user_id
FROM users u, enhanced_users eu
WHERE audit_logs.user_id = eu.id AND u.user_name = eu.username;

-- Create indexes for better performance on the users table
CREATE INDEX IF NOT EXISTS idx_users_username ON users(user_name);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_department ON users(department_id);
CREATE INDEX IF NOT EXISTS idx_users_active ON users(is_active);
CREATE INDEX IF NOT EXISTS idx_users_employee ON users(user_employee);

-- Create function to update updated_at timestamp if it doesn't exist
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create trigger to automatically update updated_at on users table
DROP TRIGGER IF EXISTS update_users_updated_at ON users;
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Now we can safely drop the enhanced_users table
-- But first, let's verify all data has been migrated
-- (This is a safety check - you may want to comment this out initially and run it manually)

-- DROP TABLE IF EXISTS enhanced_users CASCADE;

-- Add a comment to document the migration
COMMENT ON TABLE users IS 'Merged users table with RBAC functionality from enhanced_users'; 