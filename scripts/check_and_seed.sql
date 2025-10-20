-- Check and seed missing data
-- This script only inserts data that doesn't already exist

-- ==============================================
-- CHECK AND INSERT MISSING INVENTORY CATEGORIES
-- ==============================================
INSERT INTO inventory_categories (name, description, created_by)
SELECT 'Personal Protective Equipment (PPE)', 'All types of PPE including masks, gloves, gowns, etc.', 1
WHERE NOT EXISTS (SELECT 1 FROM inventory_categories WHERE name = 'Personal Protective Equipment (PPE)');

INSERT INTO inventory_categories (name, description, created_by)
SELECT 'Medical Supplies', 'General medical supplies and equipment', 1
WHERE NOT EXISTS (SELECT 1 FROM inventory_categories WHERE name = 'Medical Supplies');

INSERT INTO inventory_categories (name, description, created_by)
SELECT 'Medications', 'Pharmaceutical products and drugs', 1
WHERE NOT EXISTS (SELECT 1 FROM inventory_categories WHERE name = 'Medications');

INSERT INTO inventory_categories (name, description, created_by)
SELECT 'Laboratory Supplies', 'Lab equipment, reagents, and testing materials', 1
WHERE NOT EXISTS (SELECT 1 FROM inventory_categories WHERE name = 'Laboratory Supplies');

INSERT INTO inventory_categories (name, description, created_by)
SELECT 'Infection Control', 'Disinfectants, sanitizers, and cleaning supplies', 1
WHERE NOT EXISTS (SELECT 1 FROM inventory_categories WHERE name = 'Infection Control');

INSERT INTO inventory_categories (name, description, created_by)
SELECT 'Emergency Response', 'Emergency and first aid supplies', 1
WHERE NOT EXISTS (SELECT 1 FROM inventory_categories WHERE name = 'Emergency Response');

INSERT INTO inventory_categories (name, description, created_by)
SELECT 'Nutrition', 'Food and nutritional supplements', 1
WHERE NOT EXISTS (SELECT 1 FROM inventory_categories WHERE name = 'Nutrition');

INSERT INTO inventory_categories (name, description, created_by)
SELECT 'Logistics', 'Transportation and storage equipment', 1
WHERE NOT EXISTS (SELECT 1 FROM inventory_categories WHERE name = 'Logistics');

INSERT INTO inventory_categories (name, description, created_by)
SELECT 'Diagnostic Equipment', 'Medical diagnostic devices and equipment', 1
WHERE NOT EXISTS (SELECT 1 FROM inventory_categories WHERE name = 'Diagnostic Equipment');

INSERT INTO inventory_categories (name, description, created_by)
SELECT 'Treatment Equipment', 'Medical treatment devices and equipment', 1
WHERE NOT EXISTS (SELECT 1 FROM inventory_categories WHERE name = 'Treatment Equipment');

-- ==============================================
-- CHECK AND INSERT MISSING ROLES
-- ==============================================
INSERT INTO roles (name, description)
SELECT 'inventory_manager', 'Manages inventory, requisitions, and stock levels'
WHERE NOT EXISTS (SELECT 1 FROM roles WHERE name = 'inventory_manager');

INSERT INTO roles (name, description)
SELECT 'inventory_clerk', 'Performs stock entry and basic inventory operations'
WHERE NOT EXISTS (SELECT 1 FROM roles WHERE name = 'inventory_clerk');

INSERT INTO roles (name, description)
SELECT 'outbreak_coordinator', 'Coordinates outbreak response and patient admissions'
WHERE NOT EXISTS (SELECT 1 FROM roles WHERE name = 'outbreak_coordinator');

INSERT INTO roles (name, description)
SELECT 'case_manager', 'Case Manager with case tracking and management access'
WHERE NOT EXISTS (SELECT 1 FROM roles WHERE name = 'case_manager');

INSERT INTO roles (name, description)
SELECT 'reports', 'Reports and analytics specialist with comprehensive reporting access'
WHERE NOT EXISTS (SELECT 1 FROM roles WHERE name = 'reports');

-- ==============================================
-- CHECK AND INSERT MISSING PERMISSIONS
-- ==============================================
INSERT INTO permissions (name, description, resource, action)
SELECT 'Create Inventory Category', 'Create new inventory categories', 'inventory_categories', 'create'
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE resource = 'inventory_categories' AND action = 'create');

INSERT INTO permissions (name, description, resource, action)
SELECT 'View Inventory Categories', 'View inventory categories', 'inventory_categories', 'read'
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE resource = 'inventory_categories' AND action = 'read');

INSERT INTO permissions (name, description, resource, action)
SELECT 'Update Inventory Categories', 'Update inventory categories', 'inventory_categories', 'update'
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE resource = 'inventory_categories' AND action = 'update');

INSERT INTO permissions (name, description, resource, action)
SELECT 'Delete Inventory Categories', 'Delete inventory categories', 'inventory_categories', 'delete'
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE resource = 'inventory_categories' AND action = 'delete');

INSERT INTO permissions (name, description, resource, action)
SELECT 'Create Inventory Item', 'Create new inventory items', 'inventory_items', 'create'
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE resource = 'inventory_items' AND action = 'create');

INSERT INTO permissions (name, description, resource, action)
SELECT 'View Inventory Items', 'View inventory items', 'inventory_items', 'read'
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE resource = 'inventory_items' AND action = 'read');

INSERT INTO permissions (name, description, resource, action)
SELECT 'Update Inventory Items', 'Update inventory items', 'inventory_items', 'update'
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE resource = 'inventory_items' AND action = 'update');

INSERT INTO permissions (name, description, resource, action)
SELECT 'Delete Inventory Items', 'Delete inventory items', 'inventory_items', 'delete'
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE resource = 'inventory_items' AND action = 'delete');

INSERT INTO permissions (name, description, resource, action)
SELECT 'Create Inventory Supplier', 'Create new inventory suppliers', 'inventory_suppliers', 'create'
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE resource = 'inventory_suppliers' AND action = 'create');

INSERT INTO permissions (name, description, resource, action)
SELECT 'View Inventory Suppliers', 'View inventory suppliers', 'inventory_suppliers', 'read'
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE resource = 'inventory_suppliers' AND action = 'read');

INSERT INTO permissions (name, description, resource, action)
SELECT 'Update Inventory Suppliers', 'Update inventory suppliers', 'inventory_suppliers', 'update'
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE resource = 'inventory_suppliers' AND action = 'update');

INSERT INTO permissions (name, description, resource, action)
SELECT 'Delete Inventory Suppliers', 'Delete inventory suppliers', 'inventory_suppliers', 'delete'
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE resource = 'inventory_suppliers' AND action = 'delete');

INSERT INTO permissions (name, description, resource, action)
SELECT 'Create Inventory Transaction', 'Create new inventory transactions (e.g., stock in/out)', 'inventory_transactions', 'create'
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE resource = 'inventory_transactions' AND action = 'create');

INSERT INTO permissions (name, description, resource, action)
SELECT 'View Inventory Transactions', 'View inventory transactions', 'inventory_transactions', 'read'
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE resource = 'inventory_transactions' AND action = 'read');

INSERT INTO permissions (name, description, resource, action)
SELECT 'Update Inventory Transactions', 'Update inventory transactions', 'inventory_transactions', 'update'
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE resource = 'inventory_transactions' AND action = 'update');

INSERT INTO permissions (name, description, resource, action)
SELECT 'Delete Inventory Transactions', 'Delete inventory transactions', 'inventory_transactions', 'delete'
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE resource = 'inventory_transactions' AND action = 'delete');

INSERT INTO permissions (name, description, resource, action)
SELECT 'View Inventory Reports', 'View various inventory reports', 'inventory_reports', 'read'
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE resource = 'inventory_reports' AND action = 'read');

INSERT INTO permissions (name, description, resource, action)
SELECT 'Generate Inventory Reports', 'Generate inventory reports', 'inventory_reports', 'generate'
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE resource = 'inventory_reports' AND action = 'generate');

-- ==============================================
-- SUMMARY
-- ==============================================
SELECT 'Seed data check completed successfully!' as message;
