-- Migration script to update VHF patients table for new location schema
-- This script should be run after creating the new location tables

-- First, create the new location tables if they don't exist
CREATE TABLE IF NOT EXISTS districts (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    code VARCHAR(100) UNIQUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS subcounties (
    id SERIAL PRIMARY KEY,
    district_id INTEGER REFERENCES districts(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    code VARCHAR(100) UNIQUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS parishes (
    id SERIAL PRIMARY KEY,
    district_id INTEGER REFERENCES districts(id) ON DELETE CASCADE,
    subcounty_id INTEGER REFERENCES subcounties(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    code VARCHAR(100) UNIQUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS villages (
    id SERIAL PRIMARY KEY,
    district_id INTEGER REFERENCES districts(id) ON DELETE CASCADE,
    subcounty_id INTEGER REFERENCES subcounties(id) ON DELETE CASCADE,
    parish_id INTEGER REFERENCES parishes(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    code VARCHAR(100) UNIQUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Add new location ID columns to vhf_patients table
ALTER TABLE vhf_patients ADD COLUMN IF NOT EXISTS village_id INTEGER REFERENCES villages(id);
ALTER TABLE vhf_patients ADD COLUMN IF NOT EXISTS parish_id INTEGER REFERENCES parishes(id);
ALTER TABLE vhf_patients ADD COLUMN IF NOT EXISTS subcounty_id INTEGER REFERENCES subcounties(id);
ALTER TABLE vhf_patients ADD COLUMN IF NOT EXISTS district_id INTEGER REFERENCES districts(id);

-- Note: The old string columns (village_town, parish, subcounty, district) can be kept for backward compatibility
-- or removed after data migration is complete. For now, we'll keep them.

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_vhf_patients_village_id ON vhf_patients(village_id);
CREATE INDEX IF NOT EXISTS idx_vhf_patients_parish_id ON vhf_patients(parish_id);
CREATE INDEX IF NOT EXISTS idx_vhf_patients_subcounty_id ON vhf_patients(subcounty_id);
CREATE INDEX IF NOT EXISTS idx_vhf_patients_district_id ON vhf_patients(district_id);

-- Sample data insertion (you can modify these as needed)
INSERT INTO districts (name, code) VALUES 
('Kampala', 'KLA'),
('Wakiso', 'WKS'),
('Mukono', 'MKN')
ON CONFLICT (code) DO NOTHING;

INSERT INTO subcounties (district_id, name, code) VALUES 
((SELECT id FROM districts WHERE code = 'KLA'), 'Central Division', 'KLA_CENTRAL'),
((SELECT id FROM districts WHERE code = 'KLA'), 'Nakawa Division', 'KLA_NAKAWA'),
((SELECT id FROM districts WHERE code = 'WKS'), 'Entebbe Municipality', 'WKS_ENTEBBE'),
((SELECT id FROM districts WHERE code = 'WKS'), 'Kira Municipality', 'WKS_KIRA')
ON CONFLICT (code) DO NOTHING;

INSERT INTO parishes (district_id, subcounty_id, name, code) VALUES 
((SELECT id FROM districts WHERE code = 'KLA'), (SELECT id FROM subcounties WHERE code = 'KLA_CENTRAL'), 'Kampala Central Parish', 'KLA_CENTRAL_PARISH'),
((SELECT id FROM districts WHERE code = 'KLA'), (SELECT id FROM subcounties WHERE code = 'KLA_NAKAWA'), 'Nakawa Parish', 'KLA_NAKAWA_PARISH'),
((SELECT id FROM districts WHERE code = 'WKS'), (SELECT id FROM subcounties WHERE code = 'WKS_ENTEBBE'), 'Entebbe Parish', 'WKS_ENTEBBE_PARISH')
ON CONFLICT (code) DO NOTHING;

INSERT INTO villages (district_id, subcounty_id, parish_id, name, code) VALUES 
((SELECT id FROM districts WHERE code = 'KLA'), (SELECT id FROM subcounties WHERE code = 'KLA_CENTRAL'), (SELECT id FROM parishes WHERE code = 'KLA_CENTRAL_PARISH'), 'Kampala Central Village', 'KLA_CENTRAL_VILLAGE'),
((SELECT id FROM districts WHERE code = 'KLA'), (SELECT id FROM subcounties WHERE code = 'KLA_NAKAWA'), (SELECT id FROM parishes WHERE code = 'KLA_NAKAWA_PARISH'), 'Nakawa Village', 'KLA_NAKAWA_VILLAGE'),
((SELECT id FROM districts WHERE code = 'WKS'), (SELECT id FROM subcounties WHERE code = 'WKS_ENTEBBE'), (SELECT id FROM parishes WHERE code = 'WKS_ENTEBBE_PARISH'), 'Entebbe Village', 'WKS_ENTEBBE_VILLAGE')
ON CONFLICT (code) DO NOTHING; 