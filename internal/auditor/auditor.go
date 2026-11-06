package auditor

import (
	"context"

	"github.com/alexandrmotologa/restorecraft/internal/config"
	"github.com/alexandrmotologa/restorecraft/internal/engine"
	"github.com/alexandrmotologa/restorecraft/internal/models"
)

// Auditor coordinates relational constraint scans, page integrity checks, and parity comparisons.
type Auditor struct{}

// NewAuditor creates a new Auditor instance.
func NewAuditor() *Auditor {
	return &Auditor{}
}

// Audit runs structural, constraint, and parity evaluations on a restored database engine.
func (a *Auditor) Audit(
	ctx context.Context,
	eng engine.Engine,
	parityCfg config.ParityConfig,
	totalRtoMs int64,
	rtoSlaSeconds int,
) (models.AuditReport, float64, models.HealthGrade, error) {
	report := models.AuditReport{
		IntegrityOk:          true,
		StructuralErrors:     []string{},
		ConstraintViolations: []models.ConstraintViolation{},
		Warnings:             []string{},
	}

	// 1. Structural/Page integrity diagnostics
	structErrors, err := eng.CheckIntegrity(ctx)
	if err != nil {
		report.IntegrityOk = false
		report.StructuralErrors = append(report.StructuralErrors, err.Error())
	} else if len(structErrors) > 0 {
		report.IntegrityOk = false
		report.StructuralErrors = structErrors
	}

	// 2. Relational foreign key constraint checks
	violations, err := eng.CheckConstraints(ctx)
	if err != nil {
		report.Warnings = append(report.Warnings, "constraint check warning: "+err.Error())
	} else {
		report.ConstraintViolations = violations
		if len(violations) > 0 {
			report.IntegrityOk = false
		}
	}

	// 3. Parity checks against baseline expectations
	parityReport, err := CheckParity(ctx, eng, parityCfg)
	if err != nil {
		report.Warnings = append(report.Warnings, "parity check warning: "+err.Error())
	} else {
		report.Parity = parityReport
		if parityReport.TotalDiscrepant > 0 {
			report.IntegrityOk = false
		}
	}

	// 4. Calculate health score and letter grade
	score, grade := CalculateHealthScore(
		false,
		report.StructuralErrors,
		report.ConstraintViolations,
		report.Parity,
		totalRtoMs,
		rtoSlaSeconds,
	)

	return report, score, grade, nil
}
