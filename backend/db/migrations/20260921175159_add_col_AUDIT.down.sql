-- DOWN
DROP INDEX IF EXISTS idx_auditlog_ip;
DROP INDEX IF EXISTS idx_auditlog_success;

ALTER TABLE auditlog
    DROP COLUMN IF EXISTS ip_address,
    DROP COLUMN IF EXISTS user_agent,
    DROP COLUMN IF EXISTS success;