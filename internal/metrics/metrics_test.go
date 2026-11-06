package metrics

import (
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/alexandrmotologa/restorecraft/internal/models"
	"github.com/alexandrmotologa/restorecraft/internal/store"
)

func TestGeneratePrometheusMetrics(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "metrics_test.db")

	st, err := store.NewStore(dbPath)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer st.Close()

	// Add sample drill
	d := &models.DrillRun{
		ID:            "metrics-test-01",
		ConfigID:      "test-cfg",
		Name:          "Metrics Test Drill",
		Engine:        "sqlite",
		Status:        models.StatusCompleted,
		StartedAt:     time.Now().Add(-time.Minute),
		HealthScore:   95.0,
		HealthGrade:   models.GradeA,
		RtoSlaSeconds: 60,
		SlaMet:        true,
		Timings: models.PhaseTimings{
			TotalRtoMs: 120,
		},
	}
	if err := st.SaveDrill(d); err != nil {
		t.Fatalf("failed to save drill: %v", err)
	}

	out, err := GeneratePrometheusMetrics(st)
	if err != nil {
		t.Fatalf("GeneratePrometheusMetrics error: %v", err)
	}

	if !strings.Contains(out, "restorecraft_up 1") {
		t.Errorf("expected restorecraft_up metric in output, got: %s", out)
	}
	if !strings.Contains(out, "restorecraft_drills_total{status=\"completed\"} 1") {
		t.Errorf("expected completed drills metric in output, got: %s", out)
	}
	if !strings.Contains(out, "metrics-test-01") {
		t.Errorf("expected drill id in metrics output, got: %s", out)
	}
}
