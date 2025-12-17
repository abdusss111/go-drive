package bucket

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
)

// FileObject represents the minimal metadata required to manage objects in storage.
type FileObject struct {
	ObjectName string
	SizeBytes  int64
}

// FileIndex defines the contract used to inspect files belonging to a bucket.
type FileIndex interface {
	ListObjectsForBucket(ctx context.Context, bucketID uuid.UUID) ([]FileObject, error)
}

type repository interface {
	Create(ctx context.Context, ownerID uuid.UUID, name string, description *string) (Bucket, error)
	List(ctx context.Context, ownerID uuid.UUID) ([]Bucket, error)
	Get(ctx context.Context, ownerID, bucketID uuid.UUID) (Bucket, error)
	Delete(ctx context.Context, ownerID, bucketID uuid.UUID) error
	RecordUsageSnapshot(ctx context.Context, ownerID uuid.UUID) error
}

// Service orchestrates bucket operations.
type Service struct {
	repo         repository
	files        FileIndex
	objectStore  *minio.Client
	objectBucket string
}

// NewService constructs a bucket service.
func NewService(repo repository, files FileIndex, store *minio.Client, objectBucket string) *Service {
	return &Service{
		repo:         repo,
		files:        files,
		objectStore:  store,
		objectBucket: objectBucket,
	}
}

// CreateBucket creates a new bucket for the owner.
func (s *Service) CreateBucket(ctx context.Context, ownerID uuid.UUID, name string, description *string) (Bucket, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return Bucket{}, fmt.Errorf("bucket name required")
	}
	return s.repo.Create(ctx, ownerID, name, description)
}

// ListBuckets returns the user's buckets.
func (s *Service) ListBuckets(ctx context.Context, ownerID uuid.UUID) ([]Bucket, error) {
	return s.repo.List(ctx, ownerID)
}

// GetBucket returns a bucket ensuring ownership.
func (s *Service) GetBucket(ctx context.Context, ownerID, bucketID uuid.UUID) (Bucket, error) {
	return s.repo.Get(ctx, ownerID, bucketID)
}

// DeleteBucket removes a bucket, its metadata, and stored objects.
func (s *Service) DeleteBucket(ctx context.Context, ownerID, bucketID uuid.UUID) error {
	if _, err := s.repo.Get(ctx, ownerID, bucketID); err != nil {
		return err
	}

	if err := s.deleteObjects(ctx, bucketID); err != nil {
		return err
	}

	if err := s.repo.Delete(ctx, ownerID, bucketID); err != nil {
		return err
	}

	if err := s.repo.RecordUsageSnapshot(ctx, ownerID); err != nil {
		return err
	}
	return nil
}

func (s *Service) deleteObjects(ctx context.Context, bucketID uuid.UUID) error {
	if s.objectStore == nil || s.files == nil {
		return nil
	}
	objects, err := s.files.ListObjectsForBucket(ctx, bucketID)
	if err != nil {
		return fmt.Errorf("list bucket objects: %w", err)
	}
	for _, obj := range objects {
		if err := s.objectStore.RemoveObject(ctx, s.objectBucket, obj.ObjectName, minio.RemoveObjectOptions{}); err != nil {
			return fmt.Errorf("remove object %s: %w", obj.ObjectName, err)
		}
	}
	return nil
}
