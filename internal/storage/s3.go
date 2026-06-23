package storage

import (
	"context"
	"fmt"
	"io"
)

// S3Config holds credentials and settings for S3-compatible object storage.
type S3Config struct {
	Bucket    string
	Region    string
	Endpoint  string // optional: for S3-compatible services like Cloudflare R2
	AccessKey string
	SecretKey string
	BaseURL   string // CDN or public URL prefix
}

// S3Store is an S3-compatible object store.
// NOTE: This is an interface-compatible stub that can be wired to the AWS SDK
// or any S3-compatible client (Cloudflare R2, MinIO, etc.) without changing
// call-sites.
type S3Store struct {
	cfg S3Config
}

// NewS3Store creates an S3Store from the given config.
// An actual SDK client would be initialised here.
func NewS3Store(cfg S3Config) *S3Store {
	return &S3Store{cfg: cfg}
}

// Save uploads data from r to the bucket at key and returns the public URL.
func (s *S3Store) Save(ctx context.Context, key string, r io.Reader, contentType string) (string, error) {
	// TODO: replace with real AWS/R2 SDK upload.
	// Example (aws-sdk-go-v2):
	//   client.PutObject(ctx, &s3.PutObjectInput{Bucket: &s.cfg.Bucket, Key: &key, Body: r})
	return "", fmt.Errorf("S3 storage not yet wired — set STORAGE_DRIVER=local or implement S3 client")
}

// Delete removes the object at key.
func (s *S3Store) Delete(ctx context.Context, key string) error {
	return fmt.Errorf("S3 storage not yet wired")
}

// URL returns the public CDN/bucket URL for a key.
func (s *S3Store) URL(key string) string {
	return s.cfg.BaseURL + "/" + key
}
