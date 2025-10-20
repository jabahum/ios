-- Insert missing inventory suppliers
INSERT INTO inventory_suppliers (name, contact_person, email, phone, address, created_by) VALUES
('Ministry of Health Uganda', 'Dr. Jane Smith', 'procurement@health.go.ug', '+256-XXX-XXXX', 'Kampala, Uganda', 1),
('World Health Organization', 'Dr. John Doe', 'supplies@who.int', '+41-XXX-XXXX', 'Geneva, Switzerland', 1),
('UNICEF Uganda', 'Ms. Sarah Johnson', 'supplies@unicef.org', '+256-XXX-XXXX', 'Kampala, Uganda', 1),
('Medecins Sans Frontieres', 'Dr. Pierre Martin', 'logistics@msf.org', '+33-XXX-XXXX', 'Paris, France', 1),
('Red Cross Uganda', 'Mr. David Wilson', 'supplies@redcross.ug', '+256-XXX-XXXX', 'Kampala, Uganda', 1),
('Local Medical Suppliers', 'Mr. Ahmed Hassan', 'info@localmedical.ug', '+256-XXX-XXXX', 'Kampala, Uganda', 1),
('International Medical Corps', 'Dr. Maria Garcia', 'supplies@internationalmedicalcorps.org', '+1-XXX-XXXX', 'Los Angeles, USA', 1),
('Partners In Health', 'Dr. Paul Farmer', 'supplies@pih.org', '+1-XXX-XXXX', 'Boston, USA', 1);

-- Insert missing treatment sites
INSERT INTO treatment_sites (name, site_type, contact_person, phone, email, created_by) VALUES
('Entebbe Regional Referral Hospital', 'Hospital', 'Dr. James Okello', '+256-XXX-XXXX', 'entebbe@health.go.ug', 1),
('Mbarara Regional Referral Hospital', 'Hospital', 'Dr. Grace Nakato', '+256-XXX-XXXX', 'mbarara@health.go.ug', 1),
('Gulu Regional Referral Hospital', 'Hospital', 'Dr. Peter Ochieng', '+256-XXX-XXXX', 'gulu@health.go.ug', 1),
('Ebola Treatment Unit - Bwera', 'ETU', 'Dr. Mary Akello', '+256-XXX-XXXX', 'bwera@health.go.ug', 1),
('Ebola Treatment Unit - Mubende', 'ETU', 'Dr. John Kato', '+256-XXX-XXXX', 'mubende@health.go.ug', 1),
('Community Health Center - Kasese', 'Health Center', 'Ms. Rose Namukasa', '+256-XXX-XXXX', 'kasese@health.go.ug', 1),
('Community Health Center - Bundibugyo', 'Health Center', 'Mr. Paul Musoke', '+256-XXX-XXXX', 'bundibugyo@health.go.ug', 1);

-- Insert some sample inventory items
INSERT INTO inventory_items (name, description, category_id, item_code, unit_of_measure, minimum_stock_level, reorder_point, unit_cost, created_by) VALUES
('N95 Mask', 'Respiratory protective mask', (SELECT id FROM inventory_categories WHERE name = 'Personal Protective Equipment (PPE)' LIMIT 1), 'PPE001', 'pieces', 50, 20, 2.50, 1),
('Sterile Gloves (Box of 100)', 'Disposable sterile gloves', (SELECT id FROM inventory_categories WHERE name = 'Personal Protective Equipment (PPE)' LIMIT 1), 'PPE002', 'boxes', 30, 10, 15.00, 1),
('Paracetamol 500mg (Box of 1000)', 'Pain reliever and fever reducer', (SELECT id FROM inventory_categories WHERE name = 'Medications' LIMIT 1), 'MED001', 'boxes', 100, 50, 25.00, 1),
('Syringe 5ml', 'Disposable syringe', (SELECT id FROM inventory_categories WHERE name = 'Medical Supplies' LIMIT 1), 'MEDS001', 'pieces', 200, 100, 0.50, 1),
('COVID-19 Antigen Rapid Test Kit', 'Rapid diagnostic test for COVID-19', (SELECT id FROM inventory_categories WHERE name = 'Laboratory Supplies' LIMIT 1), 'LAB001', 'kits', 20, 10, 10.00, 1);

SELECT 'Missing data inserted successfully!' as message;
