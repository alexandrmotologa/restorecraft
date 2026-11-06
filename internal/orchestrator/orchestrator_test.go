package orchestrator

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/alexandrmotologa/restorecraft/internal/certificate"
	"github.com/alexandrmotologa/restorecraft/internal/config"
	"github.com/alexandrmotologa/restorecraft/internal/fixtures"
	"github.com/alexandrmotologa/restorecraft/internal/models"
	"github.com/alexandrmotologa/restorecraft/internal/store"
)

func TestOrchestratorEndToEndDrill(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "restorecraft_orch_test_*")
	if err != nil {
		t.Fatalf("temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	fixtureDir := filepath.Join(tempDir, "fixtures")
	if err := fixtures.GenerateAll(fixtureDir); err != nil {
		t.Fatalf("generate fixtures: %v", err)
	}

	st, err := store.NewStore(filepath.Join(tempDir, "data"))
	if err != nil {
		t.Fatalf("store: %v", err)
	}
	defer st.Close()

	cfg := config.DefaultConfig()
	cfg.Security.SigningSecret = "test-authority-hmac-secret-12345678"
	cfg.Storage.Local.BaseDir = fixtureDir

	orch := NewOrchestrator(cfg, st)

	// 1. Run clean drill
	validSnapshot := filepath.Join(fixtureDir, "valid_ecommerce.sqlite")
	cleanDrillCfg := config.DrillConfig{
		ID:            "test-clean-drill",
		Name:          "Clean E-commerce Drill",
		Engine:        "sqlite",
		StorageType:   "local",
		SnapshotPath:  validSnapshot,
		RtoSlaSeconds: 60,
		Parity: config.ParityConfig{
			ExpectedTables: []string{"users", "orders", "order_items"},
			MinRowCounts: map[string]int64{
				"users":  5,
				"orders": 5,
			},
		},
	}

	run, err := orch.RunDrill(context.Background(), cleanDrillCfg)
	if err != nil {
		t.Fatalf("run drill failed: %v", err)
	}

	if run.Status != models.StatusCompleted {
		t.Fatalf("expected completed status, got: %s", run.Status)
	}
	if run.HealthGrade != models.GradeAPlus {
		t.Fatalf("expected grade A+, got: %s (Score: %.1f)", run.HealthGrade, run.HealthScore)
	}
	if !run.Audit.IntegrityOk {
		t.Fatalf("expected integrity ok")
	}
	if run.Certificate == nil {
		t.Fatalf("expected certificate to be issued")
	}

	// Verify certificate signature
	res := certificate.VerifyCertificate(run.Certificate, cfg.Security.SigningSecret)
	if !res.Valid {
		t.Fatalf("expected valid signature, got: %s", res.Reason)
	}

	// 2. Run FK violation drill
	fkSnapshot := filepath.Join(fixtureDir, "fk_violation.sqlite")
	fkDrillCfg := config.DrillConfig{
		ID:            "test-fk-drill",
		Name:          "FK Violation Drill",
		Engine:        "sqlite",
		StorageType:   "local",
		SnapshotPath:  fkSnapshot,
		RtoSlaSeconds: 60,
	}

	fkRun, err := orch.RunDrill(context.Background(), fkDrillCfg)
	if err != nil {
		t.Fatalf("run fk drill failed: %v", err)
	}
	if fkRun.Audit.IntegrityOk {
		t.Fatalf("expected integrity ok to be false on fk violation")
	}
	if len(fkRun.Audit.ConstraintViolations) == 0 {
		t.Fatalf("expected constraint violations to be reported")
	}
	if fkRun.HealthGrade == models.GradeAPlus {
		t.Fatalf("expected health grade to be degraded, got %s", fkRun.HealthGrade)
	}

	// 3. Run corrupt header drill
	corruptSnapshot := filepath.Join(fixtureDir, "corrupt_header.sqlite")
	corruptDrillCfg := config.DrillConfig{
		ID:            "test-corrupt-drill",
		Name:          "Corrupt Header Drill",
		Engine:        "sqlite",
		StorageType:   "local",
		SnapshotPath:  corruptSnapshot,
		RtoSlaSeconds: 60,
	}

	corruptRun, err := orch.RunDrill(context.Background(), corruptDrillCfg)
	// Expect failure or grade F
	if corruptRun.HealthGrade != models.GradeF {
		t.Fatalf("expected grade F for corrupt snapshot, got %s", corruptRun.HealthGrade)
	}
}
