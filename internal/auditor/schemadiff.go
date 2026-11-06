package auditor

import (
	"context"
	"fmt"
	"strings"

	"github.com/alexandrmotologa/restorecraft/internal/engine"
	"github.com/alexandrmotologa/restorecraft/internal/models"
)

// InspectSQLiteSchema extracts detailed column schemas from an SQLite database.
func InspectSQLiteSchema(ctx context.Context, eng engine.Engine) (map[string]models.TableSchema, error) {
	tables, err := eng.GetTableList(ctx)
	if err != nil {
		return nil, err
	}

	result := make(map[string]models.TableSchema)
	sqliteEng, ok := eng.(*engine.SQLiteEngine)
	if !ok {
		// Generic fallback
		for _, tbl := range tables {
			result[tbl] = models.TableSchema{TableName: tbl}
		}
		return result, nil
	}

	for _, tbl := range tables {
		clean := strings.ReplaceAll(tbl, `"`, `""`)
		query := fmt.Sprintf(`PRAGMA table_info("%s");`, clean)
		db := sqliteEng.GetDB()
		if db == nil {
			continue
		}

		rows, err := db.QueryContext(ctx, query)
		if err != nil {
			continue
		}

		schema := models.TableSchema{TableName: tbl}
		for rows.Next() {
			var cid int
			var name, colType string
			var notNull, pk int
			var dflt any
			if err := rows.Scan(&cid, &name, &colType, &notNull, &dflt, &pk); err == nil {
				schema.Columns = append(schema.Columns, models.ColumnDefinition{
					Name:         name,
					DataType:     colType,
					Nullable:     notNull == 0,
					IsPrimaryKey: pk > 0,
				})
			}
		}
		rows.Close()
		result[tbl] = schema
	}

	return result, nil
}

// CompareSchemas evaluates actual schema against an expected baseline.
func CompareSchemas(actual map[string]models.TableSchema, expected map[string]models.TableSchema) models.SchemaDiffReport {
	report := models.SchemaDiffReport{
		Matched:        true,
		MissingTables:  []string{},
		UnexpectedTables: []string{},
		MissingColumns: make(map[string][]string),
		AlteredColumns: make(map[string][]string),
	}

	if len(expected) == 0 {
		return report
	}

	// 1. Check for missing tables
	for expTable, expSchema := range expected {
		actSchema, exists := actual[expTable]
		if !exists {
			report.MissingTables = append(report.MissingTables, expTable)
			report.Matched = false
			continue
		}

		// Check columns
		actColMap := make(map[string]models.ColumnDefinition)
		for _, col := range actSchema.Columns {
			actColMap[strings.ToLower(col.Name)] = col
		}

		for _, expCol := range expSchema.Columns {
			actCol, colExists := actColMap[strings.ToLower(expCol.Name)]
			if !colExists {
				report.MissingColumns[expTable] = append(report.MissingColumns[expTable], expCol.Name)
				report.Matched = false
				continue
			}

			// Compare data type if specified
			if expCol.DataType != "" && !strings.EqualFold(expCol.DataType, actCol.DataType) {
				diffMsg := fmt.Sprintf("Column '%s' type mismatch: expected %s, got %s", expCol.Name, expCol.DataType, actCol.DataType)
				report.AlteredColumns[expTable] = append(report.AlteredColumns[expTable], diffMsg)
				report.Matched = false
			}
		}
	}

	return report
}
