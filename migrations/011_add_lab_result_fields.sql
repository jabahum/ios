-- Add new fields to vhf_laboratory table
ALTER TABLE vhf_laboratory
ADD COLUMN IF NOT EXISTS test_result VARCHAR(255),
ADD COLUMN IF NOT EXISTS date_tested TIMESTAMP,
ADD COLUMN IF NOT EXISTS lab_name VARCHAR(255); 