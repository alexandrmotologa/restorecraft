package engine

import (
	"context"
	"io"

	"github.com/alexandrmotologa/restorecraft/internal/models"
)

// Engine defines the contract for restoring and auditing a specific database dialect.
type Engine interface {
	// RestoreSnapshot streams snapshot data into the ephemeral database.
	RestoreSnapshot(ctx context.Context, snapshotStream io.Reader, filename string) error
	// GetTableList returns the names of all user tables in the restored database.
	GetTableList(ctx context.Context) ([]string, error)
	// GetTableRowCount returns the exact row count for a given table.
	GetTableRowCount(ctx context.Context, tableName string) (int64, error)
	// CheckConstraints performs deep relational foreign key verification.
	CheckConstraints(ctx context.Context) ([]models.ConstraintViolation, error)
	// CheckIntegrity performs database page and structural integrity diagnostics.
	CheckIntegrity(ctx context.Context) ([]string, error)
	// Close terminates the engine connection.
	Close() error
}
