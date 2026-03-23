-- Setup Resource Management Permissions
-- This script creates the resource_management permission resource and assigns it to appropriate roles

-- Insert the resource_management permissions (name is NOT NULL on permissions table)
INSERT INTO permissions (name, description, resource, action) VALUES
('View Resource Management', 'View resource management dashboard and lists', 'resource_management', 'read'),
('Create Resource Management', 'Create new resources, teams, deployments, requisitions, etc.', 'resource_management', 'create'),
('Update Resource Management', 'Update existing resources, approve/reject proposals, etc.', 'resource_management', 'update'),
('Delete Resource Management', 'Delete resources, teams, deployments, etc.', 'resource_management', 'delete')
ON CONFLICT (resource, action) DO NOTHING;

-- Assign resource_management permissions to super_admin role (full access)
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
CROSS JOIN permissions p
WHERE r.name = 'super_admin'
  AND p.resource = 'resource_management'
  AND NOT EXISTS (
    SELECT 1 FROM role_permissions rp
    WHERE rp.role_id = r.id AND rp.permission_id = p.id
);

-- Assign resource_management permissions to admin role (full access)
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
CROSS JOIN permissions p
WHERE r.name = 'admin'
  AND p.resource = 'resource_management'
  AND NOT EXISTS (
    SELECT 1 FROM role_permissions rp
    WHERE rp.role_id = r.id AND rp.permission_id = p.id
);

-- Assign read and create permissions to Outbreak Coordinator role
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
CROSS JOIN permissions p
WHERE r.name = 'outbreak_coordinator'
  AND p.resource = 'resource_management'
  AND p.action IN ('read', 'create')
  AND NOT EXISTS (
    SELECT 1 FROM role_permissions rp
    WHERE rp.role_id = r.id AND rp.permission_id = p.id
);

-- Assign read permission to Case Manager role
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
CROSS JOIN permissions p
WHERE r.name = 'case_manager'
  AND p.resource = 'resource_management'
  AND p.action = 'read'
  AND NOT EXISTS (
    SELECT 1 FROM role_permissions rp
    WHERE rp.role_id = r.id AND rp.permission_id = p.id
);

-- Display the results
SELECT 
    r.name AS role_name,
    p.resource,
    p.action,
    p.description
FROM roles r
JOIN role_permissions rp ON r.id = rp.role_id
JOIN permissions p ON p.id = rp.permission_id
WHERE p.resource = 'resource_management'
ORDER BY r.name, p.action;

