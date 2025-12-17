package file

import (
	"bytes"
	"context"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/abduss/godrive/internal/bucket"
	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
)

func TestUploadStoresMetadataAndUpdatesUsage(t *testing.T) {
	repo := newFakeRepo()
	buckets := &fakeBucketStore{
		buckets: map[uuid.UUID]bucket.Bucket{},
	}
	objectStore := &fakeObjectStore{}
	service := NewService(repo, buckets, objectStore, "godrive")

	ownerID := uuid.New()
	bucketID := uuid.New()
	buckets.buckets[bucketID] = bucket.Bucket{ID: bucketID, OwnerID: ownerID, Name: "docs"}

	fileHeader := buildFileHeader(t, "file", "notes.txt", "text/plain", []byte("hello world"))

	meta, err := service.Upload(context.Background(), ownerID, bucketID, fileHeader)
	if err != nil {
		t.Fatalf("Upload returned error: %v", err)
	}

	if meta.OriginalFilename != "notes.txt" {
		t.Fatalf("unexpected filename: %s", meta.OriginalFilename)
	}
	if len(repo.records) != 1 {
		t.Fatalf("expected metadata stored, got %d", len(repo.records))
	}
	if !objectStore.putCalled {
		t.Fatalf("expected PutObject to be called")
	}
	if buckets.usageDelta != meta.SizeBytes {
		t.Fatalf("expected usage delta %d, got %d", meta.SizeBytes, buckets.usageDelta)
	}
}

func TestDeleteRemovesMetadataAndObject(t *testing.T) {
	repo := newFakeRepo()
	buckets := &fakeBucketStore{
		buckets: map[uuid.UUID]bucket.Bucket{},
	}
	objectStore := &fakeObjectStore{reader: bytes.NewReader([]byte("payload"))}
	service := NewService(repo, buckets, objectStore, "godrive")

	ownerID := uuid.New()
	bucketID := uuid.New()
	buckets.buckets[bucketID] = bucket.Bucket{ID: bucketID, OwnerID: ownerID, Name: "archive"}

	fileHeader := buildFileHeader(t, "file", "data.bin", "application/octet-stream", []byte("payload"))
	meta, err := service.Upload(context.Background(), ownerID, bucketID, fileHeader)
	if err != nil {
		t.Fatalf("Upload returned error: %v", err)
	}

	if err := service.Delete(context.Background(), ownerID, bucketID, meta.ID); err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}

	if objectStore.removeCount != 1 {
		t.Fatalf("expected RemoveObject called once, got %d", objectStore.removeCount)
	}
	if len(repo.records) != 0 {
		t.Fatalf("expected metadata removed, remaining %d", len(repo.records))
	}
	if buckets.usageDelta != 0 {
		t.Fatalf("expected usage delta reset to 0, got %d", buckets.usageDelta)
	}
}

// --- helpers & fakes ---

func buildFileHeader(t *testing.T, fieldName, filename, contentType string, content []byte) *multipart.FileHeader {
	t.Helper()
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile(fieldName, filename)
	if err != nil {
		t.Fatalf("CreateFormFile error: %v", err)
	}
	if _, err := part.Write(content); err != nil {
		t.Fatalf("write part: %v", err)
	}
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	if err := req.ParseMultipartForm(int64(len(content)) + 1024); err != nil {
		t.Fatalf("ParseMultipartForm: %v", err)
	}

	return req.MultipartForm.File[fieldName][0]
}

type fakeRepo struct {
	records map[uuid.UUID]Metadata
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{records: make(map[uuid.UUID]Metadata)}
}

func (f *fakeRepo) Create(ctx context.Context, meta Metadata) (Metadata, error) {
	f.records[meta.ID] = meta
	meta.CreatedAt = time.Now()
	meta.UpdatedAt = meta.CreatedAt
	return meta, nil
}

func (f *fakeRepo) List(ctx context.Context, ownerID, bucketID uuid.UUID) ([]Metadata, error) {
	var list []Metadata
	for _, m := range f.records {
		if m.BucketID == bucketID {
			list = append(list, m)
		}
	}
	return list, nil
}

func (f *fakeRepo) Get(ctx context.Context, ownerID, bucketID, fileID uuid.UUID) (Metadata, error) {
	meta, ok := f.records[fileID]
	if !ok {
		return Metadata{}, ErrFileNotFound
	}
	return meta, nil
}

func (f *fakeRepo) Delete(ctx context.Context, ownerID, bucketID, fileID uuid.UUID) (Metadata, error) {
	meta, ok := f.records[fileID]
	if !ok {
		return Metadata{}, ErrFileNotFound
	}
	delete(f.records, fileID)
	return meta, nil
}

type fakeBucketStore struct {
	buckets    map[uuid.UUID]bucket.Bucket
	usageDelta int64
}

func (f *fakeBucketStore) Get(ctx context.Context, ownerID, bucketID uuid.UUID) (bucket.Bucket, error) {
	b, ok := f.buckets[bucketID]
	if !ok || b.OwnerID != ownerID {
		return bucket.Bucket{}, bucket.ErrBucketNotFound
	}
	return b, nil
}

func (f *fakeBucketStore) UpdateUsage(ctx context.Context, bucketID uuid.UUID, deltaBytes int64, deltaFiles int64) error {
	f.usageDelta += deltaBytes
	return nil
}

func (f *fakeBucketStore) RecordUsageSnapshot(ctx context.Context, ownerID uuid.UUID) error {
	return nil
}

type fakeObjectStore struct {
	putCalled   bool
	removeCount int
	reader      io.Reader
}

func (f *fakeObjectStore) PutObject(ctx context.Context, bucketName, objectName string, reader io.Reader, objectSize int64, opts minio.PutObjectOptions) (minio.UploadInfo, error) {
	f.putCalled = true
	data, err := io.ReadAll(reader)
	if err != nil {
		return minio.UploadInfo{}, err
	}
	return minio.UploadInfo{Size: int64(len(data))}, nil
}

func (f *fakeObjectStore) GetObject(ctx context.Context, bucketName, objectName string, opts minio.GetObjectOptions) (io.ReadCloser, error) {
	if f.reader == nil {
		f.reader = bytes.NewReader([]byte{})
	}
	return io.NopCloser(f.reader), nil
}

func (f *fakeObjectStore) RemoveObject(ctx context.Context, bucketName, objectName string, opts minio.RemoveObjectOptions) error {
	f.removeCount++
	return nil
}
