package orchestrator

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/alexandrmotologa/restorecraft/internal/config"
	"github.com/alexandrmotologa/restorecraft/internal/models"
)

// BisectPoint records the result of testing a single chronological snapshot.
type BisectPoint struct {
	SnapshotPath string             `json:"snapshot_path"`
	HealthGrade  models.HealthGrade `json:"health_grade"`
	HealthScore  float64            `json:"health_score"`
	Status       models.DrillStatus `json:"status"`
	RtoMs        int64              `json:"rto_ms"`
	IntegrityOk  bool               `json:"integrity_ok"`
	ErrorMessage string             `json:"error_message,omitempty"`
}

// BisectReport aggregates the time-travel recovery findings across historical snapshots.
type BisectReport struct {
	TotalSnapshots       int           `json:"total_snapshots"`
	TestedSnapshots      int           `json:"tested_snapshots"`
	FirstBrokenSnapshot  string        `json:"first_broken_snapshot,omitempty"`
	FirstFailureReason   string        `json:"first_failure_reason,omitempty"`
	Timeline             []BisectPoint `json:"timeline"`
	HealthyCount         int           `json:"healthy_count"`
	BrokenCount          int           `json:"broken_count"`
}

// RunTimeTravelBisect discovers snapshots in a folder and tests them chronologically to pinpoint regression.
func (o *Orchestrator) RunTimeTravelBisect(ctx context.Context, dirPath string, engine string) (BisectReport, error) {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return BisectReport{}, fmt.Errorf("read snapshot directory: %w", err)
	}

	var snapshots []string
	for _, e := range entries {
		if !e.IsDir() {
			name := e.Name()
			lower := strings.ToLower(name)
			if strings.HasSuffix(lower, ".sqlite") || strings.HasSuffix(lower, ".db") ||
				strings.HasSuffix(lower, ".sql") || strings.HasSuffix(lower, ".gz") || strings.HasSuffix(lower, ".dump") {
				snapshots = append(snapshots, filepath.Join(dirPath, name))
			}
		}
	}

	sort.Strings(snapshots)

	report := BisectReport{
		TotalSnapshots: len(snapshots),
		Timeline:       []BisectPoint{},
	}

	if len(snapshots) == 0 {
		return report, fmt.Errorf("no valid snapshot files found in %s", dirPath)
	}

	for _, snap := range snapshots {
		drillCfg := config.DrillConfig{
			ID:            fmt.Sprintf("bisect-%s", filepath.Base(snap)),
			Name:          fmt.Sprintf("Bisect Drill (%s)", filepath.Base(snap)),
			Engine:        engine,
			StorageType:   "local",
			SnapshotPath:  snap,
			RtoSlaSeconds: 60,
		}

		drillRun, err := o.RunDrill(ctx, drillCfg)

		point := BisectPoint{
			SnapshotPath: snap,
		}

		if err != nil {
			point.Status = models.StatusFailed
			point.HealthGrade = models.GradeF
			point.HealthScore = 0.0
			point.ErrorMessage = err.Error()
			point.IntegrityOk = false
			report.BrokenCount++
			if report.FirstBrokenSnapshot == "" {
				report.FirstBrokenSnapshot = snap
				report.FirstFailureReason = err.Error()
			}
		} else {
			point.Status = drillRun.Status
			point.HealthGrade = drillRun.HealthGrade
			point.HealthScore = drillRun.HealthScore
			point.RtoMs = drillRun.Timings.TotalRtoMs
			point.IntegrityOk = drillRun.Audit.IntegrityOk
			point.ErrorMessage = drillRun.ErrorMessage

			if drillRun.HealthGrade == models.GradeF || !drillRun.Audit.IntegrityOk {
				report.BrokenCount++
				if report.FirstBrokenSnapshot == "" {
					report.FirstBrokenSnapshot = snap
					report.FirstFailureReason = fmt.Sprintf("Health Grade %s: %s", drillRun.HealthGrade, drillRun.ErrorMessage)
				}
			} else {
				report.HealthyCount++
			}
		}

		report.Timeline = append(report.Timeline, point)
		report.TestedSnapshots++
	}

	return report, nil
}
