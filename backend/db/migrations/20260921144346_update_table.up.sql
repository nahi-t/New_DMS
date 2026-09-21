-- UP
DROP TABLE IF EXISTS auditlog;

CREATE TABLE auditlog (
    id                  BIGSERIAL    PRIMARY KEY,
    user_id             BIGINT       NOT NULL REFERENCES users(id) ON DELETE SET NULL,
    user_name           VARCHAR(255) NOT NULL,
    event               VARCHAR(255) NOT NULL,
    event_happened_time TIMESTAMPTZ  NOT NULL DEFAULT now(),
    created_at          TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE INDEX idx_auditlog_user_id
    ON auditlog (user_id);

CREATE INDEX idx_auditlog_event_time
    ON auditlog (event_happened_time DESC);

CREATE INDEX idx_auditlog_event
    ON auditlog (event);

CREATE INDEX idx_auditlog_user_time
    ON auditlog (user_id, event_happened_time DESC);