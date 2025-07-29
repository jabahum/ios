-- Comprehensive Inventory Management System for Treatment Sites
-- Migration 027: Create Inventory Management Tables

-- 1. Inventory Categories
CREATE TABLE inventory_categories (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    parent_category_id INTEGER REFERENCES inventory_categories(id),
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by INTEGER REFERENCES users(user_id),
    updated_by INTEGER REFERENCES users(user_id)
);

-- 2. Inventory Items
CREATE TABLE inventory_items (
    id SERIAL PRIMARY KEY,
    name VARCHAR(200) NOT NULL,
    description TEXT,
    category_id INTEGER REFERENCES inventory_categories(id),
    item_code VARCHAR(50) UNIQUE,
    barcode VARCHAR(100),
    unit_of_measure VARCHAR(20) NOT NULL, -- pieces, boxes, liters, etc.
    minimum_stock_level INTEGER DEFAULT 0,
    maximum_stock_level INTEGER,
    reorder_point INTEGER DEFAULT 0,
    unit_cost DECIMAL(10,2),
    is_active BOOLEAN DEFAULT true,
    is_critical BOOLEAN DEFAULT false, -- for critical medical supplies
    requires_cold_storage BOOLEAN DEFAULT false,
    expiry_tracking BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by INTEGER REFERENCES users(user_id),
    updated_by INTEGER REFERENCES users(user_id)
);

-- 3. Suppliers/Vendors
CREATE TABLE inventory_suppliers (
    id SERIAL PRIMARY KEY,
    name VARCHAR(200) NOT NULL,
    contact_person VARCHAR(100),
    email VARCHAR(100),
    phone VARCHAR(20),
    address TEXT,
    tax_id VARCHAR(50),
    payment_terms VARCHAR(100),
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by INTEGER REFERENCES users(user_id),
    updated_by INTEGER REFERENCES users(user_id)
);

-- 4. Treatment Sites/Facilities
CREATE TABLE treatment_sites (
    id SERIAL PRIMARY KEY,
    name VARCHAR(200) NOT NULL,
    facility_id INTEGER REFERENCES facilities(id),
    outbreak_id INTEGER REFERENCES outbreaks(id),
    site_type VARCHAR(50), -- ETU, CTC, Community Care Center, etc.
    location_address TEXT,
    contact_person VARCHAR(100),
    phone VARCHAR(20),
    email VARCHAR(100),
    capacity INTEGER, -- number of beds/patients
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by INTEGER REFERENCES users(user_id),
    updated_by INTEGER REFERENCES users(user_id)
);

-- 5. Inventory Stock Levels (by site)
CREATE TABLE inventory_stock_levels (
    id SERIAL PRIMARY KEY,
    item_id INTEGER REFERENCES inventory_items(id),
    site_id INTEGER REFERENCES treatment_sites(id),
    current_quantity INTEGER DEFAULT 0,
    reserved_quantity INTEGER DEFAULT 0, -- allocated but not yet issued
    available_quantity INTEGER DEFAULT 0, -- current_quantity - reserved_quantity
    last_updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_by INTEGER REFERENCES users(user_id)
);

-- 6. Inventory Transactions
CREATE TABLE inventory_transactions (
    id SERIAL PRIMARY KEY,
    transaction_type VARCHAR(20) NOT NULL, -- IN, OUT, ADJUSTMENT, TRANSFER
    item_id INTEGER REFERENCES inventory_items(id),
    from_site_id INTEGER REFERENCES treatment_sites(id),
    to_site_id INTEGER REFERENCES treatment_sites(id),
    quantity INTEGER NOT NULL,
    unit_cost DECIMAL(10,2),
    total_cost DECIMAL(10,2),
    reference_number VARCHAR(50), -- PO number, requisition number, etc.
    supplier_id INTEGER REFERENCES inventory_suppliers(id),
    transaction_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expiry_date DATE, -- for items with expiry tracking
    batch_number VARCHAR(50),
    notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by INTEGER REFERENCES users(user_id)
);

-- 7. Purchase Orders
CREATE TABLE inventory_purchase_orders (
    id SERIAL PRIMARY KEY,
    po_number VARCHAR(50) UNIQUE NOT NULL,
    supplier_id INTEGER REFERENCES inventory_suppliers(id),
    site_id INTEGER REFERENCES treatment_sites(id),
    outbreak_id INTEGER REFERENCES outbreaks(id),
    order_date DATE NOT NULL,
    expected_delivery_date DATE,
    status VARCHAR(20) DEFAULT 'DRAFT', -- DRAFT, SUBMITTED, APPROVED, RECEIVED, CANCELLED
    total_amount DECIMAL(12,2) DEFAULT 0,
    notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by INTEGER REFERENCES users(user_id),
    updated_by INTEGER REFERENCES users(user_id),
    approved_by INTEGER REFERENCES users(user_id),
    approved_at TIMESTAMP
);

-- 8. Purchase Order Items
CREATE TABLE inventory_po_items (
    id SERIAL PRIMARY KEY,
    po_id INTEGER REFERENCES inventory_purchase_orders(id),
    item_id INTEGER REFERENCES inventory_items(id),
    quantity_ordered INTEGER NOT NULL,
    quantity_received INTEGER DEFAULT 0,
    unit_cost DECIMAL(10,2),
    total_cost DECIMAL(10,2),
    notes TEXT
);

-- 9. Requisitions
CREATE TABLE inventory_requisitions (
    id SERIAL PRIMARY KEY,
    req_number VARCHAR(50) UNIQUE NOT NULL,
    requesting_site_id INTEGER REFERENCES treatment_sites(id),
    approving_site_id INTEGER REFERENCES treatment_sites(id), -- can be different from requesting
    outbreak_id INTEGER REFERENCES outbreaks(id),
    request_date DATE NOT NULL,
    required_date DATE,
    priority VARCHAR(20) DEFAULT 'NORMAL', -- LOW, NORMAL, HIGH, URGENT
    status VARCHAR(20) DEFAULT 'DRAFT', -- DRAFT, SUBMITTED, APPROVED, FULFILLED, CANCELLED
    total_amount DECIMAL(12,2) DEFAULT 0,
    notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by INTEGER REFERENCES users(user_id),
    updated_by INTEGER REFERENCES users(user_id),
    approved_by INTEGER REFERENCES users(user_id),
    approved_at TIMESTAMP
);

-- 10. Requisition Items
CREATE TABLE inventory_req_items (
    id SERIAL PRIMARY KEY,
    req_id INTEGER REFERENCES inventory_requisitions(id),
    item_id INTEGER REFERENCES inventory_items(id),
    quantity_requested INTEGER NOT NULL,
    quantity_approved INTEGER DEFAULT 0,
    quantity_issued INTEGER DEFAULT 0,
    unit_cost DECIMAL(10,2),
    total_cost DECIMAL(10,2),
    notes TEXT
);

-- 11. Inventory Alerts
CREATE TABLE inventory_alerts (
    id SERIAL PRIMARY KEY,
    alert_type VARCHAR(30) NOT NULL, -- LOW_STOCK, EXPIRY, OVERSTOCK, REORDER
    item_id INTEGER REFERENCES inventory_items(id),
    site_id INTEGER REFERENCES treatment_sites(id),
    alert_message TEXT NOT NULL,
    severity VARCHAR(20) DEFAULT 'MEDIUM', -- LOW, MEDIUM, HIGH, CRITICAL
    is_active BOOLEAN DEFAULT true,
    is_acknowledged BOOLEAN DEFAULT false,
    acknowledged_by INTEGER REFERENCES users(user_id),
    acknowledged_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 12. Inventory Reports
CREATE TABLE inventory_reports (
    id SERIAL PRIMARY KEY,
    report_type VARCHAR(50) NOT NULL, -- STOCK_LEVEL, TRANSACTION, CONSUMPTION, etc.
    site_id INTEGER REFERENCES treatment_sites(id),
    outbreak_id INTEGER REFERENCES outbreaks(id),
    report_date DATE NOT NULL,
    report_data JSONB, -- flexible storage for report data
    generated_by INTEGER REFERENCES users(user_id),
    generated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 13. Inventory Settings
CREATE TABLE inventory_settings (
    id SERIAL PRIMARY KEY,
    setting_key VARCHAR(100) UNIQUE NOT NULL,
    setting_value TEXT,
    setting_description TEXT,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_by INTEGER REFERENCES users(user_id)
);

-- Insert default inventory categories
INSERT INTO inventory_categories (name, description, created_by) VALUES
('Personal Protective Equipment (PPE)', 'All types of PPE including masks, gloves, gowns, etc.', 1),
('Medical Supplies', 'General medical supplies and equipment', 1),
('Medications', 'Pharmaceutical products and drugs', 1),
('Laboratory Supplies', 'Lab equipment, reagents, and testing materials', 1),
('Infection Control', 'Disinfectants, sanitizers, and cleaning supplies', 1),
('Emergency Response', 'Emergency and first aid supplies', 1),
('Nutrition', 'Food and nutritional supplements', 1),
('Logistics', 'Transportation and storage equipment', 1);

-- Insert default inventory settings
INSERT INTO inventory_settings (setting_key, setting_value, setting_description) VALUES
('low_stock_threshold', '20', 'Percentage threshold for low stock alerts'),
('expiry_alert_days', '30', 'Days before expiry to send alerts'),
('auto_reorder_enabled', 'false', 'Enable automatic reordering'),
('require_approval', 'true', 'Require approval for requisitions'),
('max_requisition_amount', '10000', 'Maximum amount for automatic approval'),
('stock_take_frequency', '7', 'Days between stock takes'),
('critical_item_threshold', '10', 'Minimum stock level for critical items');

-- Create indexes for better performance
CREATE INDEX idx_inventory_items_category ON inventory_items(category_id);
CREATE INDEX idx_inventory_items_active ON inventory_items(is_active);
CREATE INDEX idx_inventory_stock_levels_item ON inventory_stock_levels(item_id);
CREATE INDEX idx_inventory_stock_levels_site ON inventory_stock_levels(site_id);
CREATE INDEX idx_inventory_transactions_item ON inventory_transactions(item_id);
CREATE INDEX idx_inventory_transactions_date ON inventory_transactions(transaction_date);
CREATE INDEX idx_inventory_transactions_type ON inventory_transactions(transaction_type);
CREATE INDEX idx_inventory_alerts_active ON inventory_alerts(is_active);
CREATE INDEX idx_inventory_alerts_item ON inventory_alerts(item_id);
CREATE INDEX idx_inventory_alerts_site ON inventory_alerts(site_id);

-- Create triggers for automatic updates
CREATE OR REPLACE FUNCTION update_inventory_stock_levels()
RETURNS TRIGGER AS $$
BEGIN
    -- Update stock levels when transactions occur
    IF TG_OP = 'INSERT' THEN
        -- Update available quantity
        UPDATE inventory_stock_levels 
        SET current_quantity = current_quantity + NEW.quantity,
            available_quantity = available_quantity + NEW.quantity,
            last_updated = CURRENT_TIMESTAMP,
            updated_by = NEW.created_by
        WHERE item_id = NEW.item_id AND site_id = NEW.to_site_id;
        
        -- If it's a transfer, reduce from source site
        IF NEW.transaction_type = 'TRANSFER' AND NEW.from_site_id IS NOT NULL THEN
            UPDATE inventory_stock_levels 
            SET current_quantity = current_quantity - NEW.quantity,
                available_quantity = available_quantity - NEW.quantity,
                last_updated = CURRENT_TIMESTAMP,
                updated_by = NEW.created_by
            WHERE item_id = NEW.item_id AND site_id = NEW.from_site_id;
        END IF;
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_stock_levels
    AFTER INSERT ON inventory_transactions
    FOR EACH ROW
    EXECUTE FUNCTION update_inventory_stock_levels();

-- Function to check and create stock level records
CREATE OR REPLACE FUNCTION ensure_stock_level_record(item_id INTEGER, site_id INTEGER)
RETURNS VOID AS $$
BEGIN
    INSERT INTO inventory_stock_levels (item_id, site_id, current_quantity, available_quantity)
    VALUES (item_id, site_id, 0, 0)
    ON CONFLICT (item_id, site_id) DO NOTHING;
END;
$$ LANGUAGE plpgsql; 