-- Migration: Add user quotas
-- Description: Adds quota_bytes and quota_files columns to users table
-- Default values: 1GB (1073741824 bytes) and 1000 files per user

ALTER TABLE users 
ADD COLUMN quota_bytes BIGINT NOT NULL DEFAULT 1073741824, -- 1GB default
ADD COLUMN quota_files INTEGER NOT NULL DEFAULT 1000;       -- 1000 files default

-- Add comments for documentation
COMMENT ON COLUMN users.quota_bytes IS 'Maximum storage quota in bytes for the user (default: 1GB)';
COMMENT ON COLUMN users.quota_files IS 'Maximum number of files allowed for the user (default: 1000)';

