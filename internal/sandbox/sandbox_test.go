package sandbox

import (
	"context"
	"os"
	"testing"
)

func TestNativeSandboxLifecycle(t *testing.T) {
	ctx := context.Background()
	sb := NewNativeSandbox("test_drill.db")

	if err := sb.Provision(ctx); err != nil {
		t.Fatalf("provision failed: %v", err)
	}

	details := sb.ConnectionDetails()
	if details.FilePath == "" {
		t.Fatalf("expected non-empty FilePath")
	}

	// Verify temp directory was created
	dir := sb.tempDir
	if _, err := os.Stat(dir); err != nil {
		t.Fatalf("temp dir does not exist: %v", err)
	}

	if err := sb.Cleanup(ctx); err != nil {
		t.Fatalf("cleanup failed: %v", err)
	}

	// Verify temp directory was removed
	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		t.Fatalf("expected temp dir to be removed")
	}
}
