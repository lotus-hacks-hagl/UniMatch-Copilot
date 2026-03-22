package filestore

import (
	"io"
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

type FileStore interface {
	Save(caseID uuid.UUID, fileName string, reader io.Reader) (string, error)
	GetPath(relativePath string) string
	Delete(relativePath string) error
}

type localFileStore struct {
	basePath string
}

func NewLocalFileStore(basePath string) (FileStore, error) {
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, err
	}
	return &localFileStore{basePath: basePath}, nil
}

func (s *localFileStore) Save(caseID uuid.UUID, fileName string, reader io.Reader) (string, error) {
	// Create directory for case: uploads/<case_id>/<file_name>
	relDir := caseID.String()
	fullDir := filepath.Join(s.basePath, relDir)
	if err := os.MkdirAll(fullDir, 0755); err != nil {
		return "", err
	}

	// Avoid name collision by adding a small prefix if needed, but for simplicity:
	relPath := filepath.Join(relDir, fileName)
	fullPath := filepath.Join(s.basePath, relPath)

	out, err := os.Create(fullPath)
	if err != nil {
		return "", err
	}
	defer out.Close()

	if _, err := io.Copy(out, reader); err != nil {
		return "", err
	}

	return relPath, nil
}

func (s *localFileStore) GetPath(relativePath string) string {
	return filepath.Join(s.basePath, relativePath)
}

func (s *localFileStore) Delete(relativePath string) error {
	fullPath := filepath.Join(s.basePath, relativePath)
	return os.Remove(fullPath)
}
