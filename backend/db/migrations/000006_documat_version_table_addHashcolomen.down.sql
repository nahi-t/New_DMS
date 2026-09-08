DROP INDEX IF EXISTS idx_document_versions_hashed_string;
ALTER TABLE document_versions DROP COLUMN IF EXISTS hashed_string;