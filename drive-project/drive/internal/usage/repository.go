package usage

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const repositoryTimeout = 5 * time.Second

// Repository provides access to usage and quota data.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository constructs a usage repository.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// GetUserUsage returns the total usage (bytes and files) for a user across all buckets.
func (r *Repository) GetUserUsage(ctx context.Context, userID uuid.UUID) (totalBytes, totalFiles int64, err error) {
	ctx, cancel := context.WithTimeout(ctx, repositoryTimeout)
	defer cancel()

	query := `
	SELECT COALESCE(SUM(u.total_bytes), 0) AS total_bytes,
	       COALESCE(SUM(u.file_count), 0) AS file_count
	FROM buckets b
	LEFT JOIN bucket_usage u ON u.bucket_id = b.id
	WHERE b.owner_id = $1;`

	err = r.pool.QueryRow(ctx, query, userID).Scan(&totalBytes, &totalFiles)
	if err != nil {
		return 0, 0, fmt.Errorf("get user usage: %w", err)
	}

	return totalBytes, totalFiles, nil
}

// GetUserQuota returns the quota limits (bytes and files) for a user.
func (r *Repository) GetUserQuota(ctx context.Context, userID uuid.UUID) (quotaBytes, quotaFiles int64, err error) {
	ctx, cancel := context.WithTimeout(ctx, repositoryTimeout)
	defer cancel()

	query := `SELECT quota_bytes, quota_files FROM users WHERE id = $1;`

	err = r.pool.QueryRow(ctx, query, userID).Scan(&quotaBytes, &quotaFiles)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, 0, fmt.Errorf("user not found: %w", err)
		}
		return 0, 0, fmt.Errorf("get user quota: %w", err)
	}

	return quotaBytes, quotaFiles, nil
}

// GetBucketUsage returns the usage statistics for a specific bucket.
func (r *Repository) GetBucketUsage(ctx context.Context, bucketID uuid.UUID) (totalBytes, totalFiles int64, err error) {
	ctx, cancel := context.WithTimeout(ctx, repositoryTimeout)
	defer cancel()

	query := `
	SELECT COALESCE(total_bytes, 0) AS total_bytes,
	       COALESCE(file_count, 0) AS file_count
	FROM bucket_usage
	WHERE bucket_id = $1;`

	err = r.pool.QueryRow(ctx, query, bucketID).Scan(&totalBytes, &totalFiles)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// Bucket exists but has no usage record yet
			return 0, 0, nil
		}
		return 0, 0, fmt.Errorf("get bucket usage: %w", err)
	}

	return totalBytes, totalFiles, nil
}

