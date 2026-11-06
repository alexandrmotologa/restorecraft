package engine

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/alexandrmotologa/restorecraft/internal/fixtures"
)

func TestSQLiteEngineValidRestoreAndParity(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "restorecraft_engine_test_*")
	if err != nil {
		t.Fatalf("temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	if err := fixtures.GenerateAll(tempDir); err != nil {
		t.Fatalf("generate fixtures: %v", err)
	}

	validDbPath := filepath.Join(tempDir, "valid_ecommerce.sqlite")
	targetDbPath := filepath.Join(tempDir, "restored.sqlite")

	eng := NewSQLiteEngine(targetDbPath)
	defer eng.Close()

	file, err := os.Open(validDbPath)
	if err != nil {
		t.Fatalf("open valid db: %v", err)
	}
	defer file.Close()

	ctx := context.Background()
	if err := eng.RestoreSnapshot(ctx, file, "valid_ecommerce.sqlite"); err != nil {
		t.Fatalf("restore snapshot: %v", err)
	}

	tables, err := eng.GetTableList(ctx)
	if err != nil {
		t.Fatalf("get table list: %v", err)
	}

	if len(tables) < 4 {
		t.Fatalf("expected at least 4 tables, got %d (%v)", len(tables), tables)
	}

	userCount, err := eng.GetTableRowCount(ctx, "users")
	if err != nil {
		t.Fatalf("get users row count: %v", err)
	}
	if userCount != 5 {
		t.Fatalf("expected 5 users, got %d", userCount)
	}

	violations, err := eng.CheckConstraints(ctx)
	if err != nil {
		t.Fatalf("check constraints: %v", err)
	}
	if len(violations) != 0 {
		t.Fatalf("expected 0 constraint violations in clean db, got %d", len(violations))
	}

	integrityErrors, err := eng.CheckIntegrity(ctx)
	if err != nil {
		t.Fatalf("check integrity: %v", err)
	}
	if len(integrityErrors) != 0 {
		t.Fatalf("expected 0 integrity errors, got %d (%v)", len(integrityErrors), integrityErrors)
	}
}

func TestSQLiteEngineFkViolationDetection(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "restorecraft_fk_test_*")
	if err != nil {
		t.Fatalf("temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	if err := fixtures.GenerateAll(tempDir); err != nil {
		t.Fatalf("generate fixtures: %v", err)
	}

	fkDbPath := filepath.Join(tempDir, "fk_violation.sqlite")
	targetDbPath := filepath.Join(tempDir, "restored_fk.sqlite")

	eng := NewSQLiteEngine(targetDbPath)
	defer eng.Close()

	file, err := os.Open(fkDbPath)
	if err != nil {
		t.Fatalf("open fk db: %v", err)
	}
	defer file.Close()

	ctx := context.Background()
	if err := eng.RestoreSnapshot(ctx, file, "fk_violation.sqlite"); err != nil {
		t.Fatalf("restore snapshot: %v", err)
	}

	violations, err := eng.CheckConstraints(ctx)
	if err != nil {
		t.Fatalf("check constraints: %v", err)
	}

	if len(violations) == 0 {
		t.Fatalf("expected constraint violation to be detected for orphan invoice")
	}

	found := false
	for _, v := range violations {
		if v.Table == "invoices" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected violation on table 'invoices'")
	}
}
