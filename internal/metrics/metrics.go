package metrics

import (
	"fmt"
	"strings"

	"github.com/alexandrmotologa/restorecraft/internal/store"
)

// GeneratePrometheusMetrics queries current store metrics and renders them in Prometheus exposition format.
func GeneratePrometheusMetrics(st *store.Store) (string, error) {
	status, err := st.GetSystemStatus()
	if err != nil {
		return "", err
	}

	drills, err := st.ListDrills(100)
	if err != nil {
		return "", err
	}

	var sb strings.Builder

	// Header and Help comments
	sb.WriteString("# HELP restorecraft_up Indicates if restorecraft daemon is running.\n")
	sb.WriteString("# TYPE restorecraft_up gauge\n")
	sb.WriteString("restorecraft_up 1\n\n")

	sb.WriteString("# HELP restorecraft_drills_total Total disaster recovery drills executed.\n")
	sb.WriteString("# TYPE restorecraft_drills_total counter\n")
	sb.WriteString(fmt.Sprintf("restorecraft_drills_total{status=\"completed\"} %d\n", status.SuccessfulDrills))
	sb.WriteString(fmt.Sprintf("restorecraft_drills_total{status=\"failed\"} %d\n\n", status.FailedDrills))

	sb.WriteString("# HELP restorecraft_average_rto_ms Average measured Recovery Time Objective in milliseconds.\n")
	sb.WriteString("# TYPE restorecraft_average_rto_ms gauge\n")
	sb.WriteString(fmt.Sprintf("restorecraft_average_rto_ms %d\n\n", status.AverageRtoMs))

	sb.WriteString("# HELP restorecraft_latest_health_score Latest recovery health grade score (0-100).\n")
	sb.WriteString("# TYPE restorecraft_latest_health_score gauge\n")
	sb.WriteString(fmt.Sprintf("restorecraft_latest_health_score %.1f\n\n", status.LatestScore))

	// Per-drill metrics for latest runs
	sb.WriteString("# HELP restorecraft_rto_duration_ms Recovery Time Objective per drill execution in milliseconds.\n")
	sb.WriteString("# TYPE restorecraft_rto_duration_ms gauge\n")
	for _, d := range drills {
		sb.WriteString(fmt.Sprintf("restorecraft_rto_duration_ms{drill_id=\"%s\",engine=\"%s\",status=\"%s\"} %d\n",
			escapeLabel(d.ID), escapeLabel(d.Engine), d.Status, d.Timings.TotalRtoMs))
	}
	sb.WriteString("\n")

	sb.WriteString("# HELP restorecraft_health_score Health score per drill execution (0-100).\n")
	sb.WriteString("# TYPE restorecraft_health_score gauge\n")
	for _, d := range drills {
		sb.WriteString(fmt.Sprintf("restorecraft_health_score{drill_id=\"%s\",engine=\"%s\",grade=\"%s\"} %.1f\n",
			escapeLabel(d.ID), escapeLabel(d.Engine), d.HealthGrade, d.HealthScore))
	}
	sb.WriteString("\n")

	sb.WriteString("# HELP restorecraft_constraint_violations Relational foreign key violations detected in drill.\n")
	sb.WriteString("# TYPE restorecraft_constraint_violations gauge\n")
	for _, d := range drills {
		count := len(d.Audit.ConstraintViolations)
		sb.WriteString(fmt.Sprintf("restorecraft_constraint_violations{drill_id=\"%s\",engine=\"%s\"} %d\n",
			escapeLabel(d.ID), escapeLabel(d.Engine), count))
	}

	return sb.String(), nil
}

func escapeLabel(v string) string {
	r := strings.ReplaceAll(v, `\`, `\\`)
	r = strings.ReplaceAll(r, `"`, `\"`)
	r = strings.ReplaceAll(r, "\n", `\n`)
	return r
}
