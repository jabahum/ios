-- Create districts table
CREATE TABLE IF NOT EXISTS districts (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    code VARCHAR(100) UNIQUE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create subcounties table
CREATE TABLE IF NOT EXISTS subcounties (
    id SERIAL PRIMARY KEY,
    district_id VARCHAR(100) REFERENCES districts(code) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    code VARCHAR(100) UNIQUE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create parishes table
CREATE TABLE IF NOT EXISTS parishes (
    id SERIAL PRIMARY KEY,
    district_id VARCHAR(100) REFERENCES districts(code) ON DELETE CASCADE,
    subcounty_id VARCHAR(100) REFERENCES subcounties(code) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    code VARCHAR(100) UNIQUE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create villages table
CREATE TABLE IF NOT EXISTS villages (
    id SERIAL PRIMARY KEY,
    district_id VARCHAR(100) REFERENCES districts(code) ON DELETE CASCADE,
    subcounty_id VARCHAR(100) REFERENCES subcounties(code) ON DELETE CASCADE,
    parish_id VARCHAR(100) REFERENCES parishes(code) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    code VARCHAR(100) UNIQUE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_subcounties_district_id ON subcounties(district_id);
CREATE INDEX IF NOT EXISTS idx_parishes_district_id ON parishes(district_id);
CREATE INDEX IF NOT EXISTS idx_parishes_subcounty_id ON parishes(subcounty_id);
CREATE INDEX IF NOT EXISTS idx_villages_district_id ON villages(district_id);
CREATE INDEX IF NOT EXISTS idx_villages_subcounty_id ON villages(subcounty_id);
CREATE INDEX IF NOT EXISTS idx_villages_parish_id ON villages(parish_id);

-- Insert sample data for districts
INSERT INTO districts (name, code) VALUES
('Kampala', 'KMP'),
('Wakiso', 'WAK'),
('Mukono', 'MKN'),
('Jinja', 'JNJ'),
('Mbale', 'MBL');

-- Insert sample data for subcounties
INSERT INTO subcounties (district_id, name, code) VALUES
('KMP', 'Central Division', 'KMP-CENT'),
('KMP', 'Nakawa Division', 'KMP-NAK'),
('KMP', 'Makindye Division', 'KMP-MAK'),
('WAK', 'Entebbe Municipality', 'WAK-ENT'),
('WAK', 'Nansana Municipality', 'WAK-NAN');

-- Insert sample data for parishes
INSERT INTO parishes (district_id, subcounty_id, name, code) VALUES
('KMP', 'KMP-CENT', 'Nakasero', 'KMP-CENT-NAK'),
('KMP', 'KMP-CENT', 'Kololo', 'KMP-CENT-KOL'),
('KMP', 'KMP-NAK', 'Nakawa', 'KMP-NAK-NAK'),
('WAK', 'WAK-ENT', 'Entebbe Central', 'WAK-ENT-CENT'),
('WAK', 'WAK-NAN', 'Nansana Central', 'WAK-NAN-CENT');

-- Insert sample data for villages
INSERT INTO villages (district_id, subcounty_id, parish_id, name, code) VALUES
('KMP', 'KMP-CENT', 'KMP-CENT-NAK', 'Nakasero Hill', 'KMP-CENT-NAK-HILL'),
('KMP', 'KMP-CENT', 'KMP-CENT-KOL', 'Kololo Hill', 'KMP-CENT-KOL-HILL'),
('KMP', 'KMP-NAK', 'KMP-NAK-NAK', 'Nakawa Market', 'KMP-NAK-NAK-MKT'),
('WAK', 'WAK-ENT', 'WAK-ENT-CENT', 'Entebbe Town', 'WAK-ENT-CENT-TOWN'),
('WAK', 'WAK-NAN', 'WAK-NAN-CENT', 'Nansana Town', 'WAK-NAN-CENT-TOWN'); 