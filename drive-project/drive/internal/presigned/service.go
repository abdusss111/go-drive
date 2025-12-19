package presigned

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/google/uuid"
)

// bucketChecker verifies bucket ownership.
type bucketChecker interface {
	GetBucket(ctx context.Context, ownerID, bucketID uuid.UUID) (interface{}, error)
}

// fileChecker verifies file ownership.
type fileChecker interface {
	Get(ctx context.Context, ownerID, bucketID, fileID uuid.UUID) (interface{}, error)
}

// objectStore provides presigned URL generation.
type objectStore interface {
	PresignedGetObject(ctx context.Context, bucketName string, objectName string, expiry time.Duration, reqParams interface{}) (*url.URL, error)
	PresignedPutObject(ctx context.Context, bucketName string, objectName string, expiry time.Duration) (*url.URL, error)
}

// Service provides presigned URL generation functionality.
type Service struct {
	bucketChecker bucketChecker
	fileChecker   fileChecker
	objectStore   objectStore
	objectBucket  string
	defaultTTL    time.Duration
}

// NewService constructs a presigned URL service.
func NewService(bucketChecker bucketChecker, fileChecker fileChecker, objectStore objectStore, objectBucket string) *Service {
	return &Service{
		bucketChecker: bucketChecker,
		fileChecker:   fileChecker,
		objectStore:   objectStore,
		objectBucket:  objectBucket,
		defaultTTL:    time.Hour, // Default 1 hour
	}
}

// GeneratePresignedURL creates a presigned URL for a file.
func (s *Service) GeneratePresignedURL(ctx context.Context, ownerID, bucketID, fileID uuid.UUID, method string, ttl time.Duration) (*PresignedURLResponse, error) {
	// Verify bucket ownership
	_, err := s.bucketChecker.GetBucket(ctx, ownerID, bucketID)
	if err != nil {
		return nil, fmt.Errorf("verify bucket: %w", err)
	}

	// Use default TTL if not provided
	if ttl <= 0 {
		ttl = s.defaultTTL
	}

	objectName := fmt.Sprintf("%s/%s", bucketID.String(), fileID.String())
	expiresAt := time.Now().Add(ttl)

	var presignedURL *url.URL

	switch method {
	case "GET":
		// Verify file ownership and get original filename
		metaObj, err := s.fileChecker.Get(ctx, ownerID, bucketID, fileID)
		if err != nil {
			return nil, fmt.Errorf("verify file: %w", err)
		}

		reqParams := make(url.Values)
		// If we can identify the filename, set it in the response headers
		// This ensures that when the link is clicked, the file downloads with its original name
		if meta, ok := metaObj.(interface{ GetFilename() string }); ok {
			reqParams.Set("response-content-disposition", fmt.Sprintf("attachment; filename=\"%s\"", meta.GetFilename()))
		} else if meta, ok := metaObj.(interface{ GetOriginalFilename() string }); ok {
			// Fallback for different metadata structures
			reqParams.Set("response-content-disposition", fmt.Sprintf("attachment; filename=\"%s\"", meta.GetOriginalFilename()))
		}

		presignedURL, err = s.objectStore.PresignedGetObject(ctx, s.objectBucket, objectName, ttl, reqParams)
	case "PUT":
		presignedURL, err = s.objectStore.PresignedPutObject(ctx, s.objectBucket, objectName, ttl)
	default:
		return nil, fmt.Errorf("unsupported method: %s", method)
	}

	if err != nil {
		return nil, fmt.Errorf("generate presigned URL: %w", err)
	}

	return &PresignedURLResponse{
		URL:       presignedURL.String(),
		ExpiresAt: expiresAt,
		Method:    method,
	}, nil
}

