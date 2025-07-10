-- Migration: Add Case Manager Role
-- This migration adds a case_manager role for users who should go directly to /cases

-- Create case_manager role
INSERT INTO roles (name, description) VALUES 
('case_manager', 'Case manager - can access case management with default outbreak assignment')
ON CONFLICT (name) DO NOTHING;

-- Create case management permissions
INSERT INTO permissions (name, description, resource, action) VALUES 
('Access Case Management', 'Access case management interface', 'cases', 'read'),
('Create Cases', 'Create new case records', 'cases', 'create'),
('Update Cases', 'Update case records', 'cases', 'update'),
('View Case Encounters', 'View case encounter data', 'case_encounters', 'read'),
('Create Case Encounters', 'Create new case encounters', 'case_encounters', 'create'),
('Update Case Encounters', 'Update case encounters', 'case_encounters', 'update')
ON CONFLICT (resource, action) DO NOTHING;

-- Assign permissions to case_manager role
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.name = 'case_manager' AND p.resource IN ('cases', 'case_encounters')
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- Add a user_preferences table to store default outbreak assignments
CREATE TABLE IF NOT EXISTS user_preferences (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    preference_key VARCHAR(100) NOT NULL,
    preference_value TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, preference_key)
);

-- Create index for user preferences
CREATE INDEX IF NOT EXISTS idx_user_preferences_user_id ON user_preferences(user_id);
CREATE INDEX IF NOT EXISTS idx_user_preferences_key ON user_preferences(preference_key);

-- Add comment
COMMENT ON TABLE user_preferences IS 'User preferences for default settings like outbreak assignments'; 