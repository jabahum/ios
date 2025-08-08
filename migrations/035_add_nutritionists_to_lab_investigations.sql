-- Migration to add nutritionists column to mpox_laboratory_investigations table

-- Add nutritionists column to laboratory investigations
ALTER TABLE mpox_laboratory_investigations ADD COLUMN IF NOT EXISTS nutritionists FLOAT; 