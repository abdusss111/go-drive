-- Add version column to files table
ALTER TABLE files ADD COLUMN IF NOT EXISTS version INTEGER NOT NULL DEFAULT 1;

-- Remove unique constraint on (bucket_id, object_name) to allow multiple versions
ALTER TABLE files DROP CONSTRAINT IF EXISTS files_bucket_id_object_name_key;

-- Add unique constraint on (bucket_id, object_name, version) to ensure version uniqueness
ALTER TABLE files ADD CONSTRAINT files_bucket_id_object_name_version_key 
    UNIQUE (bucket_id, object_name, version);

-- Create index for faster version queries
CREATE INDEX IF NOT EXISTS idx_files_bucket_object ON files (bucket_id, object_name, version DESC);

-- Add is_current flag to mark the current/latest version
ALTER TABLE files ADD COLUMN IF NOT EXISTS is_current BOOLEAN NOT NULL DEFAULT TRUE;

-- Create index for current version lookups
CREATE INDEX IF NOT EXISTS idx_files_current_version ON files (bucket_id, object_name) 
    WHERE is_current = TRUE;

