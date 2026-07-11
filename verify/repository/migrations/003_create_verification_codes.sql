CREATE TABLE IF NOT EXISTS verification_codes (
    id              UUID PRIMARY KEY DEFAULT uuidv7(),
    contact         VARCHAR(254) NOT NULL,
    contact_type    VARCHAR(10) NOT NULL,
    code            VARCHAR(20) NOT NULL,
    name            VARCHAR(100) NOT NULL,
    password_hash   VARCHAR(255),
    expires_at      TIMESTAMPTZ NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    used_at         TIMESTAMPTZ,
    user_id         UUID REFERENCES users(id) ON DELETE SET NULL
);
CREATE INDEX IF NOT EXISTS idx_vc_code ON verification_codes(code) WHERE used_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_vc_contact ON verification_codes(contact, contact_type);
