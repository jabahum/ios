-- Migration 029: Create RRT Team Members and Assignment System
-- This migration creates tables for managing RRT team members, their assignments, and deployment approvals

-- RRT Team Members table
CREATE TABLE IF NOT EXISTS rrt_team_members (
    id SERIAL PRIMARY KEY,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    email VARCHAR(255),
    phone VARCHAR(20),
    national_id VARCHAR(50),
    employee_id VARCHAR(50),
    organization VARCHAR(255),
    position VARCHAR(255),
    qualifications TEXT,
    specializations TEXT[], -- Array of specializations
    certifications TEXT,
    experience_years INTEGER DEFAULT 0,
    is_driver BOOLEAN DEFAULT FALSE,
    driver_license VARCHAR(50),
    driver_license_expiry DATE,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by INTEGER REFERENCES users(id)
);

-- RRT Team Member Assignments table (tracks team membership over time)
CREATE TABLE IF NOT EXISTS rrt_team_member_assignments (
    id SERIAL PRIMARY KEY,
    team_id INTEGER NOT NULL REFERENCES rrt_teams(id) ON DELETE CASCADE,
    member_id INTEGER NOT NULL REFERENCES rrt_team_members(id) ON DELETE CASCADE,
    role VARCHAR(100) NOT NULL, -- Team Lead, Deputy Lead, Member, Driver, etc.
    start_date DATE NOT NULL,
    end_date DATE,
    is_active BOOLEAN DEFAULT TRUE,
    assignment_notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by INTEGER REFERENCES users(id)
);

-- RRT Deployment Proposals table (for team deployment suggestions)
CREATE TABLE IF NOT EXISTS rrt_deployment_proposals (
    id SERIAL PRIMARY KEY,
    proposal_number VARCHAR(50) UNIQUE NOT NULL,
    outbreak_id INTEGER NOT NULL REFERENCES outbreaks(id) ON DELETE CASCADE,
    proposed_by INTEGER NOT NULL REFERENCES users(id),
    proposed_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deployment_purpose TEXT NOT NULL,
    proposed_team_composition JSONB, -- JSON structure with team members and roles
    required_skills TEXT[],
    deployment_duration_days INTEGER,
    expected_start_date DATE,
    expected_end_date DATE,
    special_requirements TEXT,
    justification TEXT NOT NULL,
    status VARCHAR(20) DEFAULT 'pending', -- pending, approved, rejected, modified
    reviewed_by INTEGER REFERENCES users(id),
    reviewed_at TIMESTAMP,
    review_notes TEXT,
    rejection_reason TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- RRT Deployment Proposal Members (specific members proposed for deployment)
CREATE TABLE IF NOT EXISTS rrt_deployment_proposal_members (
    id SERIAL PRIMARY KEY,
    proposal_id INTEGER NOT NULL REFERENCES rrt_deployment_proposals(id) ON DELETE CASCADE,
    member_id INTEGER NOT NULL REFERENCES rrt_team_members(id) ON DELETE CASCADE,
    proposed_role VARCHAR(100) NOT NULL,
    is_essential BOOLEAN DEFAULT FALSE,
    alternative_member_id INTEGER REFERENCES rrt_team_members(id),
    notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- RRT Deployment Extensions table
CREATE TABLE IF NOT EXISTS rrt_deployment_extensions (
    id SERIAL PRIMARY KEY,
    deployment_id INTEGER NOT NULL REFERENCES rrt_deployments(id) ON DELETE CASCADE,
    extension_reason TEXT NOT NULL,
    original_end_date DATE NOT NULL,
    new_end_date DATE NOT NULL,
    requested_by INTEGER NOT NULL REFERENCES users(id),
    requested_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    approved_by INTEGER REFERENCES users(id),
    approved_date TIMESTAMP,
    status VARCHAR(20) DEFAULT 'pending', -- pending, approved, rejected
    approval_notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- RRT Field Role Assignments (additional roles assigned while in field)
CREATE TABLE IF NOT EXISTS rrt_field_role_assignments (
    id SERIAL PRIMARY KEY,
    deployment_id INTEGER NOT NULL REFERENCES rrt_deployments(id) ON DELETE CASCADE,
    member_id INTEGER NOT NULL REFERENCES rrt_team_members(id) ON DELETE CASCADE,
    additional_role VARCHAR(100) NOT NULL,
    assignment_date DATE NOT NULL,
    end_date DATE,
    assigned_by INTEGER NOT NULL REFERENCES users(id),
    assignment_reason TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_rrt_team_members_active ON rrt_team_members(is_active);
CREATE INDEX IF NOT EXISTS idx_rrt_team_members_driver ON rrt_team_members(is_driver);
CREATE INDEX IF NOT EXISTS idx_rrt_team_member_assignments_team ON rrt_team_member_assignments(team_id);
CREATE INDEX IF NOT EXISTS idx_rrt_team_member_assignments_member ON rrt_team_member_assignments(member_id);
CREATE INDEX IF NOT EXISTS idx_rrt_team_member_assignments_active ON rrt_team_member_assignments(is_active);
CREATE INDEX IF NOT EXISTS idx_rrt_deployment_proposals_outbreak ON rrt_deployment_proposals(outbreak_id);
CREATE INDEX IF NOT EXISTS idx_rrt_deployment_proposals_status ON rrt_deployment_proposals(status);
CREATE INDEX IF NOT EXISTS idx_rrt_deployment_extensions_deployment ON rrt_deployment_extensions(deployment_id);
CREATE INDEX IF NOT EXISTS idx_rrt_field_role_assignments_deployment ON rrt_field_role_assignments(deployment_id);

-- Add comments for documentation
COMMENT ON TABLE rrt_team_members IS 'Master list of all RRT team members with their qualifications and specializations';
COMMENT ON TABLE rrt_team_member_assignments IS 'Tracks team membership assignments over time with start/end dates';
COMMENT ON TABLE rrt_deployment_proposals IS 'Proposed team deployments that require approval before execution';
COMMENT ON TABLE rrt_deployment_proposal_members IS 'Specific team members proposed for each deployment';
COMMENT ON TABLE rrt_deployment_extensions IS 'Tracks deployment extensions with approval workflow';
COMMENT ON TABLE rrt_field_role_assignments IS 'Additional roles assigned to team members while deployed in the field';

-- Insert some sample data
INSERT INTO rrt_team_members (first_name, last_name, email, phone, organization, position, specializations, is_driver, is_active) VALUES
('John', 'Doe', 'john.doe@example.com', '+256700123456', 'Ministry of Health', 'Epidemiologist', ARRAY['epidemiology', 'surveillance'], false, true),
('Jane', 'Smith', 'jane.smith@example.com', '+256700123457', 'National Medical Stores', 'Logistics Officer', ARRAY['logistics', 'supply_chain'], true, true),
('Dr. Michael', 'Johnson', 'michael.johnson@example.com', '+256700123458', 'Mulago Hospital', 'Infectious Disease Specialist', ARRAY['medical', 'infectious_diseases'], false, true),
('Sarah', 'Wilson', 'sarah.wilson@example.com', '+256700123459', 'WHO Uganda', 'Communication Specialist', ARRAY['communication', 'public_health'], false, true),
('David', 'Brown', 'david.brown@example.com', '+256700123460', 'Ministry of Health', 'Laboratory Technician', ARRAY['laboratory', 'diagnostics'], false, true);

-- Create a function to generate proposal numbers
CREATE OR REPLACE FUNCTION generate_proposal_number() RETURNS TEXT AS $$
DECLARE
    new_number TEXT;
    counter INTEGER;
BEGIN
    -- Get the current date in YYYYMMDD format
    new_number := 'PROP-' || TO_CHAR(CURRENT_DATE, 'YYYYMMDD') || '-';
    
    -- Get the count of proposals for today
    SELECT COALESCE(MAX(CAST(SUBSTRING(proposal_number FROM 'PROP-' || TO_CHAR(CURRENT_DATE, 'YYYYMMDD') || '-(\d+)$') AS INTEGER)), 0) + 1
    INTO counter
    FROM rrt_deployment_proposals
    WHERE proposal_number LIKE 'PROP-' || TO_CHAR(CURRENT_DATE, 'YYYYMMDD') || '-%';
    
    -- Format the counter with leading zeros
    new_number := new_number || LPAD(counter::TEXT, 4, '0');
    
    RETURN new_number;
END;
$$ LANGUAGE plpgsql;

-- Create a trigger to auto-generate proposal numbers
CREATE OR REPLACE FUNCTION set_proposal_number() RETURNS TRIGGER AS $$
BEGIN
    IF NEW.proposal_number IS NULL OR NEW.proposal_number = '' THEN
        NEW.proposal_number := generate_proposal_number();
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_set_proposal_number
    BEFORE INSERT ON rrt_deployment_proposals
    FOR EACH ROW
    EXECUTE FUNCTION set_proposal_number();
