package presigned

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const auditTimeout = 3 * time.Second

// AuditRepository stores presigned URL creation events.
type AuditRepository struct {
	pool *pgxpool.Pool
}

// NewRepository constructs a new audit repository.
func NewRepository(pool *pgxpool.Pool) *AuditRepository {
	return &AuditRepository{pool: pool}
}

// Log writes an audit record for a presigned URL generation.
func (r *AuditRepository) Log(ctx context.Context, entry AuditEntry) error {
	ctx, cancel := context.WithTimeout(ctx, auditTimeout)
	defer cancel()

	_, err := r.pool.Exec(ctx, `
INSERT INTO presigned_url_audit (user_id, bucket_id, file_id, method, expires_at)
VALUES ($1, $2, $3, $4, $5);
`, entry.UserID, entry.BucketID, entry.FileID, entry.Method, entry.ExpiresAt)
	if err != nil {
		return fmt.Errorf("insert presigned audit: %w", err)
	}
	return nil
}
