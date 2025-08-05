-- Migration 028: Add donation tracking to inventory system
-- This migration adds comprehensive donation tracking capabilities

-- Add donation-related tables and modify existing tables

-- 1. Create donors table
CREATE TABLE IF NOT EXISTS inventory_donors (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    organization VARCHAR(255),
    contact_person VARCHAR(255),
    phone VARCHAR(50),
    email VARCHAR(255),
    address TEXT,
    donor_type VARCHAR(50) DEFAULT 'individual', -- individual, organization, government, international
    country VARCHAR(100),
    registration_number VARCHAR(100),
    tax_exempt BOOLEAN DEFAULT false,
    status VARCHAR(50) DEFAULT 'active',
    notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 2. Create donation_types table
CREATE TABLE IF NOT EXISTS inventory_donation_types (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    is_monetary BOOLEAN DEFAULT false,
    status VARCHAR(50) DEFAULT 'active',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 3. Create donations table
CREATE TABLE IF NOT EXISTS inventory_donations (
    id SERIAL PRIMARY KEY,
    donor_id INTEGER REFERENCES inventory_donors(id),
    donation_type_id INTEGER REFERENCES inventory_donation_types(id),
    donation_date DATE NOT NULL,
    received_date DATE,
    description TEXT,
    monetary_value DECIMAL(15,2),
    currency VARCHAR(10) DEFAULT 'USD',
    donation_status VARCHAR(50) DEFAULT 'pending', -- pending, received, distributed, expired
    outbreak_id INTEGER REFERENCES outbreaks(id),
    treatment_site_id INTEGER REFERENCES treatment_sites(id),
    received_by INTEGER REFERENCES users(user_id),
    notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 4. Create donation_items table (for in-kind donations)
CREATE TABLE IF NOT EXISTS inventory_donation_items (
    id SERIAL PRIMARY KEY,
    donation_id INTEGER REFERENCES inventory_donations(id),
    item_id INTEGER REFERENCES inventory_items(id),
    quantity DECIMAL(10,2) NOT NULL,
    unit VARCHAR(50),
    estimated_value DECIMAL(15,2),
    condition_status VARCHAR(50) DEFAULT 'new', -- new, used, refurbished, expired
    expiry_date DATE,
    batch_number VARCHAR(100),
    serial_number VARCHAR(100),
    notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 5. Create donation_acknowledgments table
CREATE TABLE IF NOT EXISTS inventory_donation_acknowledgments (
    id SERIAL PRIMARY KEY,
    donation_id INTEGER REFERENCES inventory_donations(id),
    acknowledgment_type VARCHAR(50) DEFAULT 'letter', -- letter, email, certificate, receipt
    sent_date DATE,
    sent_by INTEGER REFERENCES users(user_id),
    recipient_name VARCHAR(255),
    recipient_email VARCHAR(255),
    acknowledgment_text TEXT,
    status VARCHAR(50) DEFAULT 'pending', -- pending, sent, delivered, confirmed
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 6. Modify inventory_items table to add donation-related fields
ALTER TABLE inventory_items ADD COLUMN IF NOT EXISTS is_donated BOOLEAN DEFAULT false;
ALTER TABLE inventory_items ADD COLUMN IF NOT EXISTS donor_id INTEGER REFERENCES inventory_donors(id);
ALTER TABLE inventory_items ADD COLUMN IF NOT EXISTS donation_date DATE;
ALTER TABLE inventory_items ADD COLUMN IF NOT EXISTS estimated_donor_value DECIMAL(15,2);
ALTER TABLE inventory_items ADD COLUMN IF NOT EXISTS condition_on_receipt VARCHAR(50) DEFAULT 'new';

-- 7. Modify inventory_transactions table to track donation sources
ALTER TABLE inventory_transactions ADD COLUMN IF NOT EXISTS donation_item_id INTEGER REFERENCES inventory_donation_items(id);
ALTER TABLE inventory_transactions ADD COLUMN IF NOT EXISTS is_donated_item BOOLEAN DEFAULT false;

-- 8. Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_inventory_donations_donor_id ON inventory_donations(donor_id);
CREATE INDEX IF NOT EXISTS idx_inventory_donations_outbreak_id ON inventory_donations(outbreak_id);
CREATE INDEX IF NOT EXISTS idx_inventory_donations_status ON inventory_donations(donation_status);
CREATE INDEX IF NOT EXISTS idx_inventory_donation_items_donation_id ON inventory_donation_items(donation_id);
CREATE INDEX IF NOT EXISTS idx_inventory_donation_items_item_id ON inventory_donation_items(item_id);
CREATE INDEX IF NOT EXISTS idx_inventory_items_donor_id ON inventory_items(donor_id);
CREATE INDEX IF NOT EXISTS idx_inventory_items_is_donated ON inventory_items(is_donated);

-- 9. Insert default donation types
INSERT INTO inventory_donation_types (name, description, is_monetary) VALUES
('Medical Supplies', 'Donation of medical supplies and equipment', false),
('Medications', 'Donation of pharmaceutical products', false),
('Personal Protective Equipment', 'PPE donations including masks, gloves, gowns', false),
('Laboratory Supplies', 'Lab equipment and testing supplies', false),
('Food and Nutrition', 'Food items and nutritional supplements', false),
('Transportation', 'Vehicle or transportation services', false),
('Cash Donation', 'Monetary donations', true),
('Equipment', 'Medical equipment and devices', false),
('Services', 'Professional services or volunteer time', false),
('Other', 'Other types of donations', false)
ON CONFLICT DO NOTHING;

-- 10. Insert sample donors
INSERT INTO inventory_donors (name, organization, donor_type, country) VALUES
('Red Cross', 'International Committee of the Red Cross', 'international', 'Switzerland'),
('WHO', 'World Health Organization', 'international', 'Switzerland'),
('USAID', 'United States Agency for International Development', 'government', 'United States'),
('UNICEF', 'United Nations Children''s Fund', 'international', 'United States'),
('Local Community', 'Local community members', 'individual', 'Uganda'),
('Pharmaceutical Company', 'Generic Pharma Ltd', 'organization', 'Uganda'),
('Ministry of Health', 'Uganda Ministry of Health', 'government', 'Uganda'),
('International NGO', 'Global Health Initiative', 'organization', 'United States')
ON CONFLICT DO NOTHING;

-- 11. Create view for donation summary
CREATE OR REPLACE VIEW inventory_donation_summary AS
SELECT 
    d.id as donation_id,
    d.donation_date,
    d.received_date,
    d.donation_status,
    dt.name as donation_type,
    dt.is_monetary,
    dr.name as donor_name,
    dr.organization as donor_organization,
    dr.donor_type,
    o.name as outbreak_name,
    ts.name as treatment_site_name,
    u.user_name as received_by_user,
    d.monetary_value,
    d.currency,
    COUNT(di.id) as item_count,
    SUM(di.quantity) as total_quantity,
    SUM(di.estimated_value) as total_estimated_value
FROM inventory_donations d
LEFT JOIN inventory_donation_types dt ON d.donation_type_id = dt.id
LEFT JOIN inventory_donors dr ON d.donor_id = dr.id
LEFT JOIN outbreaks o ON d.outbreak_id = o.id
LEFT JOIN treatment_sites ts ON d.treatment_site_id = ts.id
LEFT JOIN users u ON d.received_by = u.user_id
LEFT JOIN inventory_donation_items di ON d.id = di.donation_id
GROUP BY d.id, dt.name, dt.is_monetary, dr.name, dr.organization, dr.donor_type, 
         o.name, ts.name, u.user_name;

-- 12. Create function to update donation status
CREATE OR REPLACE FUNCTION update_donation_status()
RETURNS TRIGGER AS $$
BEGIN
    -- Update donation status based on received_date
    IF NEW.received_date IS NOT NULL AND OLD.received_date IS NULL THEN
        NEW.donation_status = 'received';
    END IF;
    
    -- Update updated_at timestamp
    NEW.updated_at = CURRENT_TIMESTAMP;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- 13. Create trigger for donation status updates
CREATE TRIGGER trigger_update_donation_status
    BEFORE UPDATE ON inventory_donations
    FOR EACH ROW
    EXECUTE FUNCTION update_donation_status();

-- 14. Create function to calculate donation value
CREATE OR REPLACE FUNCTION calculate_donation_value(donation_id INTEGER)
RETURNS DECIMAL AS $$
DECLARE
    total_value DECIMAL(15,2) := 0;
    monetary_value DECIMAL(15,2) := 0;
    item_value DECIMAL(15,2) := 0;
BEGIN
    -- Get monetary value
    SELECT COALESCE(monetary_value, 0) INTO monetary_value
    FROM inventory_donations
    WHERE id = donation_id;
    
    -- Get total estimated value of items
    SELECT COALESCE(SUM(estimated_value), 0) INTO item_value
    FROM inventory_donation_items
    WHERE donation_id = donation_id;
    
    total_value := monetary_value + item_value;
    
    RETURN total_value;
END;
$$ LANGUAGE plpgsql;

-- 15. Add comments for documentation
COMMENT ON TABLE inventory_donors IS 'Stores information about donors and organizations that provide donations';
COMMENT ON TABLE inventory_donation_types IS 'Defines different types of donations (monetary, in-kind, services, etc.)';
COMMENT ON TABLE inventory_donations IS 'Main table for tracking donations with donor and recipient information';
COMMENT ON TABLE inventory_donation_items IS 'Tracks individual items within in-kind donations';
COMMENT ON TABLE inventory_donation_acknowledgments IS 'Tracks acknowledgment letters and receipts sent to donors';
COMMENT ON COLUMN inventory_items.is_donated IS 'Indicates if this item was received as a donation';
COMMENT ON COLUMN inventory_items.donor_id IS 'References the donor who provided this item';
COMMENT ON COLUMN inventory_items.estimated_donor_value IS 'Estimated value provided by the donor';

-- 16. Create indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_donations_date_range ON inventory_donations(donation_date, received_date);
CREATE INDEX IF NOT EXISTS idx_donation_items_expiry ON inventory_donation_items(expiry_date);
CREATE INDEX IF NOT EXISTS idx_donors_type_status ON inventory_donors(donor_type, status);

-- 17. Add constraints for data integrity
ALTER TABLE inventory_donations ADD CONSTRAINT chk_donation_dates 
    CHECK (received_date IS NULL OR received_date >= donation_date);

ALTER TABLE inventory_donation_items ADD CONSTRAINT chk_quantity_positive 
    CHECK (quantity > 0);

ALTER TABLE inventory_donations ADD CONSTRAINT chk_monetary_value_positive 
    CHECK (monetary_value IS NULL OR monetary_value >= 0);

-- 18. Create view for expired donation items
CREATE OR REPLACE VIEW inventory_expired_donation_items AS
SELECT 
    di.id,
    di.donation_id,
    di.item_id,
    di.quantity,
    di.expiry_date,
    di.condition_status,
    i.name as item_name,
    d.donation_date,
    dr.name as donor_name,
    ts.name as treatment_site_name
FROM inventory_donation_items di
JOIN inventory_items i ON di.item_id = i.id
JOIN inventory_donations d ON di.donation_id = d.id
JOIN inventory_donors dr ON d.donor_id = dr.id
LEFT JOIN treatment_sites ts ON d.treatment_site_id = ts.id
WHERE di.expiry_date < CURRENT_DATE
AND di.condition_status != 'expired';

-- 19. Create function to get donation statistics
CREATE OR REPLACE FUNCTION get_donation_statistics(start_date DATE DEFAULT NULL, end_date DATE DEFAULT NULL)
RETURNS TABLE (
    total_donations INTEGER,
    total_monetary_value DECIMAL(15,2),
    total_in_kind_value DECIMAL(15,2),
    donor_count INTEGER,
    top_donor_name VARCHAR(255),
    top_donor_value DECIMAL(15,2)
) AS $$
BEGIN
    RETURN QUERY
    WITH donation_stats AS (
        SELECT 
            COUNT(DISTINCT d.id) as donations,
            COALESCE(SUM(d.monetary_value), 0) as monetary,
            COALESCE(SUM(di.estimated_value), 0) as in_kind,
            COUNT(DISTINCT d.donor_id) as donors
        FROM inventory_donations d
        LEFT JOIN inventory_donation_items di ON d.id = di.donation_id
        WHERE (start_date IS NULL OR d.donation_date >= start_date)
        AND (end_date IS NULL OR d.donation_date <= end_date)
    ),
    top_donor AS (
        SELECT 
            dr.name,
            COALESCE(SUM(d.monetary_value), 0) + COALESCE(SUM(di.estimated_value), 0) as total_value
        FROM inventory_donations d
        LEFT JOIN inventory_donation_items di ON d.id = di.donation_id
        JOIN inventory_donors dr ON d.donor_id = dr.id
        WHERE (start_date IS NULL OR d.donation_date >= start_date)
        AND (end_date IS NULL OR d.donation_date <= end_date)
        GROUP BY dr.id, dr.name
        ORDER BY total_value DESC
        LIMIT 1
    )
    SELECT 
        ds.donations,
        ds.monetary,
        ds.in_kind,
        ds.donors,
        td.name,
        td.total_value
    FROM donation_stats ds
    CROSS JOIN top_donor td;
END;
$$ LANGUAGE plpgsql; 