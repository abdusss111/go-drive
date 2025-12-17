package bucket

import (
	"time"

	"github.com/google/uuid"
)

// Bucket represents a logical container for user files.
type Bucket struct {
	ID          uuid.UUID  `json:"id"`
	OwnerID     uuid.UUID  `json:"owner_id"`
	Name        string     `json:"name"`
	Description *string    `json:"description,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	Usage       UsageStats `json:"usage"`
}

// UsageStats reflects aggregate file statistics for a bucket.
type UsageStats struct {
	TotalBytes int64 `json:"total_bytes"`
	FileCount  int64 `json:"file_count"`
}
