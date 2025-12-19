-- Remove indexes
DROP INDEX IF EXISTS idx_files_current_version;
DROP INDEX IF EXISTS idx_files_bucket_object;

-- Remove is_current column
ALTER TABLE files DROP COLUMN IF EXISTS is_current;

-- Remove version column
ALTER TABLE files DROP COLUMN IF EXISTS version;

-- Restore original unique constraint
ALTER TABLE files DROP CONSTRAINT IF EXISTS files_bucket_id_object_name_version_key;
ALTER TABLE files ADD CONSTRAINT files_bucket_id_object_name_key UNIQUE (bucket_id, object_name);

