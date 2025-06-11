-- Create surveillance focal persons table
CREATE TABLE IF NOT EXISTS surveillance_focal_persons (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    phone VARCHAR(15) NOT NULL,
    district_id INTEGER REFERENCES districts(id) ON DELETE CASCADE,
    email VARCHAR(100),
    position VARCHAR(100),
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Add index for better performance
CREATE INDEX idx_surveillance_focal_persons_district_id ON surveillance_focal_persons(district_id);

-- Insert sample data
INSERT INTO surveillance_focal_persons (name, phone, district_id, email, position) VALUES 
('John Doe', '256783261162', 1, 'john.doe@health.go.ug', 'District Surveillance Officer'),
('Jane Smith', '256753475676', 2, 'jane.smith@health.go.ug', 'District Surveillance Officer'),
('Robert Johnson', '256772345678', 3, 'robert.johnson@health.go.ug', 'District Surveillance Officer'); 