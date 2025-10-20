-- Setup Resource Management Permissions
-- This script creates the resource_management permission resource and assigns it to appropriate roles

-- Insert the resource_management resource into permissions table if it doesn't exist
INSERT INTO permissions (resource, action, description)
SELECT 'resource_management', 'read', 'View resource management dashboard and lists'
WHERE NOT EXISTS (
    SELECT 1 FROM permissions WHERE resource = 'resource_management' AND action = 'read'
);

INSERT INTO permissions (resource, action, description)
SELECT 'resource_management', 'create', 'Create new resources, teams, deployments, requisitions, etc.'
WHERE NOT EXISTS (
    SELECT 1 FROM permissions WHERE resource = 'resource_management' AND action = 'create'
);

INSERT INTO permissions (resource, action, description)
SELECT 'resource_management', 'update', 'Update existing resources, approve/reject proposals, etc.'
WHERE NOT EXISTS (
    SELECT 1 FROM permissions WHERE resource = 'resource_management' AND action = 'update'
);

INSERT INTO permissions (resource, action, description)
SELECT 'resource_management', 'delete', 'Delete resources, teams, deployments, etc.'
WHERE NOT EXISTS (
    SELECT 1 FROM permissions WHERE resource = 'resource_management' AND action = 'delete'
);

-- Assign resource_management permissions to Super Admin role (full access)
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
CROSS JOIN permissions p
WHERE r.name = 'Super Admin'
  AND p.resource = 'resource_management'
  AND NOT EXISTS (
    SELECT 1 FROM role_permissions rp
    WHERE rp.role_id = r.id AND rp.permission_id = p.id
);

-- Assign resource_management permissions to Admin role (full access)
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
CROSS JOIN permissions p
WHERE r.name = 'Admin'
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
WHERE r.name = 'Outbreak Coordinator'
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
WHERE r.name = 'Case Manager'
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

