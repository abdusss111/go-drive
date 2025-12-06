package usage

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

// quotaChecker defines the interface for checking quotas.
type quotaChecker interface {
	GetUserUsage(ctx context.Context, userID uuid.UUID) (totalBytes, totalFiles int64, err error)
	GetUserQuota(ctx context.Context, userID uuid.UUID) (quotaBytes, quotaFiles int64, err error)
	GetBucketUsage(ctx context.Context, bucketID uuid.UUID) (totalBytes, totalFiles int64, err error)
}

// Service provides usage and quota management functionality.
type Service struct {
	repo quotaChecker
}

// NewService constructs a usage service.
func NewService(repo quotaChecker) *Service {
	return &Service{repo: repo}
}

// CheckQuota verifies that adding a new file would not exceed the user's quota.
// Returns an error if the quota would be exceeded.
func (s *Service) CheckQuota(ctx context.Context, userID uuid.UUID, newFileSize, newFileCount int64) error {
	// Get current usage
	currentBytes, currentFiles, err := s.repo.GetUserUsage(ctx, userID)
	if err != nil {
		return fmt.Errorf("get user usage: %w", err)
	}

	// Get user quotas
	quotaBytes, quotaFiles, err := s.repo.GetUserQuota(ctx, userID)
	if err != nil {
		return fmt.Errorf("get user quota: %w", err)
	}

	// Check bytes quota
	// If quota is 0, it means unlimited, so we only check if quota > 0
	if quotaBytes > 0 {
		if currentBytes+newFileSize > quotaBytes {
			return fmt.Errorf("%w: current=%d, adding=%d, limit=%d",
				ErrQuotaBytesExceeded, currentBytes, newFileSize, quotaBytes)
		}
	}

	// Check files quota
	// If quota is 0, it means unlimited, so we only check if quota > 0
	if quotaFiles > 0 {
		if currentFiles+newFileCount > quotaFiles {
			return fmt.Errorf("%w: current=%d, adding=%d, limit=%d",
				ErrQuotaFilesExceeded, currentFiles, newFileCount, quotaFiles)
		}
	}

	return nil
}

// GetUserUsage returns the complete usage information for a user.
func (s *Service) GetUserUsage(ctx context.Context, userID uuid.UUID) (UsageResponse, error) {
	totalBytes, totalFiles, err := s.repo.GetUserUsage(ctx, userID)
	if err != nil {
		return UsageResponse{}, fmt.Errorf("get user usage: %w", err)
	}

	quotaBytes, quotaFiles, err := s.repo.GetUserQuota(ctx, userID)
	if err != nil {
		return UsageResponse{}, fmt.Errorf("get user quota: %w", err)
	}

	// Calculate usage percentages
	var usagePercentBytes, usagePercentFiles float64
	if quotaBytes > 0 {
		usagePercentBytes = float64(totalBytes) / float64(quotaBytes) * 100
	}
	if quotaFiles > 0 {
		usagePercentFiles = float64(totalFiles) / float64(quotaFiles) * 100
	}

	return UsageResponse{
		TotalBytes:        totalBytes,
		TotalFiles:        totalFiles,
		QuotaBytes:        quotaBytes,
		QuotaFiles:        quotaFiles,
		UsagePercentBytes: usagePercentBytes,
		UsagePercentFiles: usagePercentFiles,
	}, nil
}

// GetBucketUsage returns the usage information for a specific bucket.
func (s *Service) GetBucketUsage(ctx context.Context, bucketID uuid.UUID) (BucketUsageResponse, error) {
	totalBytes, totalFiles, err := s.repo.GetBucketUsage(ctx, bucketID)
	if err != nil {
		return BucketUsageResponse{}, fmt.Errorf("get bucket usage: %w", err)
	}

	return BucketUsageResponse{
		BucketID:   bucketID,
		TotalBytes: totalBytes,
		TotalFiles: totalFiles,
	}, nil
}

