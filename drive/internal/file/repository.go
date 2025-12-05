package file

import (
	"context"
	"fmt"
	"time"

	"github.com/abduss/godrive/internal/bucket"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const repoTimeout = 5 * time.Second

// Repository provides access to file metadata storage.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository builds a new file repository.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// Create inserts metadata for a new file.
func (r *Repository) Create(ctx context.Context, meta Metadata) (Metadata, error) {
	ctx, cancel := context.WithTimeout(ctx, repoTimeout)
	defer cancel()

	query := `
INSERT INTO files (id, bucket_id, object_name, original_filename, size_bytes, content_type, checksum, metadata)
VALUES ($1, $2, $3, $4, $5, $6, $7, NULL)
RETURNING id, bucket_id, object_name, original_filename, size_bytes, content_type, checksum, created_at, updated_at;`

	row := r.pool.QueryRow(ctx, query,
		meta.ID,
		meta.BucketID,
		meta.ObjectName,
		meta.OriginalFilename,
		meta.SizeBytes,
		meta.ContentType,
		meta.Checksum,
	)

	var stored Metadata
	if err := row.Scan(&stored.ID, &stored.BucketID, &stored.ObjectName, &stored.OriginalFilename, &stored.SizeBytes, &stored.ContentType, &stored.Checksum, &stored.CreatedAt, &stored.UpdatedAt); err != nil {
		return Metadata{}, fmt.Errorf("create file metadata: %w", err)
	}
	return stored, nil
}

// List returns files owned by the user in a bucket.
func (r *Repository) List(ctx context.Context, ownerID, bucketID uuid.UUID) ([]Metadata, error) {
	ctx, cancel := context.WithTimeout(ctx, repoTimeout)
	defer cancel()

	query := `
SELECT f.id, f.bucket_id, f.object_name, f.original_filename, f.size_bytes, f.content_type, f.checksum, f.created_at, f.updated_at
FROM files f
JOIN buckets b ON b.id = f.bucket_id
WHERE f.bucket_id = $1 AND b.owner_id = $2
ORDER BY f.created_at DESC;`

	rows, err := r.pool.Query(ctx, query, bucketID, ownerID)
	if err != nil {
		return nil, fmt.Errorf("list files: %w", err)
	}
	defer rows.Close()

	var files []Metadata
	for rows.Next() {
		var meta Metadata
		if err := rows.Scan(&meta.ID, &meta.BucketID, &meta.ObjectName, &meta.OriginalFilename, &meta.SizeBytes, &meta.ContentType, &meta.Checksum, &meta.CreatedAt, &meta.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan file metadata: %w", err)
		}
		files = append(files, meta)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate files: %w", err)
	}
	return files, nil
}

// Get fetches metadata for a single file ensuring ownership.
func (r *Repository) Get(ctx context.Context, ownerID, bucketID, fileID uuid.UUID) (Metadata, error) {
	ctx, cancel := context.WithTimeout(ctx, repoTimeout)
	defer cancel()

	query := `
SELECT f.id, f.bucket_id, f.object_name, f.original_filename, f.size_bytes, f.content_type, f.checksum, f.created_at, f.updated_at
FROM files f
JOIN buckets b ON b.id = f.bucket_id
WHERE f.id = $1 AND f.bucket_id = $2 AND b.owner_id = $3;`

	var meta Metadata
	err := r.pool.QueryRow(ctx, query, fileID, bucketID, ownerID).Scan(
		&meta.ID,
		&meta.BucketID,
		&meta.ObjectName,
		&meta.OriginalFilename,
		&meta.SizeBytes,
		&meta.ContentType,
		&meta.Checksum,
		&meta.CreatedAt,
		&meta.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return Metadata{}, ErrFileNotFound
		}
		return Metadata{}, fmt.Errorf("get file metadata: %w", err)
	}
	return meta, nil
}

// Delete removes metadata and returns the deleted record.
func (r *Repository) Delete(ctx context.Context, ownerID, bucketID, fileID uuid.UUID) (Metadata, error) {
	ctx, cancel := context.WithTimeout(ctx, repoTimeout)
	defer cancel()

	query := `
DELETE FROM files f
USING buckets b
WHERE f.id = $1
  AND f.bucket_id = $2
  AND b.id = f.bucket_id
  AND b.owner_id = $3
RETURNING f.id, f.bucket_id, f.object_name, f.original_filename, f.size_bytes, f.content_type, f.checksum, f.created_at, f.updated_at;`

	var meta Metadata
	err := r.pool.QueryRow(ctx, query, fileID, bucketID, ownerID).Scan(
		&meta.ID,
		&meta.BucketID,
		&meta.ObjectName,
		&meta.OriginalFilename,
		&meta.SizeBytes,
		&meta.ContentType,
		&meta.Checksum,
		&meta.CreatedAt,
		&meta.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return Metadata{}, ErrFileNotFound
		}
		return Metadata{}, fmt.Errorf("delete file metadata: %w", err)
	}
	return meta, nil
}

// ListObjectsForBucket returns object names for external cleanup.
func (r *Repository) ListObjectsForBucket(ctx context.Context, bucketID uuid.UUID) ([]bucket.FileObject, error) {
	ctx, cancel := context.WithTimeout(ctx, repoTimeout)
	defer cancel()

	query := `SELECT object_name, size_bytes FROM files WHERE bucket_id = $1;`

	rows, err := r.pool.Query(ctx, query, bucketID)
	if err != nil {
		return nil, fmt.Errorf("list objects for bucket: %w", err)
	}
	defer rows.Close()

	var objects []bucket.FileObject
	for rows.Next() {
		var obj bucket.FileObject
		if err := rows.Scan(&obj.ObjectName, &obj.SizeBytes); err != nil {
			return nil, fmt.Errorf("scan object name: %w", err)
		}
		objects = append(objects, obj)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate object names: %w", err)
	}
	return objects, nil
}
