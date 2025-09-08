-- Add afi_facility column to employee to support AFI-based filtering
ALTER TABLE public.employee
ADD COLUMN IF NOT EXISTS afi_facility TEXT;

-- Optional: simple index to speed up lookups by afi_facility
CREATE INDEX IF NOT EXISTS idx_employee_afi_facility ON public.employee (afi_facility);


