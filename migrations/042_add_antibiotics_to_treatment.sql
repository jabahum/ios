-- Add additional antibiotics to treatment table for outbreak 6 workflows
ALTER TABLE public.treatment
    ADD COLUMN IF NOT EXISTS ampicillin INTEGER,
    ADD COLUMN IF NOT EXISTS chloramphenicol INTEGER,
    ADD COLUMN IF NOT EXISTS amoxiclav INTEGER,
    ADD COLUMN IF NOT EXISTS azithromycin INTEGER,
    ADD COLUMN IF NOT EXISTS cefotaxime INTEGER,
    ADD COLUMN IF NOT EXISTS ceftazidime INTEGER,
    ADD COLUMN IF NOT EXISTS ciprofloxacin INTEGER,
    ADD COLUMN IF NOT EXISTS tetracycline INTEGER,
    ADD COLUMN IF NOT EXISTS imipenem INTEGER,
    ADD COLUMN IF NOT EXISTS cotrimoxazole INTEGER,
    ADD COLUMN IF NOT EXISTS gentamicin INTEGER,
    ADD COLUMN IF NOT EXISTS metronidazole INTEGER;

-- Optional: ensure route and frequency are VARCHAR to store standardized values
ALTER TABLE public.treatment
    ALTER COLUMN antibacterial_route TYPE VARCHAR(20),
    ALTER COLUMN antibacterial_freq TYPE VARCHAR(10); 