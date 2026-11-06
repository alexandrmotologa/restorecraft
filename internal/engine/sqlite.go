package engine

import (
	"bufio"
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/alexandrmotologa/restorecraft/internal/models"
	_ "modernc.org/sqlite"
)

// SQLiteEngine implements database restoration and auditing for SQLite.
type SQLiteEngine struct {
	dbPath string
	db     *sql.DB
}

// NewSQLiteEngine creates a new SQLite restore engine.
func NewSQLiteEngine(dbPath string) *SQLiteEngine {
	return &SQLiteEngine{
		dbPath: dbPath,
	}
}

// RestoreSnapshot restores either a raw SQLite binary database or a SQL text script.
func (s *SQLiteEngine) RestoreSnapshot(ctx context.Context, snapshotStream io.Reader, filename string) error {
	bufReader := bufio.NewReader(snapshotStream)
	magicHeader, _ := bufReader.Peek(16)

	// Check if this is a raw SQLite file binary header
	if bytes.HasPrefix(magicHeader, []byte("SQLite format 3")) {
		outFile, err := os.OpenFile(s.dbPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
		if err != nil {
			return fmt.Errorf("create target sqlite file: %w", err)
		}
		defer outFile.Close()

		_, err = io.Copy(outFile, bufReader)
		if err != nil {
			return fmt.Errorf("write sqlite snapshot binary: %w", err)
		}
	} else {
		// Treat as SQL statement stream
		// Connect first to ensure empty database is initialized
		connStr := fmt.Sprintf("%s?_pragma=foreign_keys(ON)&_pragma=busy_timeout(5000)", s.dbPath)
		db, err := sql.Open("sqlite", connStr)
		if err != nil {
			return fmt.Errorf("open sqlite target: %w", err)
		}
		defer db.Close()

		// Read SQL content
		sqlBytes, err := io.ReadAll(bufReader)
		if err != nil {
			return fmt.Errorf("read sql dump stream: %w", err)
		}

		if _, err := db.ExecContext(ctx, string(sqlBytes)); err != nil {
			return fmt.Errorf("execute sql restore script: %w", err)
		}
	}

	// Open persistent handle for auditing
	connStr := fmt.Sprintf("%s?_pragma=foreign_keys(ON)&_pragma=busy_timeout(5000)", s.dbPath)
	db, err := sql.Open("sqlite", connStr)
	if err != nil {
		return fmt.Errorf("open restored database: %w", err)
	}
	s.db = db

	return nil
}

func (s *SQLiteEngine) GetTableList(ctx context.Context) ([]string, error) {
	if s.db == nil {
		return nil, fmt.Errorf("database handle not open")
	}

	query := `SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%' ORDER BY name;`
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		tables = append(tables, name)
	}
	return tables, rows.Err()
}

func (s *SQLiteEngine) GetTableRowCount(ctx context.Context, tableName string) (int64, error) {
	if s.db == nil {
		return 0, fmt.Errorf("database handle not open")
	}

	// Sanitize table name to prevent SQL injection in audit queries
	cleanName := strings.ReplaceAll(tableName, `"`, `""`)
	query := fmt.Sprintf(`SELECT COUNT(*) FROM "%s";`, cleanName)

	var count int64
	err := s.db.QueryRowContext(ctx, query).Scan(&count)
	return count, err
}

func (s *SQLiteEngine) CheckConstraints(ctx context.Context) ([]models.ConstraintViolation, error) {
	if s.db == nil {
		return nil, fmt.Errorf("database handle not open")
	}

	// PRAGMA foreign_key_check returns (table, rowid, parent_table, fkid)
	rows, err := s.db.QueryContext(ctx, `PRAGMA foreign_key_check;`)
	if err != nil {
		return nil, fmt.Errorf("pragma foreign_key_check: %w", err)
	}
	defer rows.Close()

	var violations []models.ConstraintViolation
	for rows.Next() {
		var table, parent string
		var rowid int64
		var fkid int
		if err := rows.Scan(&table, &rowid, &parent, &fkid); err != nil {
			return nil, err
		}
		violations = append(violations, models.ConstraintViolation{
			Table:       table,
			Constraint:  fmt.Sprintf("foreign_key_to_%s", parent),
			Details:     fmt.Sprintf("Orphan row at rowid %d references missing row in %s", rowid, parent),
			OrphanCount: 1,
		})
	}

	return violations, rows.Err()
}

func (s *SQLiteEngine) CheckIntegrity(ctx context.Context) ([]string, error) {
	if s.db == nil {
		return nil, fmt.Errorf("database handle not open")
	}

	rows, err := s.db.QueryContext(ctx, `PRAGMA integrity_check;`)
	if err != nil {
		return []string{fmt.Sprintf("integrity check failed: %v", err)}, nil
	}
	defer rows.Close()

	var errors []string
	for rows.Next() {
		var msg string
		if err := rows.Scan(&msg); err != nil {
			return nil, err
		}
		if msg != "ok" {
			errors = append(errors, msg)
		}
	}
	return errors, rows.Err()
}

// GetDB returns the active database connection for direct auditing queries.
func (s *SQLiteEngine) GetDB() *sql.DB {
	return s.db
}

func (s *SQLiteEngine) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}
