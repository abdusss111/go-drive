package file

import "errors"

var (
	// ErrBucketMismatch indicates a file does not belong to the provided bucket or owner.
	ErrBucketMismatch = errors.New("bucket mismatch")
	// ErrFileNotFound signals that the file could not be located.
	ErrFileNotFound = errors.New("file not found")
	// ErrFileTooLarge signals that the upload exceeds configured limits.
	ErrFileTooLarge = errors.New("file too large")
)
