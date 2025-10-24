-- =====================================================
-- Resource Management System for Outbreak Response
-- =====================================================
-- This migration creates the complete Resource Management system
-- for tracking RRT deployments, resources, and outbreak logistics

-- RRT Teams
CREATE TABLE IF NOT EXISTS rrt_teams (
    id SERIAL PRIMARY KEY,
    team_name VARCHAR(255) NOT NULL,
    team_code VARCHAR(50) UNIQUE NOT NULL,
    team_type VARCHAR(100) NOT NULL, -- 'investigation', 'logistics', 'communication', 'medical'
    team_lead_name VARCHAR(255) NOT NULL,
    team_lead_phone VARCHAR(20),
    team_lead_email VARCHAR(255),
    team_size INTEGER NOT NULL DEFAULT 1,
    specializations TEXT[], -- Array of specializations
    base_location VARCHAR(255),
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by INTEGER REFERENCES users(user_id)
);

-- RRT Deployments
CREATE TABLE IF NOT EXISTS rrt_deployments (
    id SERIAL PRIMARY KEY,
    team_id INTEGER NOT NULL REFERENCES rrt_teams(id),
    outbreak_id INTEGER NOT NULL REFERENCES outbreaks(id),
    deployment_date DATE NOT NULL,
    expected_return_date DATE,
    actual_return_date DATE,
    deployment_status VARCHAR(50) DEFAULT 'deployed', -- 'deployed', 'returned', 'extended'
    deployment_purpose TEXT,
    assigned_vehicle VARCHAR(255),
    assigned_driver VARCHAR(255),
    deployment_notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by INTEGER REFERENCES users(user_id)
);

-- Resource Categories (enhanced from inventory_categories)
CREATE TABLE IF NOT EXISTS resource_categories (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    category_type VARCHAR(100) NOT NULL, -- 'medical', 'lab', 'logistics', 'it', 'ppe', 'fuel', 'stationery'
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by INTEGER REFERENCES users(user_id)
);

-- Resources (enhanced from inventory_items)
CREATE TABLE IF NOT EXISTS resources (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    resource_code VARCHAR(100) UNIQUE,
    category_id INTEGER NOT NULL REFERENCES resource_categories(id),
    unit_of_measure VARCHAR(50) NOT NULL,
    is_consumable BOOLEAN DEFAULT true,
    has_expiry BOOLEAN DEFAULT false,
    shelf_life_days INTEGER,
    is_critical BOOLEAN DEFAULT false,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by INTEGER REFERENCES users(user_id)
);

-- Storage Locations
CREATE TABLE IF NOT EXISTS storage_locations (
    id SERIAL PRIMARY KEY,
    location_name VARCHAR(255) NOT NULL,
    location_code VARCHAR(50) UNIQUE,
    location_type VARCHAR(100) NOT NULL, -- 'warehouse', 'medical_store', 'district_hub', 'field_storage'
    address TEXT,
    contact_person VARCHAR(255),
    contact_phone VARCHAR(20),
    contact_email VARCHAR(255),
    capacity_description TEXT,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by INTEGER REFERENCES users(user_id)
);

-- Donors/Source Entities
CREATE TABLE IF NOT EXISTS donors (
    id SERIAL PRIMARY KEY,
    donor_name VARCHAR(255) NOT NULL,
    donor_type VARCHAR(100) NOT NULL, -- 'government', 'ngo', 'international', 'private', 'individual'
    contact_person VARCHAR(255),
    contact_phone VARCHAR(20),
    contact_email VARCHAR(255),
    address TEXT,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by INTEGER REFERENCES users(user_id)
);

-- Stock Ledger (main tracking table)
CREATE TABLE IF NOT EXISTS stock_ledger (
    id SERIAL PRIMARY KEY,
    resource_id INTEGER NOT NULL REFERENCES resources(id),
    location_id INTEGER NOT NULL REFERENCES storage_locations(id),
    transaction_type VARCHAR(50) NOT NULL, -- 'received', 'allocated', 'used', 'returned', 'transferred', 'expired', 'damaged'
    quantity DECIMAL(10,2) NOT NULL,
    unit_cost DECIMAL(10,2), -- For tracking purposes only, not financial
    donor_id INTEGER REFERENCES donors(id),
    outbreak_id INTEGER REFERENCES outbreaks(id),
    deployment_id INTEGER REFERENCES rrt_deployments(id),
    reference_number VARCHAR(100),
    transaction_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expiry_date DATE,
    batch_number VARCHAR(100),
    notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by INTEGER REFERENCES users(user_id)
);

-- Requisitions
CREATE TABLE IF NOT EXISTS requisitions (
    id SERIAL PRIMARY KEY,
    requisition_number VARCHAR(100) UNIQUE NOT NULL,
    outbreak_id INTEGER NOT NULL REFERENCES outbreaks(id),
    deployment_id INTEGER REFERENCES rrt_deployments(id),
    requested_by INTEGER NOT NULL REFERENCES users(user_id),
    requested_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    required_date DATE,
    priority VARCHAR(50) DEFAULT 'normal', -- 'low', 'normal', 'high', 'urgent'
    status VARCHAR(50) DEFAULT 'pending', -- 'pending', 'approved', 'rejected', 'dispatched', 'completed'
    approved_by INTEGER REFERENCES users(user_id),
    approved_date TIMESTAMP,
    rejection_reason TEXT,
    dispatch_date TIMESTAMP,
    received_date TIMESTAMP,
    notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Requisition Items
CREATE TABLE IF NOT EXISTS requisition_items (
    id SERIAL PRIMARY KEY,
    requisition_id INTEGER NOT NULL REFERENCES requisitions(id),
    resource_id INTEGER NOT NULL REFERENCES resources(id),
    quantity_requested DECIMAL(10,2) NOT NULL,
    quantity_approved DECIMAL(10,2),
    quantity_dispatched DECIMAL(10,2),
    quantity_received DECIMAL(10,2),
    unit_cost DECIMAL(10,2),
    notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Dispatches
CREATE TABLE IF NOT EXISTS dispatches (
    id SERIAL PRIMARY KEY,
    dispatch_number VARCHAR(100) UNIQUE NOT NULL,
    requisition_id INTEGER NOT NULL REFERENCES requisitions(id),
    from_location_id INTEGER NOT NULL REFERENCES storage_locations(id),
    to_location_id INTEGER REFERENCES storage_locations(id),
    to_deployment_id INTEGER REFERENCES rrt_deployments(id),
    dispatched_by INTEGER NOT NULL REFERENCES users(user_id),
    dispatch_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expected_delivery_date DATE,
    actual_delivery_date TIMESTAMP,
    transport_method VARCHAR(100),
    vehicle_details VARCHAR(255),
    driver_name VARCHAR(255),
    driver_phone VARCHAR(20),
    delivery_notes TEXT,
    acknowledgment_received BOOLEAN DEFAULT false,
    acknowledgment_date TIMESTAMP,
    status VARCHAR(50) DEFAULT 'dispatched', -- 'dispatched', 'in_transit', 'delivered', 'returned'
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Dispatch Items
CREATE TABLE IF NOT EXISTS dispatch_items (
    id SERIAL PRIMARY KEY,
    dispatch_id INTEGER NOT NULL REFERENCES dispatches(id),
    resource_id INTEGER NOT NULL REFERENCES resources(id),
    quantity_dispatched DECIMAL(10,2) NOT NULL,
    batch_number VARCHAR(100),
    expiry_date DATE,
    unit_cost DECIMAL(10,2),
    notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Activity Logs
CREATE TABLE IF NOT EXISTS activity_logs (
    id SERIAL PRIMARY KEY,
    deployment_id INTEGER NOT NULL REFERENCES rrt_deployments(id),
    activity_type VARCHAR(100) NOT NULL, -- 'investigation', 'sensitization', 'meeting', 'training', 'assessment'
    activity_date DATE NOT NULL,
    start_time TIME,
    end_time TIME,
    location VARCHAR(255),
    participants_count INTEGER,
    activity_description TEXT,
    outcomes TEXT,
    challenges TEXT,
    recommendations TEXT,
    resources_used TEXT, -- JSON or text description of resources consumed
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by INTEGER REFERENCES users(user_id)
);

-- Activity Participants
CREATE TABLE IF NOT EXISTS activity_participants (
    id SERIAL PRIMARY KEY,
    activity_id INTEGER NOT NULL REFERENCES activity_logs(id),
    participant_name VARCHAR(255) NOT NULL,
    participant_type VARCHAR(100), -- 'team_member', 'community_leader', 'health_worker', 'volunteer'
    organization VARCHAR(255),
    contact_phone VARCHAR(20),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- SitRep Templates
CREATE TABLE IF NOT EXISTS sitrep_templates (
    id SERIAL PRIMARY KEY,
    template_name VARCHAR(255) NOT NULL,
    template_type VARCHAR(100) NOT NULL, -- 'weekly', 'monthly', 'outbreak_specific'
    outbreak_id INTEGER REFERENCES outbreaks(id),
    is_active BOOLEAN DEFAULT true,
    template_content TEXT, -- JSON structure for the template
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by INTEGER REFERENCES users(user_id)
);

-- Generated SitReps
CREATE TABLE IF NOT EXISTS generated_sitreps (
    id SERIAL PRIMARY KEY,
    outbreak_id INTEGER NOT NULL REFERENCES outbreaks(id),
    report_period_start DATE NOT NULL,
    report_period_end DATE NOT NULL,
    report_type VARCHAR(100) NOT NULL, -- 'weekly', 'monthly', 'final'
    report_title VARCHAR(255),
    report_content TEXT, -- JSON or HTML content
    generated_by INTEGER REFERENCES users(user_id),
    generated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    approved_by INTEGER REFERENCES users(user_id),
    approved_at TIMESTAMP,
    status VARCHAR(50) DEFAULT 'draft', -- 'draft', 'approved', 'published'
    file_path VARCHAR(500), -- Path to generated PDF/HTML file
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Outbreak Closure Criteria
CREATE TABLE IF NOT EXISTS outbreak_closure_criteria (
    id SERIAL PRIMARY KEY,
    outbreak_id INTEGER NOT NULL REFERENCES outbreaks(id),
    no_new_cases_days INTEGER DEFAULT 21,
    all_teams_demobilized BOOLEAN DEFAULT false,
    all_items_returned BOOLEAN DEFAULT false,
    final_sitrep_generated BOOLEAN DEFAULT false,
    closure_approved BOOLEAN DEFAULT false,
    closure_date DATE,
    closure_notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by INTEGER REFERENCES users(user_id)
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_rrt_deployments_outbreak ON rrt_deployments(outbreak_id);
CREATE INDEX IF NOT EXISTS idx_stock_ledger_resource ON stock_ledger(resource_id);
CREATE INDEX IF NOT EXISTS idx_stock_ledger_outbreak ON stock_ledger(outbreak_id);
CREATE INDEX IF NOT EXISTS idx_stock_ledger_deployment ON stock_ledger(deployment_id);
CREATE INDEX IF NOT EXISTS idx_requisitions_outbreak ON requisitions(outbreak_id);
CREATE INDEX IF NOT EXISTS idx_activity_logs_deployment ON activity_logs(deployment_id);
CREATE INDEX IF NOT EXISTS idx_generated_sitreps_outbreak ON generated_sitreps(outbreak_id);

-- Insert default resource categories
INSERT INTO resource_categories (name, description, category_type, created_by) VALUES
('Personal Protective Equipment', 'PPE items for outbreak response', 'ppe', 1),
('Medical Supplies', 'General medical supplies and equipment', 'medical', 1),
('Laboratory Supplies', 'Lab equipment, reagents, and testing materials', 'lab', 1),
('Logistics Equipment', 'Transport and logistics equipment', 'logistics', 1),
('IT Equipment', 'Computers, phones, and communication equipment', 'it', 1),
('Fuel and Transport', 'Fuel, vehicles, and transport resources', 'fuel', 1),
('Stationery and Forms', 'Forms, stationery, and documentation materials', 'stationery', 1),
('Food and Accommodation', 'Food supplies and accommodation resources', 'food', 1)
ON CONFLICT (name) DO NOTHING;

-- Insert default storage locations
INSERT INTO storage_locations (location_name, location_code, location_type, contact_person, created_by) VALUES
('Central Medical Store', 'CMS-001', 'warehouse', 'Store Manager', 1),
('District Health Office Store', 'DHO-001', 'medical_store', 'District Health Officer', 1),
('Regional Warehouse', 'RW-001', 'warehouse', 'Regional Manager', 1),
('Field Storage Unit', 'FSU-001', 'field_storage', 'Field Coordinator', 1)
ON CONFLICT (location_code) DO NOTHING;

-- Insert default donors
INSERT INTO donors (donor_name, donor_type, contact_person, created_by) VALUES
('National Medical Stores', 'government', 'NMS Manager', 1),
('Joint Medical Stores', 'government', 'JMS Coordinator', 1),
('World Health Organization', 'international', 'WHO Representative', 1),
('UNICEF', 'international', 'UNICEF Representative', 1),
('Ministry of Health', 'government', 'MoH Logistics Officer', 1)
ON CONFLICT (donor_name) DO NOTHING;

COMMENT ON TABLE rrt_teams IS 'Rapid Response Teams for outbreak deployment';
COMMENT ON TABLE rrt_deployments IS 'RRT deployment records linked to outbreaks';
COMMENT ON TABLE resources IS 'Master catalog of all resources and supplies';
COMMENT ON TABLE stock_ledger IS 'Main ledger tracking all resource movements';
COMMENT ON TABLE requisitions IS 'Resource requisition requests';
COMMENT ON TABLE dispatches IS 'Resource dispatch records';
COMMENT ON TABLE activity_logs IS 'RRT activity logging for SitRep generation';
COMMENT ON TABLE generated_sitreps IS 'Auto-generated situation reports';
