package presigned

import (
	"context"
	"net/url"
	"testing"
	"time"

	"github.com/abduss/godrive/internal/bucket"
	"github.com/abduss/godrive/internal/file"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestGenerateGetSuccess(t *testing.T) {
	userID := uuid.New()
	bucketID := uuid.New()
	fileID := uuid.New()

	svc := NewService(
		noopRepo{},
		&fakeFileRepo{meta: file.Metadata{ID: fileID, BucketID: bucketID, ObjectName: "obj"}},
		&fakeBucketRepo{bucket: bucket.Bucket{ID: bucketID, OwnerID: userID}},
		fakePresigner{},
		"test-bucket",
		15*time.Minute,
		0,
		"secret",
	)

	resp, err := svc.Generate(context.Background(), GenerateRequest{
		UserID:   userID,
		BucketID: bucketID,
		FileID:   fileID,
		Method:   "GET",
		TTL:      10 * time.Minute,
	})
	require.NoError(t, err)
	require.NotEmpty(t, resp.URL)
	require.Equal(t, "GET", resp.Method)
	require.False(t, resp.ExpiresAt.IsZero())
	require.NotEmpty(t, resp.ScopeToken)
}

func TestGenerateFailsOnMissingBucket(t *testing.T) {
	svc := NewService(
		nil,
		&fakeFileRepo{err: file.ErrFileNotFound},
		&fakeBucketRepo{err: bucket.ErrBucketNotFound},
		fakePresigner{},
		"test-bucket",
		15*time.Minute,
		0,
		"",
	)

	_, err := svc.Generate(context.Background(), GenerateRequest{
		UserID:   uuid.New(),
		BucketID: uuid.New(),
		FileID:   uuid.New(),
		Method:   "GET",
	})
	require.Equal(t, bucket.ErrBucketNotFound, err)
}

func TestGenerateClampsTTL(t *testing.T) {
	maxTTL := 30 * time.Minute
	bID := uuid.New()
	fID := uuid.New()
	svc := NewService(
		nil,
		&fakeFileRepo{meta: file.Metadata{ID: fID, BucketID: bID, ObjectName: "obj"}},
		&fakeBucketRepo{bucket: bucket.Bucket{ID: bID}},
		fakePresigner{},
		"test-bucket",
		10*time.Minute,
		maxTTL,
		"",
	)

	resp, err := svc.Generate(context.Background(), GenerateRequest{
		UserID:   uuid.New(),
		BucketID: uuid.New(),
		FileID:   uuid.New(),
		Method:   "GET",
		TTL:      time.Hour,
	})
	require.NoError(t, err)
	require.True(t, resp.ExpiresAt.Sub(time.Now()) <= maxTTL+time.Minute)
}

type noopRepo struct{}

func (noopRepo) Log(ctx context.Context, entry AuditEntry) error { return nil }

type fakeFileRepo struct {
	meta file.Metadata
	err  error
}

func (f *fakeFileRepo) Get(ctx context.Context, ownerID, bucketID, fileID uuid.UUID) (file.Metadata, error) {
	if f.err != nil {
		return file.Metadata{}, f.err
	}
	return f.meta, nil
}

type fakeBucketRepo struct {
	bucket bucket.Bucket
	err    error
}

func (f *fakeBucketRepo) Get(ctx context.Context, ownerID, bucketID uuid.UUID) (bucket.Bucket, error) {
	if f.err != nil {
		return bucket.Bucket{}, f.err
	}
	return f.bucket, nil
}

type fakePresigner struct{}

func (fakePresigner) PresignedGetObject(ctx context.Context, bucketName, objectName string, expiry time.Duration, reqParams url.Values) (*url.URL, error) {
	return &url.URL{Scheme: "https", Host: "example.com", Path: "/" + bucketName + "/" + objectName}, nil
}

func (fakePresigner) PresignedPutObject(ctx context.Context, bucketName, objectName string, expiry time.Duration) (*url.URL, error) {
	return &url.URL{Scheme: "https", Host: "example.com", Path: "/" + bucketName + "/" + objectName}, nil
}
