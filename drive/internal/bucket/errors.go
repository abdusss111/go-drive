package bucket

import "errors"

var (
	// ErrBucketNotFound indicates the requested bucket does not exist for the user.
	ErrBucketNotFound = errors.New("bucket not found")
	// ErrBucketNameExists is returned when a user attempts to create a duplicate bucket name.
	ErrBucketNameExists = errors.New("bucket name already exists")
)
