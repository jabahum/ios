-- Create surveillance focal persons table
CREATE TABLE IF NOT EXISTS surveillance_focal_persons (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    phone VARCHAR(20) NOT NULL,
    district_id VARCHAR(100) REFERENCES districts(code) ON DELETE CASCADE,
    email VARCHAR(255),
    position VARCHAR(255),
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create index on district_id
CREATE INDEX IF NOT EXISTS idx_surveillance_focal_persons_district_id ON surveillance_focal_persons(district_id);

-- Insert some sample data
INSERT INTO surveillance_focal_persons (name, phone, district_id, email, position) VALUES
('John Doe', '+256123456789', 'KMP', 'john.doe@example.com', 'District Health Officer'),
('Jane Smith', '+256987654321', 'KMP', 'jane.smith@example.com', 'Public Health Officer'),
('Robert Johnson', '+256456789123', 'WAK', 'robert.johnson@example.com', 'District Health Officer'),
('Mary Williams', '+256789123456', 'WAK', 'mary.williams@example.com', 'Public Health Officer'); 