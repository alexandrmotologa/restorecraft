package store

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/alexandrmotologa/restorecraft/internal/models"
)

func TestStoreLifecycle(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "restorecraft_store_test_*")
	if err != nil {
		t.Fatalf("temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	s, err := NewStore(tempDir)
	if err != nil {
		t.Fatalf("new store: %v", err)
	}
	defer s.Close()

	now := time.Now().UTC()
	drill := &models.DrillRun{
		ID:             "drill-test-1",
		ConfigID:       "config-1",
		Name:           "Test SQLite Drill",
		Engine:         "sqlite",
		Status:         models.StatusCompleted,
		SnapshotSource: "local",
		SnapshotPath:   "test.sqlite",
		SnapshotHash:   "abc123hash",
		SnapshotBytes:  1024,
		StartedAt:      now,
		CompletedAt:    &now,
		HealthScore:    98.5,
		HealthGrade:    models.GradeAPlus,
		RtoSlaSeconds:  60,
		SlaMet:         true,
		Timings: models.PhaseTimings{
			TotalRtoMs: 450,
		},
		Audit: models.AuditReport{
			IntegrityOk: true,
		},
	}

	if err := s.SaveDrill(drill); err != nil {
		t.Fatalf("save drill: %v", err)
	}

	fetched, err := s.GetDrill("drill-test-1")
	if err != nil {
		t.Fatalf("get drill: %v", err)
	}
	if fetched.Name != drill.Name {
		t.Fatalf("expected name %s, got %s", drill.Name, fetched.Name)
	}
	if fetched.HealthGrade != models.GradeAPlus {
		t.Fatalf("expected grade A+, got %s", fetched.HealthGrade)
	}

	drills, err := s.ListDrills(10)
	if err != nil {
		t.Fatalf("list drills: %v", err)
	}
	if len(drills) != 1 {
		t.Fatalf("expected 1 drill, got %d", len(drills))
	}

	status, err := s.GetSystemStatus()
	if err != nil {
		t.Fatalf("get system status: %v", err)
	}
	if status.TotalDrills != 1 || status.SuccessfulDrills != 1 {
		t.Fatalf("unexpected system status: %+v", status)
	}

	_ = filepath.Clean(tempDir)
}
