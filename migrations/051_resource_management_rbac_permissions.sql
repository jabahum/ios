-- Resource management UI routes require permissions resource = 'resource_management'
-- (see internal/routes/routes.go PermissionRequired).
-- This was previously only in scripts/ with mismatched role display names vs DB role keys.

INSERT INTO permissions (name, description, resource, action) VALUES
('View Resource Management', 'View resource management dashboard and lists', 'resource_management', 'read'),
('Create Resource Management', 'Create new resources, teams, deployments, requisitions, etc.', 'resource_management', 'create'),
('Update Resource Management', 'Update existing resources, approve/reject proposals, etc.', 'resource_management', 'update'),
('Delete Resource Management', 'Delete resources, teams, deployments, etc.', 'resource_management', 'delete')
ON CONFLICT (resource, action) DO NOTHING;

-- super_admin: full access
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
CROSS JOIN permissions p
WHERE r.name = 'super_admin'
  AND p.resource = 'resource_management'
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- admin: full access (consistent with other modules)
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
CROSS JOIN permissions p
WHERE r.name = 'admin'
  AND p.resource = 'resource_management'
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- outbreak_coordinator: read + create (when role exists, e.g. from seed scripts)
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
CROSS JOIN permissions p
WHERE r.name = 'outbreak_coordinator'
  AND p.resource = 'resource_management'
  AND p.action IN ('read', 'create')
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- case_manager: read only
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
CROSS JOIN permissions p
WHERE r.name = 'case_manager'
  AND p.resource = 'resource_management'
  AND p.action = 'read'
ON CONFLICT (role_id, permission_id) DO NOTHING;
