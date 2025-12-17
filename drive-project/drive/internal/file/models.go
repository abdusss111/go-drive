package file

import (
	"time"

	"github.com/google/uuid"
)

// Metadata represents stored information about an object.
type Metadata struct {
	ID               uuid.UUID `json:"id"`
	BucketID         uuid.UUID `json:"bucket_id"`
	ObjectName       string    `json:"object_name"`
	OriginalFilename string    `json:"original_filename"`
	SizeBytes        int64     `json:"size_bytes"`
	ContentType      string    `json:"content_type"`
	Checksum         string    `json:"checksum"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}
