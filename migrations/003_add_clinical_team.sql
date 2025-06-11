-- Add clinical_team column to encounter table
ALTER TABLE public.encounter ADD COLUMN clinical_team VARCHAR(255);

-- Add comment to the column
COMMENT ON COLUMN public.encounter.clinical_team IS 'Other clinical team members involved in the encounter'; 