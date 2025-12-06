package usage

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
)

func TestCheckQuota_Success(t *testing.T) {
	repo := &fakeQuotaChecker{
		userUsage: map[uuid.UUID]usageData{
			uuid.New(): {bytes: 100, files: 5},
		},
		userQuota: map[uuid.UUID]quotaData{
			uuid.New(): {bytes: 1000, files: 100},
		},
	}
	service := NewService(repo)

	userID := uuid.New()
	repo.userUsage[userID] = usageData{bytes: 100, files: 5}
	repo.userQuota[userID] = quotaData{bytes: 1000, files: 100}

	err := service.CheckQuota(context.Background(), userID, 500, 10)
	if err != nil {
		t.Fatalf("CheckQuota should succeed, got error: %v", err)
	}
}

func TestCheckQuota_BytesExceeded(t *testing.T) {
	repo := &fakeQuotaChecker{
		userUsage: make(map[uuid.UUID]usageData),
		userQuota: make(map[uuid.UUID]quotaData),
	}
	service := NewService(repo)

	userID := uuid.New()
	repo.userUsage[userID] = usageData{bytes: 900, files: 5}
	repo.userQuota[userID] = quotaData{bytes: 1000, files: 100}

	err := service.CheckQuota(context.Background(), userID, 200, 1)
	if err == nil {
		t.Fatal("CheckQuota should fail with bytes exceeded")
	}
	if !errors.Is(err, ErrQuotaBytesExceeded) {
		t.Fatalf("Expected ErrQuotaBytesExceeded, got: %v", err)
	}
}

func TestCheckQuota_FilesExceeded(t *testing.T) {
	repo := &fakeQuotaChecker{
		userUsage: make(map[uuid.UUID]usageData),
		userQuota: make(map[uuid.UUID]quotaData),
	}
	service := NewService(repo)

	userID := uuid.New()
	repo.userUsage[userID] = usageData{bytes: 100, files: 95}
	repo.userQuota[userID] = quotaData{bytes: 1000, files: 100}

	err := service.CheckQuota(context.Background(), userID, 50, 10)
	if err == nil {
		t.Fatal("CheckQuota should fail with files exceeded")
	}
	if !errors.Is(err, ErrQuotaFilesExceeded) {
		t.Fatalf("Expected ErrQuotaFilesExceeded, got: %v", err)
	}
}

func TestCheckQuota_ZeroQuota(t *testing.T) {
	repo := &fakeQuotaChecker{
		userUsage: make(map[uuid.UUID]usageData),
		userQuota: make(map[uuid.UUID]quotaData),
	}
	service := NewService(repo)

	userID := uuid.New()
	repo.userUsage[userID] = usageData{bytes: 0, files: 0}
	repo.userQuota[userID] = quotaData{bytes: 0, files: 0}

	// With zero quota (unlimited), upload should be allowed
	err := service.CheckQuota(context.Background(), userID, 1, 1)
	if err != nil {
		t.Fatalf("CheckQuota should succeed with zero quota (unlimited), got error: %v", err)
	}
}

func TestGetUserUsage(t *testing.T) {
	repo := &fakeQuotaChecker{
		userUsage: make(map[uuid.UUID]usageData),
		userQuota: make(map[uuid.UUID]quotaData),
	}
	service := NewService(repo)

	userID := uuid.New()
	repo.userUsage[userID] = usageData{bytes: 500, files: 10}
	repo.userQuota[userID] = quotaData{bytes: 1000, files: 100}

	usage, err := service.GetUserUsage(context.Background(), userID)
	if err != nil {
		t.Fatalf("GetUserUsage returned error: %v", err)
	}

	if usage.TotalBytes != 500 {
		t.Fatalf("Expected TotalBytes=500, got %d", usage.TotalBytes)
	}
	if usage.TotalFiles != 10 {
		t.Fatalf("Expected TotalFiles=10, got %d", usage.TotalFiles)
	}
	if usage.QuotaBytes != 1000 {
		t.Fatalf("Expected QuotaBytes=1000, got %d", usage.QuotaBytes)
	}
	if usage.QuotaFiles != 100 {
		t.Fatalf("Expected QuotaFiles=100, got %d", usage.QuotaFiles)
	}
	if usage.UsagePercentBytes != 50.0 {
		t.Fatalf("Expected UsagePercentBytes=50.0, got %f", usage.UsagePercentBytes)
	}
	if usage.UsagePercentFiles != 10.0 {
		t.Fatalf("Expected UsagePercentFiles=10.0, got %f", usage.UsagePercentFiles)
	}
}

// --- helpers & fakes ---

type usageData struct {
	bytes int64
	files int64
}

type quotaData struct {
	bytes int64
	files int64
}

type fakeQuotaChecker struct {
	userUsage map[uuid.UUID]usageData
	userQuota map[uuid.UUID]quotaData
}

func (f *fakeQuotaChecker) GetUserUsage(ctx context.Context, userID uuid.UUID) (totalBytes, totalFiles int64, err error) {
	data, ok := f.userUsage[userID]
	if !ok {
		return 0, 0, errors.New("user not found")
	}
	return data.bytes, data.files, nil
}

func (f *fakeQuotaChecker) GetUserQuota(ctx context.Context, userID uuid.UUID) (quotaBytes, quotaFiles int64, err error) {
	data, ok := f.userQuota[userID]
	if !ok {
		return 0, 0, errors.New("user not found")
	}
	return data.bytes, data.files, nil
}

func (f *fakeQuotaChecker) GetBucketUsage(ctx context.Context, bucketID uuid.UUID) (totalBytes, totalFiles int64, err error) {
	// For tests, return zero usage
	return 0, 0, nil
}

