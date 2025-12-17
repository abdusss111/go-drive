package presigned

import "time"

// PresignedURLRequest represents a request to generate a presigned URL.
type PresignedURLRequest struct {
	Method string `json:"method" binding:"required,oneof=GET PUT"` // GET for download, PUT for upload
	TTL    string `json:"ttl,omitempty"`                         // Optional TTL (e.g., "1h", "30m")
}

// PresignedURLResponse represents a presigned URL response.
type PresignedURLResponse struct {
	URL       string    `json:"url"`
	ExpiresAt time.Time `json:"expires_at"`
	Method    string    `json:"method"`
}

