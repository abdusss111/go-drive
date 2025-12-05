package storage

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/abduss/godrive/internal/config"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

const defaultObjectStoreTimeout = 5 * time.Second

// NewMinIOClient establishes a MinIO client using the provided configuration.
func NewMinIOClient(cfg config.MinIOConfig) (*minio.Client, error) {
	endpoint := cfg.Endpoint
	if !strings.Contains(endpoint, ":") {
		// default to MinIO API port when not supplied explicitly
		endpoint = fmt.Sprintf("%s:9000", endpoint)
	}

	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		Secure: cfg.UseSSL,
		Region: cfg.Region,
	})
	if err != nil {
		return nil, fmt.Errorf("create minio client: %w", err)
	}

	return client, nil
}

// EnsureBucket ensures the target bucket exists, creating it if necessary.
func EnsureBucket(ctx context.Context, client *minio.Client, bucket, region string) error {
	ctx, cancel := context.WithTimeout(ctx, defaultObjectStoreTimeout)
	defer cancel()

	exists, err := client.BucketExists(ctx, bucket)
	if err != nil {
		return fmt.Errorf("check bucket existence: %w", err)
	}

	if exists {
		return nil
	}

	if err := client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{Region: region}); err != nil {
		return fmt.Errorf("create bucket %q: %w", bucket, err)
	}

	return nil
}
