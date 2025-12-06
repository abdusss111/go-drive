package presigned

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
)

type fakeRepo struct {
	saveRecordCalled bool
	saveAuditCalled  bool
}

func (r *fakeRepo) SaveRecord(ctx context.Context, rec Record) error {
	r.saveRecordCalled = true
	return nil
}

func (r *fakeRepo) SaveAudit(ctx context.Context, rec AuditRecord) error {
	r.saveAuditCalled = true
	return nil
}

type fakeMinio struct {
	url string
}

func (m *fakeMinio) PresignedGetObject(ctx context.Context, bucket, object string, expiry time.Duration, params interface{}) (interface{}, error) {
	return struct{ URL string }{URL: m.url}, nil
}

func (m *fakeMinio) PresignedPutObject(ctx context.Context, bucket, object string, expiry time.Duration) (interface{}, error) {
	return struct{ URL string }{URL: m.url}, nil
}

func TestGenerateURL_PUT(t *testing.T) {
	repo := &fakeRepo{}
	minio := &fakeMinio{url: "https://example.com/upload"}

	svc := &Service{
		client: minio,
		ttl:    time.Minute,
		repo:   repo,
	}

	url, rec, audit, err := svc.GenerateURL(
		context.Background(),
		"bucket",
		"file.txt",
		"PUT",
		uuid.New(),
		uuid.New(),
		uuid.New(),
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if url == "" {
		t.Fatalf("url must not be empty")
	}

	if rec.Method != "PUT" {
		t.Fatalf("wrong method in record")
	}

	if audit.Method != "PUT" {
		t.Fatalf("wrong method in audit")
	}

	if !repo.saveRecordCalled {
		t.Fatalf("record must be saved")
	}

	if !repo.saveAuditCalled {
		t.Fatalf("audit must be saved")
	}
}

func TestGenerateURL_InvalidMethod(t *testing.T) {
	repo := &fakeRepo{}
	minio := &fakeMinio{}
	svc := &Service{client: minio, ttl: time.Minute, repo: repo}

	_, _, _, err := svc.GenerateURL(
		context.Background(),
		"bucket",
		"file",
		"DELETE",
		uuid.New(),
		uuid.New(),
		uuid.New(),
	)

	if err == nil {
		t.Fatalf("expected error for invalid method")
	}
}
