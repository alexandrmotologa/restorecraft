package auditor

import (
	"context"

	"github.com/alexandrmotologa/restorecraft/internal/config"
	"github.com/alexandrmotologa/restorecraft/internal/engine"
	"github.com/alexandrmotologa/restorecraft/internal/models"
)

// CheckParity compares the restored tables and row counts against configured expectations.
func CheckParity(ctx context.Context, eng engine.Engine, cfg config.ParityConfig) (models.ParityReport, error) {
	report := models.ParityReport{
		Records: []models.TableParityRecord{},
	}

	actualTables, err := eng.GetTableList(ctx)
	if err != nil {
		return report, err
	}

	tableMap := make(map[string]bool)
	for _, t := range actualTables {
		tableMap[t] = true
	}

	// 1. Check expected tables
	for _, expTable := range cfg.ExpectedTables {
		report.VerifiedTables++
		if !tableMap[expTable] {
			report.TotalDiscrepant++
			report.Records = append(report.Records, models.TableParityRecord{
				Table:        expTable,
				MissingTable: true,
				Matched:      false,
			})
			continue
		}

		actualRows, _ := eng.GetTableRowCount(ctx, expTable)
		minRequired, hasMin := cfg.MinRowCounts[expTable]

		matched := true
		if hasMin && actualRows < minRequired {
			matched = false
			report.TotalDiscrepant++
		} else {
			report.MatchedTables++
		}

		report.Records = append(report.Records, models.TableParityRecord{
			Table:         expTable,
			ExpectedRows:  minRequired,
			ActualRows:    actualRows,
			Diff:          actualRows - minRequired,
			Matched:       matched,
			MissingTable:  false,
			SchemaMatched: true,
		})
	}

	// 2. Also record any remaining tables configured with min_row_counts
	for tbl, minRequired := range cfg.MinRowCounts {
		// Skip if already checked in ExpectedTables
		alreadyChecked := false
		for _, exp := range cfg.ExpectedTables {
			if exp == tbl {
				alreadyChecked = true
				break
			}
		}
		if alreadyChecked {
			continue
		}

		report.VerifiedTables++
		if !tableMap[tbl] {
			report.TotalDiscrepant++
			report.Records = append(report.Records, models.TableParityRecord{
				Table:        tbl,
				MissingTable: true,
				Matched:      false,
			})
			continue
		}

		actualRows, _ := eng.GetTableRowCount(ctx, tbl)
		matched := actualRows >= minRequired
		if matched {
			report.MatchedTables++
		} else {
			report.TotalDiscrepant++
		}

		report.Records = append(report.Records, models.TableParityRecord{
			Table:         tbl,
			ExpectedRows:  minRequired,
			ActualRows:    actualRows,
			Diff:          actualRows - minRequired,
			Matched:       matched,
			MissingTable:  false,
			SchemaMatched: true,
		})
	}

	return report, nil
}
