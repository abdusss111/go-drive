package presigned

import (
	"time"

	"github.com/google/uuid"
)

// GenerateRequest carries the inputs required to build a presigned URL.
type GenerateRequest struct {
	UserID   uuid.UUID
	BucketID uuid.UUID
	FileID   uuid.UUID
	Method   string
	TTL      time.Duration
}

// GenerateResponse represents the presigned URL and related metadata.
type GenerateResponse struct {
	URL        string    `json:"url"`
	Method     string    `json:"method"`
	ExpiresAt  time.Time `json:"expires_at"`
	ScopeToken string    `json:"scope_token"`
}
