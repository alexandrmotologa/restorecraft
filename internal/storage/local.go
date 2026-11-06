package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// LocalProvider serves snapshots stored on local disk.
type LocalProvider struct {
	baseDir string
}

// NewLocalProvider creates a new local storage provider.
func NewLocalProvider(baseDir string) *LocalProvider {
	return &LocalProvider{baseDir: baseDir}
}

func (l *LocalProvider) resolvePath(key string) string {
	if filepath.IsAbs(key) {
		return key
	}
	// If key exists directly on disk relative to working directory, prioritize it
	if _, err := os.Stat(key); err == nil {
		return key
	}
	if l.baseDir == "" {
		return key
	}
	return filepath.Join(l.baseDir, key)
}

func (l *LocalProvider) Fetch(ctx context.Context, key string) (io.ReadCloser, int64, error) {
	fullPath := l.resolvePath(key)
	info, err := os.Stat(fullPath)
	if err != nil {
		return nil, 0, fmt.Errorf("local snapshot stat failed: %w", err)
	}

	file, err := os.Open(fullPath)
	if err != nil {
		return nil, 0, fmt.Errorf("open local snapshot failed: %w", err)
	}

	return file, info.Size(), nil
}

func (l *LocalProvider) Exists(ctx context.Context, key string) (bool, error) {
	fullPath := l.resolvePath(key)
	_, err := os.Stat(fullPath)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func (l *LocalProvider) Type() string {
	return "local"
}
