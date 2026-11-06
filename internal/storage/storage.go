package storage

import (
	"context"
	"io"
)

// SnapshotInfo contains metadata regarding an acquired backup snapshot.
type SnapshotInfo struct {
	Path        string
	Size        int64
	Sha256      string
	Compression string // "none", "gzip", "zstd"
}

// Provider represents a snapshot storage backend.
type Provider interface {
	// Fetch opens a readable stream for the snapshot at the given key/path.
	Fetch(ctx context.Context, key string) (io.ReadCloser, int64, error)
	// Exists checks whether a snapshot key exists.
	Exists(ctx context.Context, key string) (bool, error)
	// Type returns the provider name (e.g. "local", "s3").
	Type() string
}
