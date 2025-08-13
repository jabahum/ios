-- Migration 045: Create lab sample types and their specific options
-- This migration adds tables to support dynamic lab sample type selection

-- Create table for swab types
CREATE TABLE IF NOT EXISTS lab_swab_types (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create table for urine test types
CREATE TABLE IF NOT EXISTS lab_urine_types (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create table for blood test types
CREATE TABLE IF NOT EXISTS lab_blood_types (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    category VARCHAR(50), -- To group related tests
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create table for lab sample selections (to store user selections)
CREATE TABLE IF NOT EXISTS lab_sample_selections (
    id SERIAL PRIMARY KEY,
    lab_id INTEGER REFERENCES lab(lab_id) ON DELETE CASCADE,
    sample_type VARCHAR(20) NOT NULL, -- 'swab', 'urine', 'blood'
    selected_type_id INTEGER, -- References the specific type table
    other_specify TEXT, -- For "Others (Specify)" options
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Insert swab types
INSERT INTO lab_swab_types (name, description) VALUES
('Wound swab', 'Swab from wound site'),
('Eye swab', 'Swab from eye area'),
('Ear swab', 'Swab from ear canal'),
('Oropharyngeal swab', 'Swab from throat area'),
('High vaginal swab', 'Swab from vaginal area'),
('Anorectal swab', 'Swab from anal/rectal area');

-- Insert urine test types
INSERT INTO lab_urine_types (name, description) VALUES
('Chemistry', 'Urine chemistry analysis'),
('Microscopy', 'Urine microscopy examination'),
('Culture and Sensitivity', 'Urine culture and antibiotic sensitivity testing');

-- Insert blood test types
INSERT INTO lab_blood_types (name, description, category) VALUES
-- CBC related
('CBC', 'Complete Blood Count', 'CBC'),
-- RFT related
('RFT', 'Renal Function Tests', 'RFT'),
-- LFT related
('LFT Profile', 'Liver Function Tests Profile', 'LFT'),
-- Cardiac related
('Cardiac Enzymes', 'Cardiac enzyme tests', 'Cardiac'),
-- Electrolytes
('Na+', 'Sodium', 'Electrolytes'),
('K+', 'Potassium', 'Electrolytes'),
('Cl-', 'Chloride', 'Electrolytes'),
('Mg2+', 'Magnesium', 'Electrolytes'),
('Ca2+', 'Calcium', 'Electrolytes'),
-- Culture
('Blood Culture', 'Blood culture for infection', 'Culture'),
-- Iron
('Fe2+', 'Iron studies', 'Iron'),
-- CSF
('CSF', 'Cerebrospinal Fluid analysis', 'CSF'),
-- Aspirates
('Pleural Fluid', 'Pleural fluid analysis', 'Aspirates'),
('Ascitic Fluid', 'Ascitic fluid analysis', 'Aspirates'),
('Lesion Fluid', 'Lesion fluid analysis', 'Aspirates'),
('Pus', 'Pus analysis', 'Aspirates'),
-- Others
('Others', 'Other blood tests', 'Others');

-- Add indexes for better performance
CREATE INDEX idx_lab_sample_selections_lab_id ON lab_sample_selections(lab_id);
CREATE INDEX idx_lab_sample_selections_sample_type ON lab_sample_selections(sample_type);
CREATE INDEX idx_lab_blood_types_category ON lab_blood_types(category); 