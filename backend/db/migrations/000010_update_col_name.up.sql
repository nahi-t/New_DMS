-- ============================================================================
-- Migrate document_versions from local-disk schema to Cloudinary schema
-- ============================================================================

-- 1. Enable pgcrypto so gen_random_uuid() is available
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- 2. Rename existing columns to their new names
ALTER TABLE document_versions RENAME COLUMN file_path TO public_id;
ALTER TABLE document_versions RENAME COLUMN time      TO created_at;

-- 3. Convert `version` from VARCHAR(20) to INTEGER
--    USING casts each text value to int; will error if any row has non-numeric text.
ALTER TABLE document_versions
    ALTER COLUMN version TYPE INTEGER USING version::INTEGER;

-- 4. Add `original_filename` (nullable — old rows won't have it)
ALTER TABLE document_versions
    ADD COLUMN IF NOT EXISTS original_filename TEXT;

-- 5. Add `hashed_string` as nullable first, so the ALTER succeeds on existing rows
ALTER TABLE document_versions
    ADD COLUMN IF NOT EXISTS hashed_string TEXT;

-- Backfill any existing rows with a placeholder hash.
-- ⚠️  These placeholders will never match a real SHA-256 hash, so the next
--     upload to any pre-existing document will create a new version.
UPDATE document_versions SET hashed_string = '' WHERE hashed_string IS NULL;

-- 6. Now enforce NOT NULL
ALTER TABLE document_versions ALTER COLUMN hashed_string SET NOT NULL;

-- ============================================================================
-- 7. Convert `id` from BIGSERIAL to UUID
--    This is the risky part — any FK pointing at document_versions.id will
--    break. If nothing references it, we're safe.
-- ============================================================================

-- 7a. Drop the PK constraint (its name may vary; we detect it dynamically)
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

-- 7b. Drop the old sequence default and the BIGSERIAL column
ALTER TABLE document_versions DROP COLUMN id;

-- 7c. Add the new UUID column
ALTER TABLE document_versions
    ADD COLUMN id UUID NOT NULL DEFAULT gen_random_uuid();

-- 7d. Re-add the primary key
ALTER TABLE document_versions ADD CONSTRAINT document_versions_pkey PRIMARY KEY (id);

-- ============================================================================
-- 8. Recreate indexes with the new names
-- ============================================================================
DROP INDEX IF EXISTS idx_document_versions_user_id;
DROP INDEX IF EXISTS idx_document_versions_document_id;

CREATE INDEX IF NOT EXISTS idx_document_versions_user_id
    ON document_versions (user_id);

CREATE INDEX IF NOT EXISTS idx_document_versions_document_id
    ON document_versions (document_id, version DESC);