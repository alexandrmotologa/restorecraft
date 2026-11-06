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

// Audit runs structural, constraint, parity, schema diff, and PII evaluations on a restored database engine.
func (a *Auditor) Audit(
	ctx context.Context,
	eng engine.Engine,
	parityCfg config.ParityConfig,
	auditPii bool,
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

	// 4. Schema introspection and diff against expected baseline
	actualSchema, err := InspectSQLiteSchema(ctx, eng)
	if err == nil && len(actualSchema) > 0 {
		diff := CompareSchemas(actualSchema, parityCfg.ExpectedSchema)
		report.SchemaDiff = &diff
		if !diff.Matched {
			report.IntegrityOk = false
		}
	}

	// 5. PII Masking & Data Sanitization check
	if auditPii {
		piiRep, err := AuditPiiSanitization(ctx, eng)
		if err == nil {
			report.PiiAudit = &piiRep
			if !piiRep.SanitizedOk {
				report.IntegrityOk = false
			}
		}
	}

	// 6. Calculate health score and letter grade
	score, grade := CalculateHealthScore(
		false,
		report.StructuralErrors,
		report.ConstraintViolations,
		report.Parity,
		totalRtoMs,
		rtoSlaSeconds,
	)

	// Additional deductions for Schema drift and PII violations
	if report.SchemaDiff != nil && !report.SchemaDiff.Matched {
		score -= 15.0
	}
	if report.PiiAudit != nil && !report.PiiAudit.SanitizedOk {
		score -= 20.0
	}
	if score < 0 {
		score = 0
	}

	// Recalculate grade if score decreased
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

	return report, score, grade, nil
}
