CREATE TABLE IF NOT EXISTS verification_attempts (
    contact       VARCHAR(254) NOT NULL,
    contact_type  VARCHAR(10)  NOT NULL,
    attempts      INT          NOT NULL DEFAULT 1,
    last_attempt  TIMESTAMPTZ  NOT NULL DEFAULT now(),
    blocked_until TIMESTAMPTZ,
    PRIMARY KEY (contact, contact_type)
);

CREATE INDEX IF NOT EXISTS idx_va_blocked
    ON verification_attempts(contact, contact_type, blocked_until);
