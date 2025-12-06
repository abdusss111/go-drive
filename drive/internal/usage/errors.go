package usage

import "errors"

var (
	// ErrQuotaExceeded indicates that the user has exceeded their quota limit.
	ErrQuotaExceeded = errors.New("quota exceeded")

	// ErrQuotaBytesExceeded indicates that the user has exceeded their storage quota in bytes.
	ErrQuotaBytesExceeded = errors.New("storage quota exceeded")

	// ErrQuotaFilesExceeded indicates that the user has exceeded their file count quota.
	ErrQuotaFilesExceeded = errors.New("file count quota exceeded")
)

