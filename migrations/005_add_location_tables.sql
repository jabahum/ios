-- Create districts table
CREATE TABLE IF NOT EXISTS districts (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    code VARCHAR(100) UNIQUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create subcounties table
CREATE TABLE IF NOT EXISTS subcounties (
    id SERIAL PRIMARY KEY,
    district_id INTEGER REFERENCES districts(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    code VARCHAR(100) UNIQUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create parishes table
CREATE TABLE IF NOT EXISTS parishes (
    id SERIAL PRIMARY KEY,
    district_id INTEGER REFERENCES districts(id) ON DELETE CASCADE,
    subcounty_id INTEGER REFERENCES subcounties(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    code VARCHAR(100) UNIQUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create villages table
CREATE TABLE IF NOT EXISTS villages (
    id SERIAL PRIMARY KEY,
    district_id INTEGER REFERENCES districts(id) ON DELETE CASCADE,
    subcounty_id INTEGER REFERENCES subcounties(id) ON DELETE CASCADE,
    parish_id INTEGER REFERENCES parishes(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    code VARCHAR(100) UNIQUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Add indexes for better performance
CREATE INDEX idx_subcounties_district_id ON subcounties(district_id);
CREATE INDEX idx_parishes_district_id ON parishes(district_id);
CREATE INDEX idx_parishes_subcounty_id ON parishes(subcounty_id);
CREATE INDEX idx_villages_district_id ON villages(district_id);
CREATE INDEX idx_villages_subcounty_id ON villages(subcounty_id);
CREATE INDEX idx_villages_parish_id ON villages(parish_id);

-- Insert some sample data for testing
INSERT INTO districts (name, code) VALUES 
('Kampala', 'KLA'),
('Wakiso', 'WKS'),
('Mukono', 'MKN');

INSERT INTO subcounties (district_id, name, code) VALUES 
(1, 'Central Division', 'KLA-CENT'),
(1, 'Nakawa Division', 'KLA-NAK'),
(2, 'Entebbe', 'WKS-ENT'),
(2, 'Kira', 'WKS-KIR'),
(3, 'Mukono Central', 'MKN-CENT'),
(3, 'Nakifuma', 'MKN-NAK');

INSERT INTO parishes (district_id, subcounty_id, name, code) VALUES 
(1, 1, 'Nakasero', 'KLA-CENT-NAK'),
(1, 1, 'Kololo', 'KLA-CENT-KOL'),
(1, 2, 'Nakawa', 'KLA-NAK-NAK'),
(1, 2, 'Bugolobi', 'KLA-NAK-BUG'),
(2, 3, 'Entebbe Central', 'WKS-ENT-CENT'),
(2, 3, 'Katabi', 'WKS-ENT-KAT'),
(2, 4, 'Kira Central', 'WKS-KIR-CENT'),
(2, 4, 'Bweyogerere', 'WKS-KIR-BWE'),
(3, 5, 'Mukono Central', 'MKN-CENT-CENT'),
(3, 5, 'Ntawo', 'MKN-CENT-NTA'),
(3, 6, 'Nakifuma Central', 'MKN-NAK-CENT'),
(3, 6, 'Kasawo', 'MKN-NAK-KAS');

INSERT INTO villages (district_id, subcounty_id, parish_id, name, code) VALUES 
(1, 1, 1, 'Nakasero Hill', 'KLA-CENT-NAK-HILL'),
(1, 1, 1, 'Nakasero Valley', 'KLA-CENT-NAK-VAL'),
(1, 1, 2, 'Kololo Hill', 'KLA-CENT-KOL-HILL'),
(1, 1, 2, 'Kololo Valley', 'KLA-CENT-KOL-VAL'),
(1, 2, 3, 'Nakawa Market', 'KLA-NAK-NAK-MKT'),
(1, 2, 3, 'Nakawa Industrial', 'KLA-NAK-NAK-IND'),
(1, 2, 4, 'Bugolobi Flats', 'KLA-NAK-BUG-FLT'),
(1, 2, 4, 'Bugolobi Heights', 'KLA-NAK-BUG-HGT'),
(2, 3, 5, 'Entebbe Town', 'WKS-ENT-CENT-TWN'),
(2, 3, 5, 'Entebbe Port', 'WKS-ENT-CENT-PRT'),
(2, 3, 6, 'Katabi Central', 'WKS-ENT-KAT-CENT'),
(2, 3, 6, 'Katabi Market', 'WKS-ENT-KAT-MKT'),
(2, 4, 7, 'Kira Town', 'WKS-KIR-CENT-TWN'),
(2, 4, 7, 'Kira Market', 'WKS-KIR-CENT-MKT'),
(2, 4, 8, 'Bweyogerere Market', 'WKS-KIR-BWE-MKT'),
(2, 4, 8, 'Bweyogerere Central', 'WKS-KIR-BWE-CENT'),
(3, 5, 9, 'Mukono Town', 'MKN-CENT-CENT-TWN'),
(3, 5, 9, 'Mukono Market', 'MKN-CENT-CENT-MKT'),
(3, 5, 10, 'Ntawo Central', 'MKN-CENT-NTA-CENT'),
(3, 5, 10, 'Ntawo Market', 'MKN-CENT-NTA-MKT'),
(3, 6, 11, 'Nakifuma Town', 'MKN-NAK-CENT-TWN'),
(3, 6, 11, 'Nakifuma Market', 'MKN-NAK-CENT-MKT'),
(3, 6, 12, 'Kasawo Central', 'MKN-NAK-KAS-CENT'),
(3, 6, 12, 'Kasawo Market', 'MKN-NAK-KAS-MKT'); 