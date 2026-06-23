package storage

import (
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

// AllowedImageExts contains the file extensions accepted for diagram uploads.
var AllowedImageExts = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".gif":  true,
	".svg":  true,
	".webp": true,
}

// MaxImageSize is the maximum accepted upload size (5 MB).
const MaxImageSize = 5 << 20

// ImageStore manages diagram/image uploads backed by a LocalStore.
type ImageStore struct {
	local *LocalStore
}

// NewImageStore creates an ImageStore using the given local store.
func NewImageStore(local *LocalStore) *ImageStore {
	return &ImageStore{local: local}
}

// SaveDiagram stores an uploaded diagram file and returns its public URL.
// It validates the file extension and generates a unique filename.
func (s *ImageStore) SaveDiagram(file multipart.File, originalName string) (string, error) {
	ext := strings.ToLower(filepath.Ext(originalName))
	if !AllowedImageExts[ext] {
		return "", fmt.Errorf("unsupported image type: %s", ext)
	}

	// Generate a unique filename to avoid collisions.
	uniqueName := uuid.New().String() + ext
	relativePath := "diagrams/" + uniqueName

	url, err := s.local.Save(relativePath, io.LimitReader(file, MaxImageSize))
	if err != nil {
		return "", fmt.Errorf("save diagram: %w", err)
	}

	return url, nil
}

// DeleteDiagram removes a stored diagram given its public URL.
// If the URL doesn't match the local store prefix, it is silently ignored.
func (s *ImageStore) DeleteDiagram(publicURL string) error {
	// Strip the base URL prefix to get the relative path.
	relativePath := strings.TrimPrefix(publicURL, s.local.baseURL+"/")
	if relativePath == publicURL {
		// URL does not belong to local storage; skip deletion.
		return nil
	}
	return s.local.Delete(relativePath)
}
