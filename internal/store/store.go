package store

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/alexandrmotologa/restorecraft/internal/models"
	_ "modernc.org/sqlite"
)

// Store handles persistence for drill executions and recovery certificates.
type Store struct {
	db *sql.DB
	mu sync.RWMutex
}

// NewStore initializes an embedded SQLite database at the given directory path.
func NewStore(dataDir string) (*Store, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("create data directory: %w", err)
	}

	dbPath := filepath.Join(dataDir, "restorecraft.db")
	db, err := sql.Open("sqlite", dbPath+"?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)")
	if err != nil {
		return nil, fmt.Errorf("open sqlite database: %w", err)
	}

	// Single connection writer for embedded SQLite WAL safety
	db.SetMaxOpenConns(1)

	s := &Store{db: db}
	if err := s.initSchema(); err != nil {
		db.Close()
		return nil, fmt.Errorf("init schema: %w", err)
	}

	return s, nil
}

func (s *Store) initSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS drills (
		id TEXT PRIMARY KEY,
		config_id TEXT NOT NULL,
		name TEXT NOT NULL,
		engine TEXT NOT NULL,
		status TEXT NOT NULL,
		snapshot_source TEXT NOT NULL,
		snapshot_path TEXT NOT NULL,
		snapshot_hash TEXT NOT NULL,
		snapshot_bytes INTEGER NOT NULL,
		started_at DATETIME NOT NULL,
		completed_at DATETIME,
		health_score REAL NOT NULL,
		health_grade TEXT NOT NULL,
		rto_sla_seconds INTEGER NOT NULL,
		sla_met BOOLEAN NOT NULL,
		error_message TEXT,
		timings_json TEXT NOT NULL,
		audit_json TEXT NOT NULL,
		certificate_id TEXT
	);

	CREATE INDEX IF NOT EXISTS idx_drills_started_at ON drills(started_at DESC);

	CREATE TABLE IF NOT EXISTS certificates (
		certificate_id TEXT PRIMARY KEY,
		drill_id TEXT NOT NULL,
		certificate_json TEXT NOT NULL,
		issued_at DATETIME NOT NULL
	);
	`

	_, err := s.db.Exec(schema)
	return err
}

// SaveDrill inserts or updates a drill record.
func (s *Store) SaveDrill(drill *models.DrillRun) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	timingsBytes, err := json.Marshal(drill.Timings)
	if err != nil {
		return fmt.Errorf("marshal timings: %w", err)
	}

	auditBytes, err := json.Marshal(drill.Audit)
	if err != nil {
		return fmt.Errorf("marshal audit: %w", err)
	}

	var certID sql.NullString
	if drill.Certificate != nil {
		certID = sql.NullString{String: drill.Certificate.CertificateID, Valid: true}
		if err := s.saveCertificateLocked(drill.Certificate); err != nil {
			return err
		}
	}

	query := `
	INSERT INTO drills (
		id, config_id, name, engine, status, snapshot_source, snapshot_path,
		snapshot_hash, snapshot_bytes, started_at, completed_at, health_score,
		health_grade, rto_sla_seconds, sla_met, error_message, timings_json,
		audit_json, certificate_id
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	ON CONFLICT(id) DO UPDATE SET
		status = excluded.status,
		snapshot_hash = excluded.snapshot_hash,
		snapshot_bytes = excluded.snapshot_bytes,
		completed_at = excluded.completed_at,
		health_score = excluded.health_score,
		health_grade = excluded.health_grade,
		sla_met = excluded.sla_met,
		error_message = excluded.error_message,
		timings_json = excluded.timings_json,
		audit_json = excluded.audit_json,
		certificate_id = excluded.certificate_id;
	`

	var completedAt sql.NullTime
	if drill.CompletedAt != nil {
		completedAt = sql.NullTime{Time: *drill.CompletedAt, Valid: true}
	}

	_, err = s.db.Exec(query,
		drill.ID, drill.ConfigID, drill.Name, drill.Engine, string(drill.Status),
		drill.SnapshotSource, drill.SnapshotPath, drill.SnapshotHash, drill.SnapshotBytes,
		drill.StartedAt, completedAt, drill.HealthScore, string(drill.HealthGrade),
		drill.RtoSlaSeconds, drill.SlaMet, drill.ErrorMessage,
		string(timingsBytes), string(auditBytes), certID,
	)
	return err
}

func (s *Store) saveCertificateLocked(cert *models.DisasterRecoveryCertificate) error {
	certBytes, err := json.Marshal(cert)
	if err != nil {
		return fmt.Errorf("marshal certificate: %w", err)
	}

	query := `
	INSERT INTO certificates (certificate_id, drill_id, certificate_json, issued_at)
	VALUES (?, ?, ?, ?)
	ON CONFLICT(certificate_id) DO UPDATE SET certificate_json = excluded.certificate_json;
	`
	_, err = s.db.Exec(query, cert.CertificateID, cert.DrillID, string(certBytes), cert.ExecutedAt)
	return err
}

// GetDrill retrieves a single drill run by ID.
func (s *Store) GetDrill(id string) (*models.DrillRun, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	query := `
	SELECT id, config_id, name, engine, status, snapshot_source, snapshot_path,
	       snapshot_hash, snapshot_bytes, started_at, completed_at, health_score,
	       health_grade, rto_sla_seconds, sla_met, error_message, timings_json,
	       audit_json, certificate_id
	FROM drills WHERE id = ?;
	`

	row := s.db.QueryRow(query, id)
	return s.scanDrill(row)
}

// ListDrills returns the most recent drill runs.
func (s *Store) ListDrills(limit int) ([]models.DrillRun, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if limit <= 0 {
		limit = 50
	}

	query := `
	SELECT id, config_id, name, engine, status, snapshot_source, snapshot_path,
	       snapshot_hash, snapshot_bytes, started_at, completed_at, health_score,
	       health_grade, rto_sla_seconds, sla_met, error_message, timings_json,
	       audit_json, certificate_id
	FROM drills
	ORDER BY started_at DESC
	LIMIT ?;
	`

	rows, err := s.db.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []models.DrillRun
	for rows.Next() {
		drill, err := s.scanDrillRow(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, *drill)
	}

	return result, rows.Err()
}

// GetCertificate returns a stored certificate by its ID.
func (s *Store) GetCertificate(id string) (*models.DisasterRecoveryCertificate, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	query := `SELECT certificate_json FROM certificates WHERE certificate_id = ?;`
	row := s.db.QueryRow(query, id)

	var jsonStr string
	if err := row.Scan(&jsonStr); err != nil {
		return nil, err
	}

	var cert models.DisasterRecoveryCertificate
	if err := json.Unmarshal([]byte(jsonStr), &cert); err != nil {
		return nil, fmt.Errorf("unmarshal certificate: %w", err)
	}
	return &cert, nil
}

// GetSystemStatus computes summary statistics for the cockpit.
func (s *Store) GetSystemStatus() (*models.SystemStatus, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	status := &models.SystemStatus{
		LatestGrade: models.GradeA,
		LatestScore: 100.0,
	}

	query := `
	SELECT 
		COUNT(*),
		COALESCE(SUM(CASE WHEN status = 'completed' AND health_grade != 'F' THEN 1 ELSE 0 END), 0),
		COALESCE(SUM(CASE WHEN status = 'failed' OR health_grade = 'F' THEN 1 ELSE 0 END), 0)
	FROM drills;
	`
	err := s.db.QueryRow(query).Scan(&status.TotalDrills, &status.SuccessfulDrills, &status.FailedDrills)
	if err != nil {
		return nil, err
	}

	// Calculate average RTO of completed runs
	var avgRto sql.NullFloat64
	rtoQuery := `
	SELECT AVG(json_extract(timings_json, '$.total_rto_ms'))
	FROM drills
	WHERE status = 'completed';
	`
	_ = s.db.QueryRow(rtoQuery).Scan(&avgRto)
	if avgRto.Valid {
		status.AverageRtoMs = int64(avgRto.Float64)
	}

	// Retrieve latest drill grade
	var latestGrade sql.NullString
	var latestScore sql.NullFloat64
	latestQuery := `
	SELECT health_grade, health_score
	FROM drills
	WHERE status = 'completed'
	ORDER BY started_at DESC
	LIMIT 1;
	`
	_ = s.db.QueryRow(latestQuery).Scan(&latestGrade, &latestScore)
	if latestGrade.Valid {
		status.LatestGrade = models.HealthGrade(latestGrade.String)
		status.LatestScore = latestScore.Float64
	}

	return status, nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func (s *Store) scanDrill(scanner rowScanner) (*models.DrillRun, error) {
	var drill models.DrillRun
	var statusStr, gradeStr, timingsStr, auditStr string
	var completedAt sql.NullTime
	var errMsg, certID sql.NullString

	err := scanner.Scan(
		&drill.ID, &drill.ConfigID, &drill.Name, &drill.Engine, &statusStr,
		&drill.SnapshotSource, &drill.SnapshotPath, &drill.SnapshotHash, &drill.SnapshotBytes,
		&drill.StartedAt, &completedAt, &drill.HealthScore, &gradeStr,
		&drill.RtoSlaSeconds, &drill.SlaMet, &errMsg, &timingsStr,
		&auditStr, &certID,
	)
	if err != nil {
		return nil, err
	}

	drill.Status = models.DrillStatus(statusStr)
	drill.HealthGrade = models.HealthGrade(gradeStr)
	if completedAt.Valid {
		drill.CompletedAt = &completedAt.Time
	}
	if errMsg.Valid {
		drill.ErrorMessage = errMsg.String
	}

	if timingsStr != "" {
		_ = json.Unmarshal([]byte(timingsStr), &drill.Timings)
	}
	if auditStr != "" {
		_ = json.Unmarshal([]byte(auditStr), &drill.Audit)
	}

	if certID.Valid && certID.String != "" {
		cert, _ := s.GetCertificate(certID.String)
		drill.Certificate = cert
	}

	return &drill, nil
}

func (s *Store) scanDrillRow(rows *sql.Rows) (*models.DrillRun, error) {
	return s.scanDrill(rows)
}

// Close closes the underlying database handle.
func (s *Store) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}
