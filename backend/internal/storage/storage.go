// Package storage abstracts file storage with local-filesystem and S3 backends.
package storage

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/akeelnazir/osint-app/backend/internal/config"
)

// Storage is the interface for saving and retrieving uploaded files.
type Storage interface {
	Save(name string, r io.Reader) (path string, err error)
	Open(path string) (io.ReadCloser, error)
	Delete(path string) error
	URL(path string) string
}

// New selects a storage backend based on config.
func New(cfg *config.Config) (Storage, error) {
	switch cfg.FileStorage {
	case "local":
		return newLocal(cfg.LocalStoragePath)
	case "s3":
		return nil, fmt.Errorf("s3 storage backend not yet implemented; set FILE_STORAGE=local")
	default:
		return newLocal(cfg.LocalStoragePath)
	}
}

type localStorage struct{ root string }

func newLocal(root string) (*localStorage, error) {
	if root == "" {
		root = "./uploads"
	}
	if err := os.MkdirAll(root, 0o755); err != nil {
		return nil, fmt.Errorf("create storage dir: %w", err)
	}
	return &localStorage{root: root}, nil
}

func (l *localStorage) Save(name string, r io.Reader) (string, error) {
	// Place files under YYYY/MM subdirs to avoid huge flat dirs.
	subdir := filepath.Join("uploads")
	full := filepath.Join(l.root, subdir, name)
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		return "", err
	}
	f, err := os.Create(full)
	if err != nil {
		return "", err
	}
	defer f.Close()
	if _, err := io.Copy(f, r); err != nil {
		return "", err
	}
	// Return a logical path (relative to root) for DB storage.
	return filepath.Join(subdir, name), nil
}

func (l *localStorage) Open(path string) (io.ReadCloser, error) {
	return os.Open(filepath.Join(l.root, path))
}

func (l *localStorage) Delete(path string) error {
	full := filepath.Join(l.root, path)
	if err := os.Remove(full); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func (l *localStorage) URL(path string) string {
	return "/files/" + path
}
