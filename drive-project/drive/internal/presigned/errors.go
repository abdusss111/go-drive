package presigned

import "errors"

var (
	// ErrPresignedURLNotFound indicates the presigned URL was not found.
	ErrPresignedURLNotFound = errors.New("presigned URL not found")

	// ErrPresignedURLExpired indicates the presigned URL has expired.
	ErrPresignedURLExpired = errors.New("presigned URL expired")

	// ErrInvalidPresignedURL indicates the presigned URL is invalid.
	ErrInvalidPresignedURL = errors.New("invalid presigned URL")
)

