package models

import (
	"time"
)

// HealthGrade represents the safety and readiness rating.
type HealthGrade string

const (
	GradeAPlus HealthGrade = "A+"
	GradeA     HealthGrade = "A"
	GradeB     HealthGrade = "B"
	GradeC     HealthGrade = "C"
	GradeF     HealthGrade = "F"
)

// DrillStatus represents the execution state of a drill.
type DrillStatus string

const (
	StatusPending   DrillStatus = "pending"
	StatusRunning   DrillStatus = "running"
	StatusCompleted DrillStatus = "completed"
	StatusFailed    DrillStatus = "failed"
)

// PhaseTimings records nanosecond-resolution duration of each drill step.
type PhaseTimings struct {
	FetchMs      int64 `json:"fetch_ms"`
	DecompressMs int64 `json:"decompress_ms"`
	SandboxMs    int64 `json:"sandbox_ms"`
	RestoreMs    int64 `json:"restore_ms"`
	AuditMs      int64 `json:"audit_ms"`
	ParityMs     int64 `json:"parity_ms"`
	TeardownMs   int64 `json:"teardown_ms"`
	TotalRtoMs   int64 `json:"total_rto_ms"`
}

// ConstraintViolation represents a broken foreign key or relational constraint.
type ConstraintViolation struct {
	Table       string `json:"table"`
	Constraint  string `json:"constraint"`
	Details     string `json:"details"`
	OrphanCount int64  `json:"orphan_count,omitempty"`
}

// TableParityRecord holds parity comparison results for a specific table.
type TableParityRecord struct {
	Table         string `json:"table"`
	ExpectedRows  int64  `json:"expected_rows"`
	ActualRows    int64  `json:"actual_rows"`
	Diff          int64  `json:"diff"`
	Matched       bool   `json:"matched"`
	MissingTable  bool   `json:"missing_table"`
	SchemaMatched bool   `json:"schema_matched"`
}

// ParityReport contains the overall parity comparison details.
type ParityReport struct {
	VerifiedTables int                 `json:"verified_tables"`
	MatchedTables  int                 `json:"matched_tables"`
	TotalDiscrepant int                `json:"total_discrepant"`
	Records        []TableParityRecord `json:"records"`
}

// AuditReport aggregates all integrity and parity verification findings.
type AuditReport struct {
	IntegrityOk          bool                  `json:"integrity_ok"`
	StructuralErrors     []string              `json:"structural_errors"`
	ConstraintViolations []ConstraintViolation `json:"constraint_violations"`
	Parity               ParityReport          `json:"parity"`
	Warnings             []string              `json:"warnings"`
}

// DisasterRecoveryCertificate is the auditable cryptographic document issued after each drill.
type DisasterRecoveryCertificate struct {
	CertificateID string       `json:"certificate_id"`
	DrillID       string       `json:"drill_id"`
	DrillName     string       `json:"drill_name"`
	DatabaseEngine string      `json:"database_engine"`
	SnapshotPath  string       `json:"snapshot_path"`
	SnapshotHash  string       `json:"snapshot_hash"`
	SnapshotBytes int64        `json:"snapshot_bytes"`
	ExecutedAt    time.Time    `json:"executed_at"`
	HealthScore   float64      `json:"health_score"`
	HealthGrade   HealthGrade  `json:"health_grade"`
	RtoMs         int64        `json:"rto_ms"`
	RtoSlaMs      int64        `json:"rto_sla_ms"`
	SlaMet        bool         `json:"sla_met"`
	IntegrityOk   bool         `json:"integrity_ok"`
	AuditSummary  string       `json:"audit_summary"`
	SignerIdentity string      `json:"signer_identity"`
	Signature     string       `json:"signature"`
}

// DrillRun represents a single disaster recovery drill execution record.
type DrillRun struct {
	ID             string                       `json:"id"`
	ConfigID       string                       `json:"config_id"`
	Name           string                       `json:"name"`
	Engine         string                       `json:"engine"`
	Status         DrillStatus                  `json:"status"`
	SnapshotSource string                       `json:"snapshot_source"`
	SnapshotPath   string                       `json:"snapshot_path"`
	SnapshotHash   string                       `json:"snapshot_hash"`
	SnapshotBytes  int64                        `json:"snapshot_bytes"`
	StartedAt      time.Time                    `json:"started_at"`
	CompletedAt    *time.Time                   `json:"completed_at,omitempty"`
	Timings        PhaseTimings                 `json:"timings"`
	HealthScore    float64                      `json:"health_score"`
	HealthGrade    HealthGrade                  `json:"health_grade"`
	RtoSlaSeconds  int                          `json:"rto_sla_seconds"`
	SlaMet         bool                         `json:"sla_met"`
	ErrorMessage   string                       `json:"error_message,omitempty"`
	Audit          AuditReport                  `json:"audit"`
	Certificate    *DisasterRecoveryCertificate `json:"certificate,omitempty"`
}

// LogEntry is a real-time event log emitted during drill execution.
type LogEntry struct {
	DrillID   string    `json:"drill_id"`
	Timestamp time.Time `json:"timestamp"`
	Level     string    `json:"level"` // "INFO", "WARN", "ERROR", "STEP"
	Phase     string    `json:"phase"` // "fetch", "sandbox", "restore", "audit", "teardown"
	Message   string    `json:"message"`
}

// SystemStatus provides high-level metrics for the recovery cockpit.
type SystemStatus struct {
	TotalDrills     int         `json:"total_drills"`
	SuccessfulDrills int        `json:"successful_drills"`
	FailedDrills    int         `json:"failed_drills"`
	AverageRtoMs    int64       `json:"average_rto_ms"`
	LatestGrade     HealthGrade `json:"latest_grade"`
	LatestScore     float64     `json:"latest_score"`
	ActiveDrillID   string      `json:"active_drill_id,omitempty"`
}
