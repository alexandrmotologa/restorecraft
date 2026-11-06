package auditor

import (
	"testing"

	"github.com/alexandrmotologa/restorecraft/internal/models"
)

func TestCalculateHealthScoreClean(t *testing.T) {
	parity := models.ParityReport{
		VerifiedTables: 3,
		MatchedTables:  3,
	}

	score, grade := CalculateHealthScore(false, nil, nil, parity, 500, 60)
	if score != 100.0 {
		t.Fatalf("expected score 100.0, got %f", score)
	}
	if grade != models.GradeAPlus {
		t.Fatalf("expected grade A+, got %s", grade)
	}
}

func TestCalculateHealthScoreWithViolations(t *testing.T) {
	parity := models.ParityReport{
		Records: []models.TableParityRecord{
			{Table: "users", Matched: false},
		},
	}
	violations := []models.ConstraintViolation{
		{Table: "orders", Constraint: "fk_user"},
	}

	score, grade := CalculateHealthScore(false, nil, violations, parity, 500, 60)
	// 100 - 20 (violation) - 10 (parity) = 70.0 -> Grade B
	if score != 70.0 {
		t.Fatalf("expected score 70.0, got %f", score)
	}
	if grade != models.GradeB {
		t.Fatalf("expected grade B, got %s", grade)
	}
}

func TestCalculateHealthScoreFailed(t *testing.T) {
	score, grade := CalculateHealthScore(true, nil, nil, models.ParityReport{}, 500, 60)
	if score != 0.0 || grade != models.GradeF {
		t.Fatalf("expected 0.0 and Grade F on restore failure")
	}
}

func TestCompareSchemas(t *testing.T) {
	actual := map[string]models.TableSchema{
		"users": {
			TableName: "users",
			Columns: []models.ColumnDefinition{
				{Name: "id", DataType: "INTEGER", Nullable: false, IsPrimaryKey: true},
				{Name: "email", DataType: "TEXT", Nullable: false},
			},
		},
	}

	expectedMatch := map[string]models.TableSchema{
		"users": {
			TableName: "users",
			Columns: []models.ColumnDefinition{
				{Name: "id", DataType: "INTEGER"},
				{Name: "email", DataType: "TEXT"},
			},
		},
	}

	diffMatch := CompareSchemas(actual, expectedMatch)
	if !diffMatch.Matched {
		t.Fatalf("expected schemas to match")
	}

	expectedMissingCol := map[string]models.TableSchema{
		"users": {
			TableName: "users",
			Columns: []models.ColumnDefinition{
				{Name: "id", DataType: "INTEGER"},
				{Name: "phone", DataType: "TEXT"},
			},
		},
	}

	diffMismatch := CompareSchemas(actual, expectedMissingCol)
	if diffMismatch.Matched {
		t.Fatalf("expected schema mismatch due to missing column 'phone'")
	}
	if len(diffMismatch.MissingColumns["users"]) == 0 {
		t.Fatalf("expected missing column 'phone' to be reported")
	}
}
