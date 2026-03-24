-- Migration 052: RRT deployment proposals (with auto proposal_number) + activity_logs
--
-- DO NOT copy-paste only the trigger or activity_logs section — that always fails.
--
-- Apply the whole file using ONE of:
--   bash scripts/apply-migration-052.sh
--   pwsh scripts/apply-migration-052.ps1
--   psql -v ON_ERROR_STOP=1 "$DATABASE_URL" -f migrations/052_rrt_deployment_proposals_and_activity_logs.sql
--   (in psql, from repo root)  \i migrations/052_rrt_deployment_proposals_and_activity_logs.sql
--
-- Prerequisites: public.outbreaks and public.users must exist.

SELECT migration_052_start,
       instruction
FROM (
  SELECT NOW() AS migration_052_start,
         '052: execute FROM LINE 1 — creates rrt_teams, rrt_deployments, proposals, then trigger, then activity_logs' AS instruction
) AS _;

-- ---------------------------------------------------------------------------
-- Core RRT tables (same as start of 028 — needed for deployments + activity_logs FKs)
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
-- RRT team members (required by rrt_deployment_proposal_members FK)
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

-- ---------------------------------------------------------------------------
-- Deployment proposals
-- ---------------------------------------------------------------------------
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

COMMENT ON TABLE rrt_team_members IS 'Master list of RRT team members';
COMMENT ON TABLE rrt_team_member_assignments IS 'Team membership over time';
COMMENT ON TABLE rrt_deployment_proposals IS 'Proposed deployments pending approval';
COMMENT ON TABLE rrt_deployment_proposal_members IS 'Members proposed per deployment';
COMMENT ON TABLE rrt_deployment_extensions IS 'Deployment extension requests';
COMMENT ON TABLE rrt_field_role_assignments IS 'Extra field roles during deployment';

-- Auto-generate proposal_number when NULL or empty (matches app INSERT omitting column)
CREATE OR REPLACE FUNCTION generate_proposal_number() RETURNS TEXT AS $$
DECLARE
    d TEXT := TO_CHAR(CURRENT_DATE, 'YYYYMMDD');
    prefix TEXT := 'PROP-' || d || '-';
    suffix TEXT;
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

    suffix := LPAD(next_n::TEXT, 4, '0');
    RETURN prefix || suffix;
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
-- Activity logs (RRT / resource management) — safe if 028 already created them
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

-- =====================================================================
-- END migration 052
-- =====================================================================
-- If CREATE TRIGGER or CREATE activity_logs failed with "relation does not exist",
-- you pasted only the bottom of this file. Re-run the ENTIRE file from the top
-- (the SELECT + CREATE rrt_teams + … + trigger + activity_logs), e.g. in psql:
--   \i migrations/052_rrt_deployment_proposals_and_activity_logs.sql
-- =====================================================================
