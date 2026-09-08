ALTER TABLE document_versions 
ADD COLUMN IF NOT EXISTS hashed_string VARCHAR(64) NOT NULL DEFAULT '';

-- 3. Create an index on hashed_string for fast lookup and duplicate content checks
CREATE INDEX IF NOT EXISTS idx_document_versions_hashed_string ON document_versions(hashed_string);