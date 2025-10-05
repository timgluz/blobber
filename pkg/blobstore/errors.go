package blobstore

import "errors"

var (
	ErrNoValidCredentials = errors.New("no valid credentials provided")
	ErrConfigLoadFailed   = errors.New("failed to load configuration")
	ErrBlobNotFound       = errors.New("blob not found")
	ErrBucketNotFound     = errors.New("bucket not found")
	ErrNoValidBucket      = errors.New("no valid bucket provided")
	ErrNoValidBlobClient  = errors.New("no valid blob client provided")
)
