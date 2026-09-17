-- ============================================================================
-- Roll back to the original schema
-- ============================================================================

-- 1. Drop the UUID id, restore BIGSERIAL
DO $$
DECLARE
    pk_name TEXT;
BEGIN
    SELECT conname INTO pk_name
    FROM pg_constraint
    WHERE conrelid = 'document_versions'::regclass
      AND contype  = 'p';

    IF pk_name IS NOT NULL THEN
        EXECUTE format('ALTER TABLE document_versions DROP CONSTRAINT %I', pk_name);
    END IF;
END $$;

ALTER TABLE document_versions DROP COLUMN id;
ALTER TABLE document_versions ADD COLUMN id BIGSERIAL PRIMARY KEY;

-- 2. Drop the new columns
ALTER TABLE document_versions DROP COLUMN IF EXISTS original_filename;
ALTER TABLE document_versions DROP COLUMN IF EXISTS hashed_string;

-- 3. Rename back
ALTER TABLE document_versions RENAME COLUMN public_id  TO file_path;
ALTER TABLE document_versions RENAME COLUMN created_at TO time;

-- 4. Convert version back to VARCHAR(20)
ALTER TABLE document_versions
    ALTER COLUMN version TYPE VARCHAR(20) USING version::TEXT;

-- 5. Restore original indexes
DROP INDEX IF EXISTS idx_document_versions_user_id;
DROP INDEX IF EXISTS idx_document_versions_document_id;

CREATE INDEX IF NOT EXISTS idx_document_versions_user_id
    ON document_versions (user_id);

CREATE INDEX IF NOT EXISTS idx_document_versions_document_id
    ON document_versions (document_id);