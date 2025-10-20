-- Seed Basic Data for Inventory and RBAC
-- This script populates the database with essential data for dropdown lists

-- ==============================================
-- INVENTORY CATEGORIES
-- ==============================================
INSERT INTO inventory_categories (name, description, is_active, created_by) VALUES
('Personal Protective Equipment (PPE)', 'All types of PPE including masks, gloves, gowns, etc.', true, 1),
('Medical Supplies', 'General medical supplies and equipment', true, 1),
('Medications', 'Pharmaceutical products and drugs', true, 1),
('Laboratory Supplies', 'Lab equipment, reagents, and testing materials', true, 1),
('Infection Control', 'Disinfectants, sanitizers, and cleaning supplies', true, 1),
('Emergency Response', 'Emergency and first aid supplies', true, 1),
('Nutrition', 'Food and nutritional supplements', true, 1),
('Logistics', 'Transportation and storage equipment', true, 1),
('Diagnostic Equipment', 'Medical diagnostic devices and equipment', true, 1),
('Treatment Equipment', 'Medical treatment devices and equipment', true, 1)
ON CONFLICT (name) DO NOTHING;

-- ==============================================
-- INVENTORY SUPPLIERS
-- ==============================================
INSERT INTO inventory_suppliers (name, contact_person, email, phone, address, is_active, created_by) VALUES
('Ministry of Health Uganda', 'Dr. Jane Smith', 'procurement@health.go.ug', '+256-XXX-XXXX', 'Kampala, Uganda', true, 1),
('World Health Organization', 'Dr. John Doe', 'supplies@who.int', '+41-XXX-XXXX', 'Geneva, Switzerland', true, 1),
('UNICEF Uganda', 'Ms. Sarah Johnson', 'supplies@unicef.org', '+256-XXX-XXXX', 'Kampala, Uganda', true, 1),
('Medecins Sans Frontieres', 'Dr. Pierre Martin', 'logistics@msf.org', '+33-XXX-XXXX', 'Paris, France', true, 1),
('Red Cross Uganda', 'Mr. David Wilson', 'supplies@redcross.ug', '+256-XXX-XXXX', 'Kampala, Uganda', true, 1),
('Local Medical Suppliers', 'Mr. Ahmed Hassan', 'info@localmedical.ug', '+256-XXX-XXXX', 'Kampala, Uganda', true, 1),
('International Medical Corps', 'Dr. Maria Garcia', 'supplies@internationalmedicalcorps.org', '+1-XXX-XXXX', 'Los Angeles, USA', true, 1),
('Partners In Health', 'Dr. Paul Farmer', 'supplies@pih.org', '+1-XXX-XXXX', 'Boston, USA', true, 1)
ON CONFLICT (name) DO NOTHING;

-- ==============================================
-- TREATMENT SITES
-- ==============================================
INSERT INTO treatment_sites (name, location, site_type, contact_person, phone, email, is_active, created_by) VALUES
('Mulago National Referral Hospital', 'Kampala', 'Hospital', 'Dr. Sarah Mubiru', '+256-XXX-XXXX', 'mulago@health.go.ug', true, 1),
('Entebbe Regional Referral Hospital', 'Entebbe', 'Hospital', 'Dr. James Okello', '+256-XXX-XXXX', 'entebbe@health.go.ug', true, 1),
('Mbarara Regional Referral Hospital', 'Mbarara', 'Hospital', 'Dr. Grace Nakato', '+256-XXX-XXXX', 'mbarara@health.go.ug', true, 1),
('Gulu Regional Referral Hospital', 'Gulu', 'Hospital', 'Dr. Peter Ochieng', '+256-XXX-XXXX', 'gulu@health.go.ug', true, 1),
('Ebola Treatment Unit - Bwera', 'Bwera', 'ETU', 'Dr. Mary Akello', '+256-XXX-XXXX', 'bwera@health.go.ug', true, 1),
('Ebola Treatment Unit - Mubende', 'Mubende', 'ETU', 'Dr. John Kato', '+256-XXX-XXXX', 'mubende@health.go.ug', true, 1),
('Community Health Center - Kasese', 'Kasese', 'Health Center', 'Ms. Rose Namukasa', '+256-XXX-XXXX', 'kasese@health.go.ug', true, 1),
('Community Health Center - Bundibugyo', 'Bundibugyo', 'Health Center', 'Mr. Paul Musoke', '+256-XXX-XXXX', 'bundibugyo@health.go.ug', true, 1)
ON CONFLICT (name) DO NOTHING;

-- ==============================================
-- DEPARTMENTS
-- ==============================================
INSERT INTO departments (name, description, is_active) VALUES 
('Administration', 'System administration and management', true),
('Surveillance', 'Disease surveillance and monitoring', true),
('Laboratory', 'Laboratory services and testing', true),
('Clinical', 'Clinical services and patient care', true),
('Data Management', 'Data entry and management', true),
('Reports', 'Reporting and analytics', true),
('Emergency Response', 'Emergency response and outbreak management', true),
('Logistics', 'Supply chain and logistics management', true),
('Training', 'Training and capacity building', true),
('Quality Assurance', 'Quality assurance and compliance', true)
ON CONFLICT (name) DO NOTHING;

-- ==============================================
-- ROLES
-- ==============================================
INSERT INTO roles (name, description, is_active) VALUES 
('super_admin', 'Super Administrator with full system access', true),
('admin', 'Administrator with management access', true),
('manager', 'Department manager with oversight access', true),
('data_entry', 'Data entry personnel', true),
('viewer', 'Read-only access to reports and data', true),
('lab_technician', 'Laboratory technician with lab-specific access', true),
('surveillance_officer', 'Surveillance officer with monitoring access', true),
('clinical_officer', 'Clinical officer with patient care access', true),
('inventory_manager', 'Inventory and supply chain management', true),
('emergency_coordinator', 'Emergency response coordination', true),
('field_officer', 'Field operations and data collection', true),
('quality_officer', 'Quality assurance and compliance', true)
ON CONFLICT (name) DO NOTHING;

-- ==============================================
-- PERMISSIONS
-- ==============================================
INSERT INTO permissions (name, description, resource, action, is_active) VALUES 
-- VHF Patients permissions
('Create VHF Patients', 'Create new VHF patient records', 'vhf_patients', 'create', true),
('View VHF Patients', 'View VHF patient records', 'vhf_patients', 'read', true),
('Update VHF Patients', 'Update VHF patient records', 'vhf_patients', 'update', true),
('Delete VHF Patients', 'Delete VHF patient records', 'vhf_patients', 'delete', true),
('Export VHF Patients', 'Export VHF patient data', 'vhf_patients', 'export', true),

-- Measles permissions
('Create Measles Cases', 'Create new Measles case records', 'measles_cases', 'create', true),
('View Measles Cases', 'View Measles case records', 'measles_cases', 'read', true),
('Update Measles Cases', 'Update Measles case records', 'measles_cases', 'update', true),
('Delete Measles Cases', 'Delete Measles case records', 'measles_cases', 'delete', true),
('Export Measles Cases', 'Export Measles case data', 'measles_cases', 'export', true),

-- Polio permissions
('Create Polio Cases', 'Create new Polio case records', 'polio_cases', 'create', true),
('View Polio Cases', 'View Polio case records', 'polio_cases', 'read', true),
('Update Polio Cases', 'Update Polio case records', 'polio_cases', 'update', true),
('Delete Polio Cases', 'Delete Polio case records', 'polio_cases', 'delete', true),
('Export Polio Cases', 'Export Polio case data', 'polio_cases', 'export', true),

-- Mpox permissions
('Create Mpox Cases', 'Create new Mpox case records', 'mpox_cases', 'create', true),
('View Mpox Cases', 'View Mpox case records', 'mpox_cases', 'read', true),
('Update Mpox Cases', 'Update Mpox case records', 'mpox_cases', 'update', true),
('Delete Mpox Cases', 'Delete Mpox case records', 'mpox_cases', 'delete', true),
('Export Mpox Cases', 'Export Mpox case data', 'mpox_cases', 'export', true),

-- Users permissions
('Create Users', 'Create new user accounts', 'users', 'create', true),
('View Users', 'View user accounts', 'users', 'read', true),
('Update Users', 'Update user accounts', 'users', 'update', true),
('Delete Users', 'Delete user accounts', 'users', 'delete', true),
('Manage User Roles', 'Assign and manage user roles', 'users', 'manage_roles', true),

-- Roles permissions
('Create Roles', 'Create new roles', 'roles', 'create', true),
('View Roles', 'View roles', 'roles', 'read', true),
('Update Roles', 'Update roles', 'roles', 'update', true),
('Delete Roles', 'Delete roles', 'roles', 'delete', true),

-- Permissions management
('Create Permissions', 'Create new permissions', 'permissions', 'create', true),
('View Permissions', 'View permissions', 'permissions', 'read', true),
('Update Permissions', 'Update permissions', 'permissions', 'update', true),
('Delete Permissions', 'Delete permissions', 'permissions', 'delete', true),

-- Inventory permissions
('Create Inventory Items', 'Create new inventory items', 'inventory', 'create', true),
('View Inventory Items', 'View inventory items', 'inventory', 'read', true),
('Update Inventory Items', 'Update inventory items', 'inventory', 'update', true),
('Delete Inventory Items', 'Delete inventory items', 'inventory', 'delete', true),
('Manage Inventory Categories', 'Manage inventory categories', 'inventory', 'manage_categories', true),
('Manage Inventory Suppliers', 'Manage inventory suppliers', 'inventory', 'manage_suppliers', true),
('View Inventory Reports', 'View inventory reports', 'inventory', 'view_reports', true),

-- Outbreaks permissions
('Create Outbreaks', 'Create new outbreaks', 'outbreaks', 'create', true),
('View Outbreaks', 'View outbreaks', 'outbreaks', 'read', true),
('Update Outbreaks', 'Update outbreaks', 'outbreaks', 'update', true),
('Delete Outbreaks', 'Delete outbreaks', 'outbreaks', 'delete', true),
('Close Outbreaks', 'Close outbreaks', 'outbreaks', 'close', true),

-- Reports permissions
('View Reports', 'View system reports', 'reports', 'read', true),
('Export Reports', 'Export report data', 'reports', 'export', true),
('Create Reports', 'Create custom reports', 'reports', 'create', true),

-- System permissions
('View System Settings', 'View system settings', 'system', 'read', true),
('Update System Settings', 'Update system settings', 'system', 'update', true),
('View Audit Logs', 'View audit logs', 'audit', 'read', true),
('Manage Database', 'Database management operations', 'database', 'manage', true)
ON CONFLICT (resource, action) DO NOTHING;

-- ==============================================
-- ROLE-PERMISSION ASSIGNMENTS
-- ==============================================
-- Super Admin gets all permissions
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.name = 'super_admin'
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- Admin gets most permissions except super admin functions
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.name = 'admin' 
AND p.resource NOT IN ('database', 'audit')
AND p.action NOT IN ('manage')
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- Manager gets read/write access to cases and reports
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.name = 'manager' 
AND (
    (p.resource IN ('vhf_patients', 'measles_cases', 'polio_cases', 'mpox_cases') AND p.action IN ('read', 'update', 'export'))
    OR (p.resource = 'reports' AND p.action IN ('read', 'export'))
    OR (p.resource = 'inventory' AND p.action IN ('read', 'view_reports'))
)
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- Data Entry gets create/read/update access to cases
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.name = 'data_entry' 
AND p.resource IN ('vhf_patients', 'measles_cases', 'polio_cases', 'mpox_cases') 
AND p.action IN ('create', 'read', 'update')
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- Lab Technician gets lab-specific permissions
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.name = 'lab_technician' 
AND (
    (p.resource IN ('vhf_patients', 'measles_cases', 'polio_cases', 'mpox_cases') AND p.action IN ('read', 'update'))
    OR (p.resource = 'inventory' AND p.action IN ('read', 'view_reports'))
)
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- Surveillance Officer gets surveillance permissions
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.name = 'surveillance_officer' 
AND (
    (p.resource IN ('vhf_patients', 'measles_cases', 'polio_cases', 'mpox_cases', 'outbreaks') AND p.action IN ('read', 'create', 'update', 'export'))
    OR (p.resource = 'reports' AND p.action IN ('read', 'export'))
)
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- Clinical Officer gets clinical permissions
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.name = 'clinical_officer' 
AND p.resource IN ('vhf_patients', 'measles_cases', 'polio_cases', 'mpox_cases') 
AND p.action IN ('create', 'read', 'update')
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- Inventory Manager gets inventory permissions
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.name = 'inventory_manager' 
AND p.resource = 'inventory'
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- Emergency Coordinator gets emergency response permissions
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.name = 'emergency_coordinator' 
AND (
    (p.resource IN ('vhf_patients', 'measles_cases', 'polio_cases', 'mpox_cases', 'outbreaks') AND p.action IN ('create', 'read', 'update', 'export'))
    OR (p.resource = 'reports' AND p.action IN ('read', 'export'))
    OR (p.resource = 'inventory' AND p.action IN ('read', 'view_reports'))
)
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- Field Officer gets field data collection permissions
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.name = 'field_officer' 
AND p.resource IN ('vhf_patients', 'measles_cases', 'polio_cases', 'mpox_cases') 
AND p.action IN ('create', 'read', 'update')
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- Quality Officer gets quality assurance permissions
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.name = 'quality_officer' 
AND (
    (p.resource IN ('vhf_patients', 'measles_cases', 'polio_cases', 'mpox_cases') AND p.action IN ('read', 'update'))
    OR (p.resource = 'reports' AND p.action IN ('read', 'export'))
    OR (p.resource = 'audit' AND p.action = 'read')
)
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- Viewer gets read-only access
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.name = 'viewer' 
AND p.action IN ('read', 'export')
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- ==============================================
-- INVENTORY SETTINGS
-- ==============================================
INSERT INTO inventory_settings (setting_key, setting_value, setting_description, is_active) VALUES
('low_stock_threshold', '20', 'Percentage threshold for low stock alerts', true),
('expiry_alert_days', '30', 'Days before expiry to send alerts', true),
('auto_reorder_enabled', 'false', 'Enable automatic reordering', true),
('require_approval', 'true', 'Require approval for requisitions', true),
('max_requisition_amount', '10000', 'Maximum amount for automatic approval', true),
('stock_take_frequency', '7', 'Days between stock takes', true),
('default_unit_of_measure', 'pieces', 'Default unit of measure for new items', true),
('enable_barcode_scanning', 'true', 'Enable barcode scanning functionality', true),
('enable_expiry_tracking', 'true', 'Enable expiry date tracking', true),
('enable_cold_storage_tracking', 'true', 'Enable cold storage tracking', true)
ON CONFLICT (setting_key) DO NOTHING;

-- ==============================================
-- SAMPLE INVENTORY ITEMS
-- ==============================================
INSERT INTO inventory_items (name, description, category_id, item_code, unit_of_measure, minimum_stock_level, unit_cost, is_active, is_critical, created_by) VALUES
('N95 Respirator Masks', 'N95 filtering facepiece respirators', (SELECT id FROM inventory_categories WHERE name = 'Personal Protective Equipment (PPE)' LIMIT 1), 'PPE-001', 'pieces', 100, 2.50, true, true, 1),
('Disposable Gloves', 'Nitrile examination gloves', (SELECT id FROM inventory_categories WHERE name = 'Personal Protective Equipment (PPE)' LIMIT 1), 'PPE-002', 'boxes', 50, 15.00, true, true, 1),
('Isolation Gowns', 'Disposable isolation gowns', (SELECT id FROM inventory_categories WHERE name = 'Personal Protective Equipment (PPE)' LIMIT 1), 'PPE-003', 'pieces', 200, 5.00, true, true, 1),
('Face Shields', 'Disposable face shields', (SELECT id FROM inventory_categories WHERE name = 'Personal Protective Equipment (PPE)' LIMIT 1), 'PPE-004', 'pieces', 100, 3.00, true, true, 1),
('Ebola RDT Test Kits', 'Rapid diagnostic test kits for Ebola', (SELECT id FROM inventory_categories WHERE name = 'Laboratory Supplies' LIMIT 1), 'LAB-001', 'kits', 20, 25.00, true, true, 1),
('PCR Reagents', 'PCR testing reagents for Ebola', (SELECT id FROM inventory_categories WHERE name = 'Laboratory Supplies' LIMIT 1), 'LAB-002', 'kits', 10, 150.00, true, true, 1),
('Blood Collection Tubes', 'Sterile blood collection tubes', (SELECT id FROM inventory_categories WHERE name = 'Laboratory Supplies' LIMIT 1), 'LAB-003', 'tubes', 500, 0.50, true, false, 1),
('Disinfectant Solution', 'Hospital-grade disinfectant', (SELECT id FROM inventory_categories WHERE name = 'Infection Control' LIMIT 1), 'IC-001', 'liters', 50, 8.00, true, true, 1),
('Hand Sanitizer', 'Alcohol-based hand sanitizer', (SELECT id FROM inventory_categories WHERE name = 'Infection Control' LIMIT 1), 'IC-002', 'bottles', 100, 3.00, true, true, 1),
('IV Fluids', 'Normal saline solution', (SELECT id FROM inventory_categories WHERE name = 'Medical Supplies' LIMIT 1), 'MED-001', 'bags', 200, 2.00, true, true, 1)
ON CONFLICT (item_code) DO NOTHING;

-- ==============================================
-- SAMPLE OUTBREAKS (if outbreaks table exists)
-- ==============================================
-- Note: This assumes you have an outbreaks table. Adjust as needed.
-- INSERT INTO outbreaks (name, description, status, start_date, created_by) VALUES
-- ('Ebola Outbreak 2024', 'Ebola virus disease outbreak in Western Uganda', 'active', CURRENT_DATE, 1),
-- ('Measles Outbreak 2024', 'Measles outbreak in Central Uganda', 'active', CURRENT_DATE, 1),
-- ('Polio Surveillance 2024', 'Polio surveillance activities', 'active', CURRENT_DATE, 1),
-- ('Mpox Outbreak 2024', 'Monkeypox outbreak in Eastern Uganda', 'active', CURRENT_DATE, 1)
-- ON CONFLICT (name) DO NOTHING;

-- ==============================================
-- VERIFICATION QUERIES
-- ==============================================
-- Run these to verify the data was inserted correctly:

-- SELECT 'Inventory Categories' as table_name, COUNT(*) as count FROM inventory_categories
-- UNION ALL
-- SELECT 'Inventory Suppliers', COUNT(*) FROM inventory_suppliers
-- UNION ALL
-- SELECT 'Treatment Sites', COUNT(*) FROM treatment_sites
-- UNION ALL
-- SELECT 'Departments', COUNT(*) FROM departments
-- UNION ALL
-- SELECT 'Roles', COUNT(*) FROM roles
-- UNION ALL
-- SELECT 'Permissions', COUNT(*) FROM permissions
-- UNION ALL
-- SELECT 'Role-Permission Assignments', COUNT(*) FROM role_permissions
-- UNION ALL
-- SELECT 'Inventory Items', COUNT(*) FROM inventory_items
-- UNION ALL
-- SELECT 'Inventory Settings', COUNT(*) FROM inventory_settings;

COMMIT;
