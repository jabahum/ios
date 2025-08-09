-- Migration to give user_id 1 all user roles
-- This will insert any missing role assignments for user_id 1

-- Insert all roles for user_id 1 (will ignore duplicates due to unique constraint)
INSERT INTO user_roles (user_id, role_id) 
SELECT 1, id FROM roles 
WHERE id NOT IN (
    SELECT role_id FROM user_roles WHERE user_id = 1
)
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