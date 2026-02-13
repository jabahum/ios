-- Add HBC (home-based care) columns to clients table if they don't exist
-- Fixes: pq: column "hbc_followup" does not exist when loading case list

ALTER TABLE public.clients
ADD COLUMN IF NOT EXISTS hbc_followup VARCHAR(50);

ALTER TABLE public.clients
ADD COLUMN IF NOT EXISTS hbc_phone VARCHAR(50);

ALTER TABLE public.clients
ADD COLUMN IF NOT EXISTS hbc_language INTEGER;

COMMENT ON COLUMN public.clients.hbc_followup IS 'Home-based care follow-up type (e.g. ivr)';
COMMENT ON COLUMN public.clients.hbc_phone IS 'Home-based care contact phone';
COMMENT ON COLUMN public.clients.hbc_language IS 'Home-based care preferred language';
