ALTER TABLE verification_codes
ADD COLUMN IF NOT EXISTS is_phantom BOOLEAN NOT NULL DEFAULT FALSE;

CREATE INDEX IF NOT EXISTS idx_vc_phantom
ON verification_codes(is_phantom) WHERE used_at IS NULL;
