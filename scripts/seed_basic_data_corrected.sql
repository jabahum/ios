-- Seed Basic Data for Inventory and RBAC
-- This script populates the database with essential data for dropdown lists
-- Updated to work with the actual database schema

-- ==============================================
-- INVENTORY CATEGORIES (if table exists)
-- ==============================================
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'inventory_categories') THEN
        INSERT INTO inventory_categories (name, description, created_by) VALUES
        ('Personal Protective Equipment (PPE)', 'All types of PPE including masks, gloves, gowns, etc.', 1),
        ('Medical Supplies', 'General medical supplies and equipment', 1),
        ('Medications', 'Pharmaceutical products and drugs', 1),
        ('Laboratory Supplies', 'Lab equipment, reagents, and testing materials', 1),
        ('Infection Control', 'Disinfectants, sanitizers, and cleaning supplies', 1),
        ('Emergency Response', 'Emergency and first aid supplies', 1),
        ('Nutrition', 'Food and nutritional supplements', 1),
        ('Logistics', 'Transportation and storage equipment', 1),
        ('Diagnostic Equipment', 'Medical diagnostic devices and equipment', 1),
        ('Treatment Equipment', 'Medical treatment devices and equipment', 1)
        ON CONFLICT (name) DO UPDATE SET description = EXCLUDED.description;
        
        RAISE NOTICE 'Inventory categories inserted successfully';
    ELSE
        RAISE NOTICE 'inventory_categories table does not exist, skipping...';
    END IF;
END $$;

-- ==============================================
-- INVENTORY SUPPLIERS (if table exists)
-- ==============================================
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'inventory_suppliers') THEN
        INSERT INTO inventory_suppliers (name, contact_person, email, phone, address, created_by) VALUES
        ('Ministry of Health Uganda', 'Dr. Jane Smith', 'procurement@health.go.ug', '+256-XXX-XXXX', 'Kampala, Uganda', 1),
        ('World Health Organization', 'Dr. John Doe', 'supplies@who.int', '+41-XXX-XXXX', 'Geneva, Switzerland', 1),
        ('UNICEF Uganda', 'Ms. Sarah Johnson', 'supplies@unicef.org', '+256-XXX-XXXX', 'Kampala, Uganda', 1),
        ('Medecins Sans Frontieres', 'Dr. Pierre Martin', 'logistics@msf.org', '+33-XXX-XXXX', 'Paris, France', 1),
        ('Red Cross Uganda', 'Mr. David Wilson', 'supplies@redcross.ug', '+256-XXX-XXXX', 'Kampala, Uganda', 1),
        ('Local Medical Suppliers', 'Mr. Ahmed Hassan', 'info@localmedical.ug', '+256-XXX-XXXX', 'Kampala, Uganda', 1),
        ('International Medical Corps', 'Dr. Maria Garcia', 'supplies@internationalmedicalcorps.org', '+1-XXX-XXXX', 'Los Angeles, USA', 1),
        ('Partners In Health', 'Dr. Paul Farmer', 'supplies@pih.org', '+1-XXX-XXXX', 'Boston, USA', 1)
        ON CONFLICT (name) DO UPDATE SET description = EXCLUDED.description;
        
        RAISE NOTICE 'Inventory suppliers inserted successfully';
    ELSE
        RAISE NOTICE 'inventory_suppliers table does not exist, skipping...';
    END IF;
END $$;

-- ==============================================
-- TREATMENT SITES (if table exists)
-- ==============================================
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'treatment_sites') THEN
        INSERT INTO treatment_sites (name, site_type, contact_person, phone, email, created_by) VALUES
        ('Mulago National Referral Hospital', 'Hospital', 'Dr. Sarah Mubiru', '+256-XXX-XXXX', 'mulago@health.go.ug', 1),
        ('Entebbe Regional Referral Hospital', 'Hospital', 'Dr. James Okello', '+256-XXX-XXXX', 'entebbe@health.go.ug', 1),
        ('Mbarara Regional Referral Hospital', 'Hospital', 'Dr. Grace Nakato', '+256-XXX-XXXX', 'mbarara@health.go.ug', 1),
        ('Gulu Regional Referral Hospital', 'Hospital', 'Dr. Peter Ochieng', '+256-XXX-XXXX', 'gulu@health.go.ug', 1),
        ('Ebola Treatment Unit - Bwera', 'ETU', 'Dr. Mary Akello', '+256-XXX-XXXX', 'bwera@health.go.ug', 1),
        ('Ebola Treatment Unit - Mubende', 'ETU', 'Dr. John Kato', '+256-XXX-XXXX', 'mubende@health.go.ug', 1),
        ('Community Health Center - Kasese', 'Health Center', 'Ms. Rose Namukasa', '+256-XXX-XXXX', 'kasese@health.go.ug', 1),
        ('Community Health Center - Bundibugyo', 'Health Center', 'Mr. Paul Musoke', '+256-XXX-XXXX', 'bundibugyo@health.go.ug', 1)
        ON CONFLICT (name) DO UPDATE SET description = EXCLUDED.description;
        
        RAISE NOTICE 'Treatment sites inserted successfully';
    ELSE
        RAISE NOTICE 'treatment_sites table does not exist, skipping...';
    END IF;
END $$;

-- ==============================================
-- DEPARTMENTS (if table exists)
-- ==============================================
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'departments') THEN
        INSERT INTO departments (name, description) VALUES
        ('Administration', 'System administration and management'),
        ('Surveillance', 'Disease surveillance and monitoring'),
        ('Laboratory', 'Laboratory services and testing'),
        ('Clinical', 'Clinical services and patient care'),
        ('Data Management', 'Data entry and management'),
        ('Reports', 'Reporting and analytics'),
        ('Emergency Response', 'Emergency response and coordination'),
        ('Logistics', 'Supply chain and logistics management'),
        ('Training', 'Training and capacity building'),
        ('Quality Assurance', 'Quality assurance and compliance')
        ON CONFLICT (name) DO UPDATE SET description = EXCLUDED.description;
        
        RAISE NOTICE 'Departments inserted successfully';
    ELSE
        RAISE NOTICE 'departments table does not exist, skipping...';
    END IF;
END $$;

-- ==============================================
-- ROLES (if table exists)
-- ==============================================
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'roles') THEN
        INSERT INTO roles (name, description) VALUES
        ('super_admin', 'Super Administrator with full system access'),
        ('admin', 'Administrator with management access'),
        ('manager', 'Department manager with oversight access'),
        ('data_entry', 'Data entry personnel'),
        ('viewer', 'Read-only access to reports and data'),
        ('lab_technician', 'Laboratory technician with lab-specific access'),
        ('surveillance_officer', 'Surveillance officer with monitoring access'),
        ('inventory_manager', 'Manages inventory, requisitions, and stock levels'),
        ('inventory_clerk', 'Performs stock entry and basic inventory operations'),
        ('outbreak_coordinator', 'Coordinates outbreak response and patient admissions'),
        ('case_manager', 'Case Manager with case tracking and management access'),
        ('reports', 'Reports and analytics specialist with comprehensive reporting access')
        ON CONFLICT (name) DO UPDATE SET description = EXCLUDED.description;
        
        RAISE NOTICE 'Roles inserted successfully';
    ELSE
        RAISE NOTICE 'roles table does not exist, skipping...';
    END IF;
END $$;

-- ==============================================
-- PERMISSIONS (if table exists)
-- ==============================================
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'permissions') THEN
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
        
        -- Inventory permissions
        ('Create Inventory Category', 'Create new inventory categories', 'inventory_categories', 'create'),
        ('View Inventory Categories', 'View inventory categories', 'inventory_categories', 'read'),
        ('Update Inventory Categories', 'Update inventory categories', 'inventory_categories', 'update'),
        ('Delete Inventory Categories', 'Delete inventory categories', 'inventory_categories', 'delete'),
        
        ('Create Inventory Item', 'Create new inventory items', 'inventory_items', 'create'),
        ('View Inventory Items', 'View inventory items', 'inventory_items', 'read'),
        ('Update Inventory Items', 'Update inventory items', 'inventory_items', 'update'),
        ('Delete Inventory Items', 'Delete inventory items', 'inventory_items', 'delete'),
        
        ('Create Inventory Supplier', 'Create new inventory suppliers', 'inventory_suppliers', 'create'),
        ('View Inventory Suppliers', 'View inventory suppliers', 'inventory_suppliers', 'read'),
        ('Update Inventory Suppliers', 'Update inventory suppliers', 'inventory_suppliers', 'update'),
        ('Delete Inventory Suppliers', 'Delete inventory suppliers', 'inventory_suppliers', 'delete'),
        
        ('Create Inventory Transaction', 'Create new inventory transactions (e.g., stock in/out)', 'inventory_transactions', 'create'),
        ('View Inventory Transactions', 'View inventory transactions', 'inventory_transactions', 'read'),
        ('Update Inventory Transactions', 'Update inventory transactions', 'inventory_transactions', 'update'),
        ('Delete Inventory Transactions', 'Delete inventory transactions', 'inventory_transactions', 'delete'),
        
        ('View Inventory Reports', 'View various inventory reports', 'inventory_reports', 'read'),
        ('Generate Inventory Reports', 'Generate inventory reports', 'inventory_reports', 'generate')
        ON CONFLICT (resource, action) DO UPDATE SET name = EXCLUDED.name, description = EXCLUDED.description;
        
        RAISE NOTICE 'Permissions inserted successfully';
    ELSE
        RAISE NOTICE 'permissions table does not exist, skipping...';
    END IF;
END $$;

-- ==============================================
-- ROLE-PERMISSION ASSIGNMENTS (if tables exist)
-- ==============================================
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'role_permissions') THEN
        -- Assign Inventory Manager role permissions
        INSERT INTO role_permissions (role_id, permission_id)
        SELECT r.id, p.id FROM roles r, permissions p
        WHERE r.name = 'inventory_manager' AND p.resource IN ('inventory_categories', 'inventory_items', 'inventory_suppliers', 'inventory_transactions', 'inventory_reports')
        ON CONFLICT (role_id, permission_id) DO NOTHING;
        
        -- Assign Inventory Clerk role permissions (read and create transactions)
        INSERT INTO role_permissions (role_id, permission_id)
        SELECT r.id, p.id FROM roles r, permissions p
        WHERE r.name = 'inventory_clerk' AND p.resource IN ('inventory_items', 'inventory_transactions') AND p.action IN ('read', 'create')
        ON CONFLICT (role_id, permission_id) DO NOTHING;
        
        -- Assign Outbreak Coordinator role permissions (example: view patients, create admissions)
        INSERT INTO role_permissions (role_id, permission_id)
        SELECT r.id, p.id FROM roles r, permissions p
        WHERE r.name = 'outbreak_coordinator' AND p.resource IN ('vhf_patients', 'cases') AND p.action IN ('read', 'create')
        ON CONFLICT (role_id, permission_id) DO NOTHING;
        
        RAISE NOTICE 'Role-permission assignments completed successfully';
    ELSE
        RAISE NOTICE 'role_permissions table does not exist, skipping...';
    END IF;
END $$;

-- ==============================================
-- INVENTORY ITEMS (if table exists)
-- ==============================================
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'inventory_items') THEN
        INSERT INTO inventory_items (name, description, category_id, item_code, unit_of_measure, minimum_stock_level, reorder_point, unit_cost, created_by) VALUES
        ('N95 Mask', 'Respiratory protective mask', (SELECT id FROM inventory_categories WHERE name = 'Personal Protective Equipment (PPE)' LIMIT 1), 'PPE001', 'pieces', 50, 20, 2.50, 1),
        ('Sterile Gloves (Box of 100)', 'Disposable sterile gloves', (SELECT id FROM inventory_categories WHERE name = 'Personal Protective Equipment (PPE)' LIMIT 1), 'PPE002', 'boxes', 30, 10, 15.00, 1),
        ('Paracetamol 500mg (Box of 1000)', 'Pain reliever and fever reducer', (SELECT id FROM inventory_categories WHERE name = 'Medications' LIMIT 1), 'MED001', 'boxes', 100, 50, 25.00, 1),
        ('Syringe 5ml', 'Disposable syringe', (SELECT id FROM inventory_categories WHERE name = 'Medical Supplies' LIMIT 1), 'MEDS001', 'pieces', 200, 100, 0.50, 1),
        ('COVID-19 Antigen Rapid Test Kit', 'Rapid diagnostic test for COVID-19', (SELECT id FROM inventory_categories WHERE name = 'Laboratory Supplies' LIMIT 1), 'LAB001', 'kits', 20, 10, 10.00, 1),
        ('Hand Sanitizer 500ml', 'Alcohol-based hand sanitizer', (SELECT id FROM inventory_categories WHERE name = 'Infection Control' LIMIT 1), 'IC001', 'bottles', 50, 20, 5.00, 1),
        ('Emergency Blanket', 'Thermal emergency blanket', (SELECT id FROM inventory_categories WHERE name = 'Emergency Response' LIMIT 1), 'ER001', 'pieces', 25, 10, 3.00, 1),
        ('Protein Powder 1kg', 'Nutritional supplement', (SELECT id FROM inventory_categories WHERE name = 'Nutrition' LIMIT 1), 'NUT001', 'kg', 10, 5, 20.00, 1),
        ('Transport Cooler Box', 'Insulated transport container', (SELECT id FROM inventory_categories WHERE name = 'Logistics' LIMIT 1), 'LOG001', 'pieces', 5, 2, 50.00, 1),
        ('Digital Thermometer', 'Digital medical thermometer', (SELECT id FROM inventory_categories WHERE name = 'Diagnostic Equipment' LIMIT 1), 'DIAG001', 'pieces', 10, 5, 25.00, 1)
        ON CONFLICT (item_code) DO UPDATE SET name = EXCLUDED.name, description = EXCLUDED.description;
        
        RAISE NOTICE 'Inventory items inserted successfully';
    ELSE
        RAISE NOTICE 'inventory_items table does not exist, skipping...';
    END IF;
END $$;

-- ==============================================
-- SUMMARY
-- ==============================================
DO $$
DECLARE
    category_count INTEGER := 0;
    supplier_count INTEGER := 0;
    site_count INTEGER := 0;
    dept_count INTEGER := 0;
    role_count INTEGER := 0;
    perm_count INTEGER := 0;
    item_count INTEGER := 0;
BEGIN
    -- Count inserted records
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'inventory_categories') THEN
        SELECT COUNT(*) INTO category_count FROM inventory_categories;
    END IF;
    
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'inventory_suppliers') THEN
        SELECT COUNT(*) INTO supplier_count FROM inventory_suppliers;
    END IF;
    
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'treatment_sites') THEN
        SELECT COUNT(*) INTO site_count FROM treatment_sites;
    END IF;
    
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'departments') THEN
        SELECT COUNT(*) INTO dept_count FROM departments;
    END IF;
    
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'roles') THEN
        SELECT COUNT(*) INTO role_count FROM roles;
    END IF;
    
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'permissions') THEN
        SELECT COUNT(*) INTO perm_count FROM permissions;
    END IF;
    
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'inventory_items') THEN
        SELECT COUNT(*) INTO item_count FROM inventory_items;
    END IF;
    
    RAISE NOTICE '==============================================';
    RAISE NOTICE 'SEED DATA SUMMARY';
    RAISE NOTICE '==============================================';
    RAISE NOTICE 'Inventory Categories: % records', category_count;
    RAISE NOTICE 'Inventory Suppliers: % records', supplier_count;
    RAISE NOTICE 'Treatment Sites: % records', site_count;
    RAISE NOTICE 'Departments: % records', dept_count;
    RAISE NOTICE 'Roles: % records', role_count;
    RAISE NOTICE 'Permissions: % records', perm_count;
    RAISE NOTICE 'Inventory Items: % records', item_count;
    RAISE NOTICE '==============================================';
    RAISE NOTICE 'Seed data setup complete!';
    RAISE NOTICE 'You can now:';
    RAISE NOTICE '- Create users and assign roles';
    RAISE NOTICE '- Create inventory items and manage categories';
    RAISE NOTICE '- Set up permissions for different user types';
    RAISE NOTICE '- All dropdown lists should now be populated';
    RAISE NOTICE '==============================================';
END $$;
