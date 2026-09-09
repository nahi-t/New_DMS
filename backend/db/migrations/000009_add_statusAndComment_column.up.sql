ALTER TABLE documents
    ADD COLUMN status VARCHAR(50) NOT NULL DEFAULT 'waiting for approval',
    ADD COLUMN comment TEXT;