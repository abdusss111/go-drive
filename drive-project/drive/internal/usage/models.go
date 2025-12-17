package usage

import "github.com/google/uuid"

// UsageResponse represents the usage statistics for a user.
type UsageResponse struct {
	TotalBytes        int64   `json:"total_bytes"`
	TotalFiles        int64   `json:"total_files"`
	QuotaBytes        int64   `json:"quota_bytes"`
	QuotaFiles        int64   `json:"quota_files"`
	UsagePercentBytes float64 `json:"usage_percent_bytes"`
	UsagePercentFiles float64 `json:"usage_percent_files"`
}

// BucketUsageResponse represents the usage statistics for a specific bucket.
type BucketUsageResponse struct {
	BucketID   uuid.UUID `json:"bucket_id"`
	TotalBytes int64     `json:"total_bytes"`
	TotalFiles int64     `json:"total_files"`
}

