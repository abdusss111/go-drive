-- Rollback migration: Remove user quotas

ALTER TABLE users 
DROP COLUMN IF EXISTS quota_bytes,
DROP COLUMN IF EXISTS quota_files;

