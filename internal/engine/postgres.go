package engine

import (
	"bufio"
	"context"
	"database/sql"
	"fmt"
	"io"
	"strings"

	"github.com/alexandrmotologa/restorecraft/internal/models"
	"github.com/alexandrmotologa/restorecraft/internal/sandbox"
	_ "github.com/lib/pq"
)

// PostgresEngine implements database restoration and auditing for PostgreSQL.
type PostgresEngine struct {
	connDetails sandbox.ConnectionDetails
	db          *sql.DB
}

// NewPostgresEngine creates a new PostgreSQL restore and verification engine.
func NewPostgresEngine(details sandbox.ConnectionDetails) *PostgresEngine {
	return &PostgresEngine{
		connDetails: details,
	}
}

func (p *PostgresEngine) getConnString() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable connect_timeout=10",
		p.connDetails.Host,
		p.connDetails.Port,
		p.connDetails.User,
		p.connDetails.Password,
		p.connDetails.DatabaseName,
	)
}

func (p *PostgresEngine) connect() error {
	if p.db != nil {
		return nil
	}
	db, err := sql.Open("postgres", p.getConnString())
	if err != nil {
		return err
	}
	p.db = db
	return nil
}

func (p *PostgresEngine) RestoreSnapshot(ctx context.Context, snapshotStream io.Reader, filename string) error {
	if err := p.connect(); err != nil {
		return fmt.Errorf("connect to postgres: %w", err)
	}

	// Read and execute SQL statements
	scanner := bufio.NewScanner(snapshotStream)
	// Allow up to 16MB buffer for large insert statements
	buf := make([]byte, 64*1024)
	scanner.Buffer(buf, 16*1024*1024)

	var sb strings.Builder
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "--") || trimmed == "" {
			continue
		}

		sb.WriteString(line)
		sb.WriteString("\n")

		if strings.HasSuffix(trimmed, ";") {
			stmt := sb.String()
			sb.Reset()

			if _, err := p.db.ExecContext(ctx, stmt); err != nil {
				return fmt.Errorf("execute postgres restore sql: %w", err)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("scan postgres dump: %w", err)
	}

	// Flush any trailing statement
	if remaining := strings.TrimSpace(sb.String()); remaining != "" {
		if _, err := p.db.ExecContext(ctx, remaining); err != nil {
			return fmt.Errorf("execute trailing postgres statement: %w", err)
		}
	}

	return nil
}

func (p *PostgresEngine) GetTableList(ctx context.Context) ([]string, error) {
	if err := p.connect(); err != nil {
		return nil, err
	}

	query := `
	SELECT table_name
	FROM information_schema.tables
	WHERE table_schema = 'public' AND table_type = 'BASE TABLE'
	ORDER BY table_name;
	`
	rows, err := p.db.QueryContext(ctx, query)
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

func (p *PostgresEngine) GetTableRowCount(ctx context.Context, tableName string) (int64, error) {
	if err := p.connect(); err != nil {
		return 0, err
	}

	clean := strings.ReplaceAll(tableName, `"`, `""`)
	query := fmt.Sprintf(`SELECT COUNT(*) FROM "%s";`, clean)

	var count int64
	err := p.db.QueryRowContext(ctx, query).Scan(&count)
	return count, err
}

func (p *PostgresEngine) CheckConstraints(ctx context.Context) ([]models.ConstraintViolation, error) {
	if err := p.connect(); err != nil {
		return nil, err
	}

	// Query unvalidated constraints in pg_constraint
	query := `
	SELECT 
		conrelid::regclass::text AS table_name,
		conname AS constraint_name
	FROM pg_constraint
	WHERE NOT convalidated AND contype = 'f';
	`
	rows, err := p.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var violations []models.ConstraintViolation
	for rows.Next() {
		var tbl, con string
		if err := rows.Scan(&tbl, &con); err != nil {
			return nil, err
		}
		violations = append(violations, models.ConstraintViolation{
			Table:      tbl,
			Constraint: con,
			Details:    "Foreign key constraint is marked NOT convalidated",
		})
	}
	return violations, rows.Err()
}

func (p *PostgresEngine) CheckIntegrity(ctx context.Context) ([]string, error) {
	if err := p.connect(); err != nil {
		return nil, err
	}

	// Validate that tables are accessible and have valid page layouts
	tables, err := p.GetTableList(ctx)
	if err != nil {
		return nil, err
	}

	var errors []string
	for _, tbl := range tables {
		clean := strings.ReplaceAll(tbl, `"`, `""`)
		_, err := p.db.ExecContext(ctx, fmt.Sprintf(`ANALYZE "%s";`, clean))
		if err != nil {
			errors = append(errors, fmt.Sprintf("ANALYZE table %s failed: %v", tbl, err))
		}
	}
	return errors, nil
}

func (p *PostgresEngine) Close() error {
	if p.db != nil {
		return p.db.Close()
	}
	return nil
}
