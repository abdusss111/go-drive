package file

import (
	"context"
	"io"

	"github.com/minio/minio-go/v7"
)

// MinIOStore adapts minio.Client to the objectStore interface.
type MinIOStore struct {
	client *minio.Client
}

// NewMinIOStore constructs an adapter.
func NewMinIOStore(client *minio.Client) *MinIOStore {
	return &MinIOStore{client: client}
}

func (s *MinIOStore) PutObject(ctx context.Context, bucketName, objectName string, reader io.Reader, objectSize int64, opts minio.PutObjectOptions) (minio.UploadInfo, error) {
	return s.client.PutObject(ctx, bucketName, objectName, reader, objectSize, opts)
}

func (s *MinIOStore) GetObject(ctx context.Context, bucketName, objectName string, opts minio.GetObjectOptions) (io.ReadCloser, error) {
	return s.client.GetObject(ctx, bucketName, objectName, opts)
}

func (s *MinIOStore) RemoveObject(ctx context.Context, bucketName, objectName string, opts minio.RemoveObjectOptions) error {
	return s.client.RemoveObject(ctx, bucketName, objectName, opts)
}
