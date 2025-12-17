CREATE TABLE IF NOT EXISTS presigned_url_audit (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL,
    bucket_id uuid NOT NULL,
    file_id uuid NOT NULL,
    method text NOT NULL,
    expires_at timestamptz NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_presigned_audit_user ON presigned_url_audit (user_id);
CREATE INDEX IF NOT EXISTS idx_presigned_audit_bucket ON presigned_url_audit (bucket_id);
CREATE INDEX IF NOT EXISTS idx_presigned_audit_file ON presigned_url_audit (file_id);

