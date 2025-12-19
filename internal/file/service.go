package file

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"mime/multipart"
	"strings"

	"github.com/abduss/godrive/internal/bucket"
	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
)

const (
	defaultMaxFileSize = 100 * 1024 * 1024 // 100MB
)

// Service manages file lifecycle operations.
type metadataStore interface {
	Create(ctx context.Context, meta Metadata) (Metadata, error)
	List(ctx context.Context, ownerID, bucketID uuid.UUID, opts ListOptions) ([]Metadata, error)
	Get(ctx context.Context, ownerID, bucketID, fileID uuid.UUID) (Metadata, error)
	Delete(ctx context.Context, ownerID, bucketID, fileID uuid.UUID) (Metadata, error)
	FindCurrentVersion(ctx context.Context, ownerID, bucketID uuid.UUID, originalFilename string) (Metadata, error)
	MarkVersionsAsNotCurrent(ctx context.Context, bucketID uuid.UUID, originalFilename string) error
	GetVersions(ctx context.Context, ownerID, bucketID uuid.UUID, originalFilename string) ([]Metadata, error)
}

type Service struct {
	repo         metadataStore
	buckets      bucketStore
	objectStore  objectStore
	objectBucket string
	maxFileSize  int64
}

type bucketStore interface {
	Get(ctx context.Context, ownerID, bucketID uuid.UUID) (bucket.Bucket, error)
	UpdateUsage(ctx context.Context, bucketID uuid.UUID, deltaBytes int64, deltaFiles int64) error
	RecordUsageSnapshot(ctx context.Context, ownerID uuid.UUID) error
}

type objectStore interface {
	PutObject(ctx context.Context, bucketName, objectName string, reader io.Reader, objectSize int64, opts minio.PutObjectOptions) (minio.UploadInfo, error)
	GetObject(ctx context.Context, bucketName, objectName string, opts minio.GetObjectOptions) (io.ReadCloser, error)
	RemoveObject(ctx context.Context, bucketName, objectName string, opts minio.RemoveObjectOptions) error
}

// NewService constructs a file service.
func NewService(repo metadataStore, buckets bucketStore, store objectStore, objectBucket string) *Service {
	return &Service{
		repo:         repo,
		buckets:      buckets,
		objectStore:  store,
		objectBucket: objectBucket,
		maxFileSize:  defaultMaxFileSize,
	}
}

// Upload creates metadata and stores the object contents.
func (s *Service) Upload(ctx context.Context, ownerID, bucketID uuid.UUID, fileHeader *multipart.FileHeader) (Metadata, error) {
	if fileHeader == nil {
		return Metadata{}, fmt.Errorf("missing file payload")
	}

	if _, err := s.buckets.Get(ctx, ownerID, bucketID); err != nil {
		return Metadata{}, translateBucketError(err)
	}

	size := fileHeader.Size
	if size > s.maxFileSize {
		return Metadata{}, ErrFileTooLarge
	}

	originalFilename := sanitizeFilename(fileHeader.Filename)
	
	// Check if file with same name exists
	existingFile, err := s.repo.FindCurrentVersion(ctx, ownerID, bucketID, originalFilename)
	version := 1
	if err == nil {
		// File exists - create new version
		version = existingFile.Version + 1
		// Mark old versions as not current
		if err := s.repo.MarkVersionsAsNotCurrent(ctx, bucketID, originalFilename); err != nil {
			return Metadata{}, fmt.Errorf("mark versions as not current: %w", err)
		}
	} else if err != ErrFileNotFound {
		return Metadata{}, fmt.Errorf("check existing file: %w", err)
	}

	fileID := uuid.New()
	objectName := fmt.Sprintf("%s/%s", bucketID.String(), fileID.String())

	file, err := fileHeader.Open()
	if err != nil {
		return Metadata{}, fmt.Errorf("open upload file: %w", err)
	}
	defer file.Close()

	hasher := sha256.New()
	reader := io.TeeReader(file, hasher)

	putOpts := minio.PutObjectOptions{
		ContentType: detectContentType(fileHeader),
	}

	uploadInfo, err := s.objectStore.PutObject(ctx, s.objectBucket, objectName, reader, size, putOpts)
	if err != nil {
		return Metadata{}, fmt.Errorf("store object: %w", err)
	}

	actualSize := uploadInfo.Size
	if actualSize <= 0 {
		actualSize = size
	}
	if s.maxFileSize > 0 && actualSize > s.maxFileSize {
		_ = s.objectStore.RemoveObject(ctx, s.objectBucket, objectName, minio.RemoveObjectOptions{})
		return Metadata{}, ErrFileTooLarge
	}

	checksum := hex.EncodeToString(hasher.Sum(nil))

	meta := Metadata{
		ID:               fileID,
		BucketID:         bucketID,
		ObjectName:       objectName,
		OriginalFilename: originalFilename,
		SizeBytes:        actualSize,
		ContentType:      putOpts.ContentType,
		Checksum:         checksum,
		Version:          version,
		IsCurrent:        true,
	}

	stored, err := s.repo.Create(ctx, meta)
	if err != nil {
		_ = s.objectStore.RemoveObject(ctx, s.objectBucket, objectName, minio.RemoveObjectOptions{})
		return Metadata{}, err
	}

	// Update usage: if new version, only update size delta; if new file, increment count
	deltaFiles := int64(0)
	deltaBytes := stored.SizeBytes
	if version > 1 {
		// New version - subtract old file size, add new size
		deltaBytes = stored.SizeBytes - existingFile.SizeBytes
	} else {
		// New file - increment count
		deltaFiles = 1
	}
	
	if err := s.buckets.UpdateUsage(ctx, bucketID, deltaBytes, deltaFiles); err != nil {
		return Metadata{}, err
	}
	_ = s.buckets.RecordUsageSnapshot(ctx, ownerID)

	return stored, nil
}

// GetVersions returns all versions of a file.
func (s *Service) GetVersions(ctx context.Context, ownerID, bucketID uuid.UUID, originalFilename string) ([]Metadata, error) {
	if _, err := s.buckets.Get(ctx, ownerID, bucketID); err != nil {
		return nil, translateBucketError(err)
	}
	return s.repo.GetVersions(ctx, ownerID, bucketID, originalFilename)
}

// ListOptions contains filtering and pagination options for listing files.
type ListOptions struct {
	FilenameFilter string // Partial match on original_filename
	ContentType    string // Exact match on content_type
	MinSize        *int64 // Minimum file size in bytes
	MaxSize        *int64 // Maximum file size in bytes
	SortBy         string // Sort field: "name", "size", "created_at", "updated_at"
	SortOrder      string // "asc" or "desc"
	Limit          int    // Maximum number of results
	Offset         int    // Number of results to skip
}

// List returns file metadata for a user's bucket with optional filtering.
func (s *Service) List(ctx context.Context, ownerID, bucketID uuid.UUID, opts ListOptions) ([]Metadata, error) {
	if _, err := s.buckets.Get(ctx, ownerID, bucketID); err != nil {
		return nil, translateBucketError(err)
	}
	return s.repo.List(ctx, ownerID, bucketID, opts)
}

// Download retrieves metadata and object reader.
func (s *Service) Download(ctx context.Context, ownerID, bucketID, fileID uuid.UUID) (Metadata, io.ReadCloser, error) {
	meta, err := s.repo.Get(ctx, ownerID, bucketID, fileID)
	if err != nil {
		return Metadata{}, nil, err
	}

	object, err := s.objectStore.GetObject(ctx, s.objectBucket, meta.ObjectName, minio.GetObjectOptions{})
	if err != nil {
		return Metadata{}, nil, fmt.Errorf("fetch object: %w", err)
	}

	return meta, object, nil
}

// Delete removes the file from storage and metadata.
func (s *Service) Delete(ctx context.Context, ownerID, bucketID, fileID uuid.UUID) error {
	meta, err := s.repo.Delete(ctx, ownerID, bucketID, fileID)
	if err != nil {
		return err
	}

	if err := s.objectStore.RemoveObject(ctx, s.objectBucket, meta.ObjectName, minio.RemoveObjectOptions{}); err != nil {
		return fmt.Errorf("remove object: %w", err)
	}

	if err := s.buckets.UpdateUsage(ctx, bucketID, -meta.SizeBytes, -1); err != nil {
		return err
	}
	_ = s.buckets.RecordUsageSnapshot(ctx, ownerID)
	return nil
}

func detectContentType(fileHeader *multipart.FileHeader) string {
	if fileHeader == nil {
		return "application/octet-stream"
	}
	contentType := fileHeader.Header.Get("Content-Type")
	if contentType != "" {
		return contentType
	}
	return "application/octet-stream"
}

func sanitizeFilename(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return "upload"
	}
	return name
}

func translateBucketError(err error) error {
	switch err {
	case bucket.ErrBucketNotFound:
		return ErrBucketMismatch
	default:
		return err
	}
}
