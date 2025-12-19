package presigned

import (
	"context"
	"net/url"
	"time"

	"github.com/abduss/godrive/internal/bucket"
	"github.com/abduss/godrive/internal/file"
	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
)

// BucketServiceAdapter adapts bucket.Service to bucketChecker interface.
type BucketServiceAdapter struct {
	Service *bucket.Service
}

func (a *BucketServiceAdapter) GetBucket(ctx context.Context, ownerID, bucketID uuid.UUID) (interface{}, error) {
	return a.Service.GetBucket(ctx, ownerID, bucketID)
}

// FileRepositoryAdapter adapts file.Repository to fileChecker interface.
type FileRepositoryAdapter struct {
	Repo *file.Repository
}

func (a *FileRepositoryAdapter) Get(ctx context.Context, ownerID, bucketID, fileID uuid.UUID) (interface{}, error) {
	return a.Repo.Get(ctx, ownerID, bucketID, fileID)
}

// MinIOClientAdapter adapts minio.Client to objectStore interface.
type MinIOClientAdapter struct {
	Client *minio.Client
}

func (a *MinIOClientAdapter) PresignedGetObject(ctx context.Context, bucketName string, objectName string, expiry time.Duration, reqParams interface{}) (*url.URL, error) {
	var values url.Values
	if reqParams != nil {
		if v, ok := reqParams.(url.Values); ok {
			values = v
		} else if m, ok := reqParams.(map[string][]string); ok {
			values = m
		}
	}
	return a.Client.PresignedGetObject(ctx, bucketName, objectName, expiry, values)
}

func (a *MinIOClientAdapter) PresignedPutObject(ctx context.Context, bucketName string, objectName string, expiry time.Duration) (*url.URL, error) {
	return a.Client.PresignedPutObject(ctx, bucketName, objectName, expiry)
}

