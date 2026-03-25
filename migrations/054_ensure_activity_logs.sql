-- Migration 054: RRT team members system (029) + core RRT tables + activity_logs
--
-- Intended for databases where 028/029/052 were skipped, failed, or only partially applied.
-- Idempotent: uses CREATE TABLE IF NOT EXISTS, CREATE OR REPLACE FUNCTION, DROP TRIGGER IF EXISTS.
--
-- Prerequisites: public.outbreaks, public.users
--
-- Apply the whole file (do not paste from the middle):
--   ./scripts/apply-migration-054.sh
--   psql -v ON_ERROR_STOP=1 "$DATABASE_URL" -f migrations/054_ensure_activity_logs.sql

-- ---------------------------------------------------------------------------
-- Core RRT (normally from 028) — required by assignments, extensions, activity_logs
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS rrt_teams (
    id SERIAL PRIMARY KEY,
    team_name VARCHAR(255) NOT NULL,
    team_code VARCHAR(50) UNIQUE NOT NULL,
    team_type VARCHAR(100) NOT NULL,
    team_lead_name VARCHAR(255) NOT NULL,
    team_lead_phone VARCHAR(20),
    team_lead_email VARCHAR(255),
    team_size INTEGER NOT NULL DEFAULT 1,
    specializations TEXT[],
    base_location VARCHAR(255),
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by INTEGER REFERENCES users(user_id)
);

CREATE TABLE IF NOT EXISTS rrt_deployments (
    id SERIAL PRIMARY KEY,
    team_id INTEGER NOT NULL REFERENCES rrt_teams(id),
    outbreak_id INTEGER NOT NULL REFERENCES outbreaks(id),
    deployment_date DATE NOT NULL,
    expected_return_date DATE,
    actual_return_date DATE,
    deployment_status VARCHAR(50) DEFAULT 'deployed',
    deployment_purpose TEXT,
    assigned_vehicle VARCHAR(255),
    assigned_driver VARCHAR(255),
    deployment_notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by INTEGER REFERENCES users(user_id)
);

CREATE INDEX IF NOT EXISTS idx_rrt_deployments_outbreak ON rrt_deployments(outbreak_id);

COMMENT ON TABLE rrt_teams IS 'Rapid Response Teams for outbreak deployment';
COMMENT ON TABLE rrt_deployments IS 'RRT deployment records linked to outbreaks';

-- ---------------------------------------------------------------------------
-- RRT team members system (aligned with 029_create_rrt_team_members_system.sql)
-- ---------------------------------------------------------------------------
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
    specializations TEXT[],
    certifications TEXT,
    experience_years INTEGER DEFAULT 0,
    is_driver BOOLEAN DEFAULT FALSE,
    driver_license VARCHAR(50),
    driver_license_expiry DATE,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by INTEGER REFERENCES users(user_id)
);

CREATE TABLE IF NOT EXISTS rrt_team_member_assignments (
    id SERIAL PRIMARY KEY,
    team_id INTEGER NOT NULL REFERENCES rrt_teams(id) ON DELETE CASCADE,
    member_id INTEGER NOT NULL REFERENCES rrt_team_members(id) ON DELETE CASCADE,
    role VARCHAR(100) NOT NULL,
    start_date DATE NOT NULL,
    end_date DATE,
    is_active BOOLEAN DEFAULT TRUE,
    assignment_notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by INTEGER REFERENCES users(user_id)
);

CREATE TABLE IF NOT EXISTS rrt_deployment_proposals (
    id SERIAL PRIMARY KEY,
    proposal_number VARCHAR(50) UNIQUE NOT NULL,
    outbreak_id INTEGER NOT NULL REFERENCES outbreaks(id) ON DELETE CASCADE,
    proposed_by INTEGER NOT NULL REFERENCES users(user_id),
    proposed_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deployment_purpose TEXT NOT NULL,
    proposed_team_composition JSONB,
    required_skills TEXT[],
    deployment_duration_days INTEGER,
    expected_start_date DATE,
    expected_end_date DATE,
    special_requirements TEXT,
    justification TEXT NOT NULL,
    status VARCHAR(20) DEFAULT 'pending',
    reviewed_by INTEGER REFERENCES users(user_id),
    reviewed_at TIMESTAMP,
    review_notes TEXT,
    rejection_reason TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

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

CREATE TABLE IF NOT EXISTS rrt_deployment_extensions (
    id SERIAL PRIMARY KEY,
    deployment_id INTEGER NOT NULL REFERENCES rrt_deployments(id) ON DELETE CASCADE,
    extension_reason TEXT NOT NULL,
    original_end_date DATE NOT NULL,
    new_end_date DATE NOT NULL,
    requested_by INTEGER NOT NULL REFERENCES users(user_id),
    requested_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    approved_by INTEGER REFERENCES users(user_id),
    approved_date TIMESTAMP,
    status VARCHAR(20) DEFAULT 'pending',
    approval_notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS rrt_field_role_assignments (
    id SERIAL PRIMARY KEY,
    deployment_id INTEGER NOT NULL REFERENCES rrt_deployments(id) ON DELETE CASCADE,
    member_id INTEGER NOT NULL REFERENCES rrt_team_members(id) ON DELETE CASCADE,
    additional_role VARCHAR(100) NOT NULL,
    assignment_date DATE NOT NULL,
    end_date DATE,
    assigned_by INTEGER NOT NULL REFERENCES users(user_id),
    assignment_reason TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_rrt_team_members_active ON rrt_team_members(is_active);
CREATE INDEX IF NOT EXISTS idx_rrt_team_members_driver ON rrt_team_members(is_driver);
CREATE INDEX IF NOT EXISTS idx_rrt_team_member_assignments_team ON rrt_team_member_assignments(team_id);
CREATE INDEX IF NOT EXISTS idx_rrt_team_member_assignments_member ON rrt_team_member_assignments(member_id);
CREATE INDEX IF NOT EXISTS idx_rrt_team_member_assignments_active ON rrt_team_member_assignments(is_active);
CREATE INDEX IF NOT EXISTS idx_rrt_deployment_proposals_outbreak ON rrt_deployment_proposals(outbreak_id);
CREATE INDEX IF NOT EXISTS idx_rrt_deployment_proposals_status ON rrt_deployment_proposals(status);
CREATE INDEX IF NOT EXISTS idx_rrt_deployment_extensions_deployment ON rrt_deployment_extensions(deployment_id);
CREATE INDEX IF NOT EXISTS idx_rrt_field_role_assignments_deployment ON rrt_field_role_assignments(deployment_id);

COMMENT ON TABLE rrt_team_members IS 'Master list of all RRT team members with their qualifications and specializations';
COMMENT ON TABLE rrt_team_member_assignments IS 'Tracks team membership assignments over time with start/end dates';
COMMENT ON TABLE rrt_deployment_proposals IS 'Proposed team deployments that require approval before execution';
COMMENT ON TABLE rrt_deployment_proposal_members IS 'Specific team members proposed for each deployment';
COMMENT ON TABLE rrt_deployment_extensions IS 'Tracks deployment extensions with approval workflow';
COMMENT ON TABLE rrt_field_role_assignments IS 'Additional roles assigned to team members while deployed in the field';

-- Optional sample members (029): only if table is still empty — safe to re-run
INSERT INTO rrt_team_members (first_name, last_name, email, phone, organization, position, specializations, is_driver, is_active)
SELECT v.first_name, v.last_name, v.email, v.phone, v.organization, v.position, v.specializations, v.is_driver, v.is_active
FROM (
    VALUES
        ('Wafula', 'Hannah Isaac', 'wafula.hannah@example.com', '+256700123456', 'Ministry of Health', 'Epidemiologist', ARRAY['epidemiology', 'surveillance']::TEXT[], false, true),
        ('Nakye', 'Bibiana', 'nakye.shortie@example.com', '+256700123457', 'National Medical Stores', 'Logistics Officer', ARRAY['logistics', 'supply_chain']::TEXT[], true, true),
        ('Dr. Michael', 'Turyasingura', 'michael.tall@example.com', '+256700123458', 'Mulago Hospital', 'Infectious Disease Specialist', ARRAY['medical', 'infectious_diseases']::TEXT[], false, true),
        ('Mukulupya Patrick', 'Ayebazibwe', 'patrick.mukulupya@example.com', '+256700123459', 'WHO Uganda', 'Communication Specialist', ARRAY['communication', 'public_health']::TEXT[], false, true),
        ('Savio', 'Brown', 'savio.brown@example.com', '+256700123460', 'Ministry of Health', 'Laboratory Technician', ARRAY['laboratory', 'diagnostics']::TEXT[], false, true)
) AS v(first_name, last_name, email, phone, organization, position, specializations, is_driver, is_active)
WHERE NOT EXISTS (SELECT 1 FROM rrt_team_members LIMIT 1);

-- Proposal numbers: robust generator (POSIX-safe; avoids \d in regex like original 029)
CREATE OR REPLACE FUNCTION generate_proposal_number() RETURNS TEXT AS $$
DECLARE
    d TEXT := TO_CHAR(CURRENT_DATE, 'YYYYMMDD');
    prefix TEXT := 'PROP-' || d || '-';
    next_n INTEGER;
BEGIN
    SELECT COALESCE(MAX(
        CASE
            WHEN proposal_number LIKE prefix || '%'
                 AND length(proposal_number) > length(prefix)
                 AND substring(proposal_number FROM length(prefix) + 1) ~ '^[0-9]+$'
            THEN CAST(substring(proposal_number FROM length(prefix) + 1) AS INTEGER)
            ELSE NULL
        END
    ), 0) + 1
    INTO next_n
    FROM rrt_deployment_proposals;

    RETURN prefix || LPAD(next_n::TEXT, 4, '0');
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION set_proposal_number() RETURNS TRIGGER AS $$
BEGIN
    IF NEW.proposal_number IS NULL OR btrim(NEW.proposal_number) = '' THEN
        NEW.proposal_number := generate_proposal_number();
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_set_proposal_number ON rrt_deployment_proposals;
CREATE TRIGGER trigger_set_proposal_number
    BEFORE INSERT ON rrt_deployment_proposals
    FOR EACH ROW
    EXECUTE PROCEDURE set_proposal_number();

-- ---------------------------------------------------------------------------
-- Activity logs (resource management API / SitRep support)
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS activity_logs (
    id SERIAL PRIMARY KEY,
    deployment_id INTEGER NOT NULL REFERENCES rrt_deployments(id),
    activity_type VARCHAR(100) NOT NULL,
    activity_date DATE NOT NULL,
    start_time TIME,
    end_time TIME,
    location VARCHAR(255),
    participants_count INTEGER,
    activity_description TEXT,
    outcomes TEXT,
    challenges TEXT,
    recommendations TEXT,
    resources_used TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by INTEGER REFERENCES users(user_id)
);

CREATE TABLE IF NOT EXISTS activity_participants (
    id SERIAL PRIMARY KEY,
    activity_id INTEGER NOT NULL REFERENCES activity_logs(id) ON DELETE CASCADE,
    participant_name VARCHAR(255) NOT NULL,
    participant_type VARCHAR(100),
    organization VARCHAR(255),
    contact_phone VARCHAR(20),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_activity_logs_deployment ON activity_logs(deployment_id);

COMMENT ON TABLE activity_logs IS 'RRT activity logging (e.g. SitRep support)';
COMMENT ON TABLE activity_participants IS 'Participants linked to an activity log entry';
