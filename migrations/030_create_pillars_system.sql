-- Create pillars table
CREATE TABLE IF NOT EXISTS pillars (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    pillar_head_id INTEGER REFERENCES users(id),
    pillar_head_name VARCHAR(255),
    pillar_head_email VARCHAR(255),
    pillar_head_phone VARCHAR(50),
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by INTEGER REFERENCES users(id),
    updated_by INTEGER REFERENCES users(id)
);

-- Create pillar changes tracking table
CREATE TABLE IF NOT EXISTS pillar_changes (
    id SERIAL PRIMARY KEY,
    pillar_id INTEGER NOT NULL REFERENCES pillars(id) ON DELETE CASCADE,
    change_type VARCHAR(50) NOT NULL, -- 'head_change', 'status_change', 'info_update'
    old_value TEXT,
    new_value TEXT,
    change_reason TEXT,
    changed_by INTEGER REFERENCES users(id),
    changed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    notes TEXT
);

-- Add pillar_id to rrt_team_members table
ALTER TABLE rrt_team_members ADD COLUMN IF NOT EXISTS pillar_id INTEGER REFERENCES pillars(id);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_pillars_active ON pillars(is_active);
CREATE INDEX IF NOT EXISTS idx_pillars_head ON pillars(pillar_head_id);
CREATE INDEX IF NOT EXISTS idx_pillar_changes_pillar ON pillar_changes(pillar_id);
CREATE INDEX IF NOT EXISTS idx_pillar_changes_date ON pillar_changes(changed_at);
CREATE INDEX IF NOT EXISTS idx_rrt_team_members_pillar ON rrt_team_members(pillar_id);

-- Insert some sample pillars
INSERT INTO pillars (name, description, pillar_head_name, pillar_head_email, pillar_head_phone, is_active, created_by) VALUES
('Epidemiology', 'Disease surveillance, outbreak investigation, and epidemiological analysis', 'Dr. Sarah Johnson', 'sarah.johnson@health.gov', '+256-700-123456', true, 1),
('Laboratory', 'Diagnostic testing, sample analysis, and laboratory quality assurance', 'Dr. Michael Chen', 'michael.chen@health.gov', '+256-700-234567', true, 1),
('Clinical Care', 'Patient management, treatment protocols, and clinical guidelines', 'Dr. Emily Rodriguez', 'emily.rodriguez@health.gov', '+256-700-345678', true, 1),
('Logistics', 'Supply chain management, equipment distribution, and resource allocation', 'Mr. David Kimani', 'david.kimani@health.gov', '+256-700-456789', true, 1),
('Communication', 'Public health messaging, risk communication, and community engagement', 'Ms. Grace Mwangi', 'grace.mwangi@health.gov', '+256-700-567890', true, 1),
('Data Management', 'Health information systems, data analysis, and reporting', 'Mr. James Ochieng', 'james.ochieng@health.gov', '+256-700-678901', true, 1)
ON CONFLICT DO NOTHING;
