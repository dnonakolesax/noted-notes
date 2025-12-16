package file

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type FileStore struct {
	base string
	mu   sync.Mutex
}

func New(baseDir string) *FileStore {
	return &FileStore{base: baseDir}
}

func (fs *FileStore) pathFor(id string) string {
	// Минимальная санитаризация id для файловой системы.
	safe := strings.TrimSpace(id)
	safe = strings.ReplaceAll(safe, "..", "")
	safe = strings.ReplaceAll(safe, "/", "_")
	safe = strings.ReplaceAll(safe, "\\", "_")
	if safe == "" {
		safe = "default"
	}
	return filepath.Join(fs.base, safe+".am")
}

func (fs *FileStore) Load(ctx context.Context, id string) ([]byte, error) {
	_ = ctx

	fs.mu.Lock()
	defer fs.mu.Unlock()

	if err := os.MkdirAll(fs.base, 0o755); err != nil {
		return nil, err
	}

	p := fs.pathFor(id)
	b, err := os.ReadFile(p)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, os.ErrNotExist
		}
		return nil, err
	}
	return b, nil
}

func (fs *FileStore) Save(ctx context.Context, id string, data []byte) error {
	_ = ctx

	fs.mu.Lock()
	defer fs.mu.Unlock()

	if err := os.MkdirAll(fs.base, 0o755); err != nil {
		return err
	}

	p := fs.pathFor(id)
	tmp := p + ".tmp"

	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, p)
}
