-- UP
ALTER TABLE auditlog
    ADD COLUMN IF NOT EXISTS ip_address INET,
    ADD COLUMN IF NOT EXISTS user_agent TEXT,
    ADD COLUMN IF NOT EXISTS success    BOOLEAN NOT NULL DEFAULT TRUE;

CREATE INDEX IF NOT EXISTS idx_auditlog_success
    ON auditlog (success)
    WHERE success = FALSE;

CREATE INDEX IF NOT EXISTS idx_auditlog_ip
    ON auditlog (ip_address);