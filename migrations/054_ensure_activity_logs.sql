-- Migration 054: Ensure activity_logs (and required rrt_teams / rrt_deployments)
--
-- Use when the API returns: pq: relation "activity_logs" does not exist
-- Prerequisites: public.outbreaks, public.users
--
-- Paste the ENTIRE file, or run:
--   psql -v ON_ERROR_STOP=1 "$DATABASE_URL" -f migrations/054_ensure_activity_logs.sql

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
