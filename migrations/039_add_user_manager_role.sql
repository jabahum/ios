-- Migration to add a dedicated user_manager role
-- This role will have full user management capabilities

-- Insert the new user_manager role
INSERT INTO roles (name, description) 
VALUES ('user_manager', 'Dedicated role for managing users, including create, read, update, and delete operations')
ON CONFLICT (name) DO NOTHING;

-- Get the role_id for the new user_manager role
DO $$
DECLARE
    user_manager_role_id INTEGER;
    permission_ids INTEGER[];
    perm_id INTEGER;
BEGIN
    -- Get the role_id for user_manager
    SELECT id INTO user_manager_role_id FROM roles WHERE name = 'user_manager';
    
    -- Get all permission IDs for 'users' resource
    SELECT ARRAY_AGG(id) INTO permission_ids FROM permissions WHERE resource = 'users';
    
    -- Assign all user permissions to the user_manager role
    FOREACH perm_id IN ARRAY permission_ids
    LOOP
        INSERT INTO role_permissions (role_id, permission_id) 
        VALUES (user_manager_role_id, perm_id)
        ON CONFLICT (role_id, permission_id) DO NOTHING;
    END LOOP;
END $$;

-- Assign the user_manager role to user_id 1
INSERT INTO user_roles (user_id, role_id) 
SELECT 1, id FROM roles WHERE name = 'user_manager'
ON CONFLICT (user_id, role_id) DO NOTHING;

-- Verify the result
SELECT 
    u.user_name,
    COUNT(ur.role_id) as total_roles,
    STRING_AGG(r.name, ', ' ORDER BY r.name) as role_names
FROM users u
LEFT JOIN user_roles ur ON u.user_id = ur.user_id
LEFT JOIN roles r ON ur.role_id = r.id
WHERE u.user_id = 1
GROUP BY u.user_id, u.user_name; 