CREATE TABLE document_shares (
    id            BIGSERIAL PRIMARY KEY,
    document_id   BIGINT NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    user_id       BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    permission    VARCHAR(20) NOT NULL DEFAULT 'viewer',  -- 'viewer' or 'editor'
    shared_by     BIGINT NOT NULL REFERENCES users(id),
    shared_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at    TIMESTAMPTZ,  -- optional: auto-expire the share

    UNIQUE (document_id, user_id)  -- one share per user per document
);

CREATE INDEX idx_shares_document ON document_shares (document_id);
CREATE INDEX idx_shares_user ON document_shares (user_id);
CREATE INDEX idx_shares_permission ON document_shares (document_id, permission);