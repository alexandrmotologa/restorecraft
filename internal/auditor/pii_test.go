package auditor

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/alexandrmotologa/restorecraft/internal/engine"
	"github.com/alexandrmotologa/restorecraft/internal/fixtures"
)

func TestAuditPiiSanitizationCleanDatabase(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "restorecraft_pii_test_*")
	if err != nil {
		t.Fatalf("temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	if err := fixtures.GenerateAll(tempDir); err != nil {
		t.Fatalf("failed to generate fixtures: %v", err)
	}

	validDbPath := filepath.Join(tempDir, "valid_ecommerce.sqlite")
	targetDbPath := filepath.Join(tempDir, "restored.sqlite")

	eng := engine.NewSQLiteEngine(targetDbPath)
	defer eng.Close()

	file, err := os.Open(validDbPath)
	if err != nil {
		t.Fatalf("open valid db: %v", err)
	}
	defer file.Close()

	ctx := context.Background()
	if err := eng.RestoreSnapshot(ctx, file, "valid_ecommerce.sqlite"); err != nil {
		t.Fatalf("restore failed: %v", err)
	}

	report, err := AuditPiiSanitization(ctx, eng)
	if err != nil {
		t.Fatalf("AuditPiiSanitization error: %v", err)
	}

	if !report.Checked {
		t.Errorf("expected report.Checked to be true")
	}
	// valid_ecommerce uses @example.com synthetic emails, so it should be sanitized
	if !report.SanitizedOk {
		t.Errorf("expected sanitized_ok to be true for synthetic fixture, got findings: %+v", report.Findings)
	}
}
