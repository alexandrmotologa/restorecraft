package auditor

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/alexandrmotologa/restorecraft/internal/engine"
	"github.com/alexandrmotologa/restorecraft/internal/models"
)

var (
	// Standard email pattern
	emailRegex = regexp.MustCompile(`(?i)\b[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}\b`)
	ccRegex    = regexp.MustCompile(`\b(?:\d{4}[ -]?){3}\d{4}\b`)
	ssnRegex   = regexp.MustCompile(`\b\d{3}-\d{2}-\d{4}\b`)
	phoneRegex = regexp.MustCompile(`\b(?:\+?1[-. ]?)?\(?\d{3}\)?[-. ]?\d{3}[-. ]?\d{4}\b`)
)

func isNonTestEmail(email string) bool {
	lower := strings.ToLower(email)
	if strings.Contains(lower, "@example.com") || strings.Contains(lower, "@test.com") || strings.Contains(lower, "@mock.local") {
		return false
	}
	return true
}

// AuditPiiSanitization scans string columns in SQLite databases to assert PII has been masked.
func AuditPiiSanitization(ctx context.Context, eng engine.Engine) (models.PiiAuditReport, error) {
	report := models.PiiAuditReport{
		SanitizedOk: true,
		Checked:     true,
		Findings:    []models.PiiFinding{},
	}

	sqliteEng, ok := eng.(*engine.SQLiteEngine)
	if !ok {
		return report, nil
	}

	db := sqliteEng.GetDB()
	if db == nil {
		return report, nil
	}

	tables, err := eng.GetTableList(ctx)
	if err != nil {
		return report, err
	}

	for _, tbl := range tables {
		cleanTbl := strings.ReplaceAll(tbl, `"`, `""`)
		// Get text columns
		infoQuery := fmt.Sprintf(`PRAGMA table_info("%s");`, cleanTbl)
		rows, err := db.QueryContext(ctx, infoQuery)
		if err != nil {
			continue
		}

		var textCols []string
		for rows.Next() {
			var cid, notNull, pk int
			var name, colType string
			var dflt any
			if err := rows.Scan(&cid, &name, &colType, &notNull, &dflt, &pk); err == nil {
				upperType := strings.ToUpper(colType)
				if strings.Contains(upperType, "TEXT") || strings.Contains(upperType, "CHAR") || strings.Contains(upperType, "VARCHAR") || upperType == "" {
					textCols = append(textCols, name)
				}
			}
		}
		rows.Close()

		// Sample up to 50 rows per table to check for unmasked data
		for _, col := range textCols {
			cleanCol := strings.ReplaceAll(col, `"`, `""`)
			sampleQuery := fmt.Sprintf(`SELECT "%s" FROM "%s" WHERE "%s" IS NOT NULL LIMIT 50;`, cleanCol, cleanTbl, cleanCol)
			sRows, err := db.QueryContext(ctx, sampleQuery)
			if err != nil {
				continue
			}

			emailHits := 0
			ccHits := 0
			ssnHits := 0
			phoneHits := 0

			for sRows.Next() {
				var val string
				if err := sRows.Scan(&val); err == nil {
					if emailRegex.MatchString(val) && isNonTestEmail(val) {
						emailHits++
					}
					if ccRegex.MatchString(val) {
						ccHits++
					}
					if ssnRegex.MatchString(val) {
						ssnHits++
					}
					if phoneRegex.MatchString(val) {
						phoneHits++
					}
				}
			}
			sRows.Close()

			if emailHits > 0 {
				report.SanitizedOk = false
				report.Findings = append(report.Findings, models.PiiFinding{
					Table:       tbl,
					Column:      col,
					PiiType:     "email",
					SampleCount: emailHits,
					Description: fmt.Sprintf("Found %d unmasked non-test email addresses", emailHits),
				})
			}
			if ccHits > 0 {
				report.SanitizedOk = false
				report.Findings = append(report.Findings, models.PiiFinding{
					Table:       tbl,
					Column:      col,
					PiiType:     "credit_card",
					SampleCount: ccHits,
					Description: fmt.Sprintf("Found %d potential unmasked credit card numbers", ccHits),
				})
			}
			if ssnHits > 0 {
				report.SanitizedOk = false
				report.Findings = append(report.Findings, models.PiiFinding{
					Table:       tbl,
					Column:      col,
					PiiType:     "ssn",
					SampleCount: ssnHits,
					Description: fmt.Sprintf("Found %d potential SSN/tax identifier numbers", ssnHits),
				})
			}
			if phoneHits > 0 {
				report.SanitizedOk = false
				report.Findings = append(report.Findings, models.PiiFinding{
					Table:       tbl,
					Column:      col,
					PiiType:     "phone",
					SampleCount: phoneHits,
					Description: fmt.Sprintf("Found %d unmasked phone numbers", phoneHits),
				})
			}
		}
	}

	return report, nil
}
