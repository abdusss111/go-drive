package bucket

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const repositoryTimeout = 5 * time.Second

// Repository allows access to bucket persistence.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository constructs a bucket repository.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// Create inserts a new bucket for the owner.
func (r *Repository) Create(ctx context.Context, ownerID uuid.UUID, name string, description *string) (Bucket, error) {
	ctx, cancel := context.WithTimeout(ctx, repositoryTimeout)
	defer cancel()

	name = strings.TrimSpace(name)
	bucketID := uuid.New()

	query := `
INSERT INTO buckets (id, owner_id, name, description)
VALUES ($1, $2, $3, $4)
RETURNING id, owner_id, name, description, created_at, updated_at;`

	row := r.pool.QueryRow(ctx, query, bucketID, ownerID, name, description)

	var bucket Bucket
	if err := row.Scan(&bucket.ID, &bucket.OwnerID, &bucket.Name, &bucket.Description, &bucket.CreatedAt, &bucket.UpdatedAt); err != nil {
		if isUniqueViolation(err) {
			return Bucket{}, ErrBucketNameExists
		}
		return Bucket{}, fmt.Errorf("create bucket: %w", err)
	}

	if err := r.ensureUsageRow(ctx, bucket.ID); err != nil {
		return Bucket{}, err
	}

	return bucket, nil
}

// List returns all buckets owned by the user.
func (r *Repository) List(ctx context.Context, ownerID uuid.UUID) ([]Bucket, error) {
	ctx, cancel := context.WithTimeout(ctx, repositoryTimeout)
	defer cancel()

	query := `
SELECT b.id,
       b.owner_id,
       b.name,
       b.description,
       b.created_at,
       b.updated_at,
       COALESCE(u.total_bytes, 0) AS total_bytes,
       COALESCE(u.file_count, 0) AS file_count
FROM buckets b
LEFT JOIN bucket_usage u ON u.bucket_id = b.id
WHERE b.owner_id = $1
ORDER BY b.created_at DESC;`

	rows, err := r.pool.Query(ctx, query, ownerID)
	if err != nil {
		return nil, fmt.Errorf("list buckets: %w", err)
	}
	defer rows.Close()

	var buckets []Bucket
	for rows.Next() {
		var bucket Bucket
		if err := rows.Scan(&bucket.ID, &bucket.OwnerID, &bucket.Name, &bucket.Description, &bucket.CreatedAt, &bucket.UpdatedAt, &bucket.Usage.TotalBytes, &bucket.Usage.FileCount); err != nil {
			return nil, fmt.Errorf("scan bucket: %w", err)
		}
		buckets = append(buckets, bucket)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate buckets: %w", err)
	}
	return buckets, nil
}

// Get fetches a single bucket ensuring ownership.
func (r *Repository) Get(ctx context.Context, ownerID, bucketID uuid.UUID) (Bucket, error) {
	ctx, cancel := context.WithTimeout(ctx, repositoryTimeout)
	defer cancel()

	query := `
SELECT b.id,
       b.owner_id,
       b.name,
       b.description,
       b.created_at,
       b.updated_at,
       COALESCE(u.total_bytes, 0) AS total_bytes,
       COALESCE(u.file_count, 0) AS file_count
FROM buckets b
LEFT JOIN bucket_usage u ON u.bucket_id = b.id
WHERE b.id = $1 AND b.owner_id = $2;`

	var bucket Bucket
	err := r.pool.QueryRow(ctx, query, bucketID, ownerID).Scan(
		&bucket.ID,
		&bucket.OwnerID,
		&bucket.Name,
		&bucket.Description,
		&bucket.CreatedAt,
		&bucket.UpdatedAt,
		&bucket.Usage.TotalBytes,
		&bucket.Usage.FileCount,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Bucket{}, ErrBucketNotFound
		}
		return Bucket{}, fmt.Errorf("get bucket: %w", err)
	}

	return bucket, nil
}

// Delete removes a bucket owned by the user.
func (r *Repository) Delete(ctx context.Context, ownerID, bucketID uuid.UUID) error {
	ctx, cancel := context.WithTimeout(ctx, repositoryTimeout)
	defer cancel()

	commandTag, err := r.pool.Exec(ctx, `DELETE FROM buckets WHERE id = $1 AND owner_id = $2;`, bucketID, ownerID)
	if err != nil {
		return fmt.Errorf("delete bucket: %w", err)
	}
	if commandTag.RowsAffected() == 0 {
		return ErrBucketNotFound
	}
	return nil
}

// UpdateUsage increments or decrements usage statistics.
func (r *Repository) UpdateUsage(ctx context.Context, bucketID uuid.UUID, deltaBytes int64, deltaFiles int64) error {
	ctx, cancel := context.WithTimeout(ctx, repositoryTimeout)
	defer cancel()

	query := `
INSERT INTO bucket_usage (bucket_id, total_bytes, file_count, updated_at)
VALUES ($1, $2, $3, NOW())
ON CONFLICT (bucket_id)
DO UPDATE SET
    total_bytes = GREATEST(bucket_usage.total_bytes + EXCLUDED.total_bytes, 0),
    file_count  = GREATEST(bucket_usage.file_count + EXCLUDED.file_count, 0),
    updated_at  = NOW();`

	if _, err := r.pool.Exec(ctx, query, bucketID, deltaBytes, deltaFiles); err != nil {
		return fmt.Errorf("update usage: %w", err)
	}
	return nil
}

// RecordUsageSnapshot inserts an aggregate usage snapshot for the owner.
func (r *Repository) RecordUsageSnapshot(ctx context.Context, ownerID uuid.UUID) error {
	ctx, cancel := context.WithTimeout(ctx, repositoryTimeout)
	defer cancel()

	query := `
WITH stats AS (
    SELECT COALESCE(SUM(u.total_bytes), 0) AS total_bytes,
           COALESCE(SUM(u.file_count), 0) AS file_count
    FROM buckets b
    LEFT JOIN bucket_usage u ON u.bucket_id = b.id
    WHERE b.owner_id = $1
)
INSERT INTO usage_snapshots (user_id, total_bytes, file_count)
SELECT $1, stats.total_bytes, stats.file_count FROM stats;`

	if _, err := r.pool.Exec(ctx, query, ownerID); err != nil {
		return fmt.Errorf("record usage snapshot: %w", err)
	}
	return nil
}

func (r *Repository) ensureUsageRow(ctx context.Context, bucketID uuid.UUID) error {
	if _, err := r.pool.Exec(ctx, `
INSERT INTO bucket_usage (bucket_id, total_bytes, file_count)
VALUES ($1, 0, 0)
ON CONFLICT (bucket_id) DO NOTHING;
`, bucketID); err != nil {
		return fmt.Errorf("ensure usage row: %w", err)
	}
	return nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}
