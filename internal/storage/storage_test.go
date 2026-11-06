package storage

import (
	"bytes"
	"compress/gzip"
	"context"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestLocalProviderAndHashingReader(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "restorecraft_storage_test_*")
	if err != nil {
		t.Fatalf("temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	content := []byte("SELECT * FROM users;\n")
	testFile := filepath.Join(tempDir, "test.sql")
	if err := os.WriteFile(testFile, content, 0644); err != nil {
		t.Fatalf("write test file: %v", err)
	}

	provider := NewLocalProvider(tempDir)
	exists, err := provider.Exists(context.Background(), "test.sql")
	if err != nil || !exists {
		t.Fatalf("expected test.sql to exist")
	}

	rc, size, err := provider.Fetch(context.Background(), "test.sql")
	if err != nil {
		t.Fatalf("fetch failed: %v", err)
	}
	defer rc.Close()

	if size != int64(len(content)) {
		t.Fatalf("expected size %d, got %d", len(content), size)
	}

	hr := NewHashingReader(rc)
	readBytes, err := io.ReadAll(hr)
	if err != nil {
		t.Fatalf("read all: %v", err)
	}
	if !bytes.Equal(readBytes, content) {
		t.Fatalf("mismatched content")
	}
	if hr.SumHex() == "" {
		t.Fatalf("expected non-empty sha256 checksum")
	}
}

func TestGzipDecompression(t *testing.T) {
	raw := []byte("CREATE TABLE items (id INT);")
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	_, _ = gw.Write(raw)
	_ = gw.Close()

	stream, comp, err := OpenDecompressedStream(&buf, "backup.sql.gz")
	if err != nil {
		t.Fatalf("open decompressed stream: %v", err)
	}
	if comp != "gzip" {
		t.Fatalf("expected gzip, got %s", comp)
	}

	out, err := io.ReadAll(stream)
	if err != nil {
		t.Fatalf("read decompressed: %v", err)
	}
	if !bytes.Equal(out, raw) {
		t.Fatalf("decompressed content mismatch: got %s", string(out))
	}
}
