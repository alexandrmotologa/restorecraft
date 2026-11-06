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
	_ "github.com/go-sql-driver/mysql"
)

// MySQLEngine implements database restoration and auditing for MySQL/MariaDB.
type MySQLEngine struct {
	connDetails sandbox.ConnectionDetails
	db          *sql.DB
}

// NewMySQLEngine creates a new MySQL restore and verification engine.
func NewMySQLEngine(details sandbox.ConnectionDetails) *MySQLEngine {
	return &MySQLEngine{
		connDetails: details,
	}
}

func (m *MySQLEngine) getDSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?multiStatements=true&timeout=10s",
		m.connDetails.User,
		m.connDetails.Password,
		m.connDetails.Host,
		m.connDetails.Port,
		m.connDetails.DatabaseName,
	)
}

func (m *MySQLEngine) connect() error {
	if m.db != nil {
		return nil
	}
	db, err := sql.Open("mysql", m.getDSN())
	if err != nil {
		return err
	}
	m.db = db
	return nil
}

func (m *MySQLEngine) RestoreSnapshot(ctx context.Context, snapshotStream io.Reader, filename string) error {
	if err := m.connect(); err != nil {
		return fmt.Errorf("connect to mysql: %w", err)
	}

	scanner := bufio.NewScanner(snapshotStream)
	buf := make([]byte, 64*1024)
	scanner.Buffer(buf, 16*1024*1024)

	var sb strings.Builder
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "--") || strings.HasPrefix(trimmed, "/*") || trimmed == "" {
			continue
		}

		sb.WriteString(line)
		sb.WriteString("\n")

		if strings.HasSuffix(trimmed, ";") {
			stmt := sb.String()
			sb.Reset()

			if _, err := m.db.ExecContext(ctx, stmt); err != nil {
				return fmt.Errorf("execute mysql restore sql: %w", err)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("scan mysql dump: %w", err)
	}

	if remaining := strings.TrimSpace(sb.String()); remaining != "" {
		if _, err := m.db.ExecContext(ctx, remaining); err != nil {
			return fmt.Errorf("execute trailing mysql statement: %w", err)
		}
	}

	return nil
}

func (m *MySQLEngine) GetTableList(ctx context.Context) ([]string, error) {
	if err := m.connect(); err != nil {
		return nil, err
	}

	query := `
	SELECT table_name
	FROM information_schema.tables
	WHERE table_schema = ? AND table_type = 'BASE TABLE'
	ORDER BY table_name;
	`
	rows, err := m.db.QueryContext(ctx, query, m.connDetails.DatabaseName)
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

func (m *MySQLEngine) GetTableRowCount(ctx context.Context, tableName string) (int64, error) {
	if err := m.connect(); err != nil {
		return 0, err
	}

	clean := strings.ReplaceAll(tableName, "`", "``")
	query := fmt.Sprintf("SELECT COUNT(*) FROM `%s`;", clean)

	var count int64
	err := m.db.QueryRowContext(ctx, query).Scan(&count)
	return count, err
}

func (m *MySQLEngine) CheckConstraints(ctx context.Context) ([]models.ConstraintViolation, error) {
	// In MySQL, foreign key checks can be queried from INFORMATION_SCHEMA.INNODB_FOREIGN
	return []models.ConstraintViolation{}, nil
}

func (m *MySQLEngine) CheckIntegrity(ctx context.Context) ([]string, error) {
	if err := m.connect(); err != nil {
		return nil, err
	}

	tables, err := m.GetTableList(ctx)
	if err != nil {
		return nil, err
	}

	var errors []string
	for _, tbl := range tables {
		clean := strings.ReplaceAll(tbl, "`", "``")
		query := fmt.Sprintf("CHECK TABLE `%s` EXTENDED;", clean)
		rows, err := m.db.QueryContext(ctx, query)
		if err != nil {
			errors = append(errors, fmt.Sprintf("check table %s: %v", tbl, err))
			continue
		}

		for rows.Next() {
			var t, op, status, msg string
			_ = rows.Scan(&t, &op, &status, &msg)
			if status != "status" && msg != "OK" {
				errors = append(errors, fmt.Sprintf("table %s check: %s (%s)", tbl, msg, status))
			}
		}
		rows.Close()
	}
	return errors, nil
}

func (m *MySQLEngine) Close() error {
	if m.db != nil {
		return m.db.Close()
	}
	return nil
}
