package storage

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// LocalStore stores files on the local filesystem.
type LocalStore struct {
	baseDir string
	baseURL string // URL prefix for serving files
}

// NewLocalStore creates a LocalStore rooted at baseDir, served at baseURL.
func NewLocalStore(baseDir, baseURL string) (*LocalStore, error) {
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, fmt.Errorf("create storage dir: %w", err)
	}
	return &LocalStore{baseDir: baseDir, baseURL: baseURL}, nil
}

// Save writes data from r to a file at the given relative path and returns its public URL.
func (s *LocalStore) Save(relativePath string, r io.Reader) (string, error) {
	dest := filepath.Join(s.baseDir, relativePath)

	// Create parent directories if needed.
	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		return "", fmt.Errorf("create dirs: %w", err)
	}

	f, err := os.Create(dest)
	if err != nil {
		return "", fmt.Errorf("create file: %w", err)
	}
	defer f.Close()

	if _, err := io.Copy(f, r); err != nil {
		return "", fmt.Errorf("write file: %w", err)
	}

	return s.URL(relativePath), nil
}

// Delete removes a file at the given relative path.
func (s *LocalStore) Delete(relativePath string) error {
	dest := filepath.Join(s.baseDir, relativePath)
	if err := os.Remove(dest); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("delete file: %w", err)
	}
	return nil
}

// URL returns the public URL for a file at relativePath.
func (s *LocalStore) URL(relativePath string) string {
	return s.baseURL + "/" + relativePath
}
