-- DOWN
DROP INDEX IF EXISTS idx_auditlog_event;
DROP INDEX IF EXISTS idx_auditlog_event_time;
DROP INDEX IF EXISTS idx_auditlog_user_id;

DROP TABLE IF EXISTS auditlog;