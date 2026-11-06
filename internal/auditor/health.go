package auditor

import (
	"math"

	"github.com/alexandrmotologa/restorecraft/internal/models"
)

// CalculateHealthScore computes the health grade and score from audit results and RTO metrics.
func CalculateHealthScore(
	restoreFailed bool,
	structuralErrors []string,
	constraintViolations []models.ConstraintViolation,
	parity models.ParityReport,
	totalRtoMs int64,
	rtoSlaSeconds int,
) (float64, models.HealthGrade) {
	if restoreFailed {
		return 0.0, models.GradeF
	}

	score := 100.0

	// Deduct for structural/page integrity errors (damaged pages, bit-rot, corruption)
	if len(structuralErrors) > 0 {
		// Catastrophic penalty for damaged pages or checksum errors
		score -= float64(len(structuralErrors)) * 55.0
	}

	// Deduct for relational foreign key violations (20 points each, capped at 50)
	fkPenalty := float64(len(constraintViolations)) * 20.0
	if fkPenalty > 50.0 {
		fkPenalty = 50.0
	}
	score -= fkPenalty

	// Deduct for parity discrepancies
	for _, rec := range parity.Records {
		if rec.MissingTable {
			score -= 25.0
		} else if !rec.Matched {
			score -= 10.0
		}
	}

	// Deduct for RTO SLA overruns
	if rtoSlaSeconds > 0 {
		slaMs := int64(rtoSlaSeconds) * 1000
		if totalRtoMs > slaMs {
			overrunRatio := float64(totalRtoMs-slaMs) / float64(slaMs)
			rtoPenalty := overrunRatio * 25.0
			if rtoPenalty > 30.0 {
				rtoPenalty = 30.0
			}
			score -= rtoPenalty
		}
	}

	if score < 0.0 {
		score = 0.0
	}
	score = math.Round(score*10) / 10

	var grade models.HealthGrade
	switch {
	case score >= 95.0:
		grade = models.GradeAPlus
	case score >= 85.0:
		grade = models.GradeA
	case score >= 70.0:
		grade = models.GradeB
	case score >= 50.0:
		grade = models.GradeC
	default:
		grade = models.GradeF
	}

	return score, grade
}
