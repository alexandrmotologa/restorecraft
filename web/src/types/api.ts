export type HealthGrade = 'A+' | 'A' | 'B' | 'C' | 'F';

export type DrillStatus = 'pending' | 'running' | 'completed' | 'failed';

export interface PhaseTimings {
  fetch_ms: number;
  decompress_ms: number;
  sandbox_ms: number;
  restore_ms: number;
  audit_ms: number;
  parity_ms: number;
  teardown_ms: number;
  total_rto_ms: number;
}

export interface ConstraintViolation {
  table: string;
  constraint: string;
  details: string;
  orphan_count?: number;
}

export interface TableParityRecord {
  table: string;
  expected_rows: number;
  actual_rows: number;
  diff: number;
  matched: boolean;
  missing_table: boolean;
  schema_matched: boolean;
}

export interface ParityReport {
  verified_tables: number;
  matched_tables: number;
  total_discrepant: number;
  records: TableParityRecord[];
}

export interface AuditReport {
  integrity_ok: boolean;
  structural_errors: string[];
  constraint_violations: ConstraintViolation[];
  parity: ParityReport;
  warnings: string[];
}

export interface DisasterRecoveryCertificate {
  certificate_id: string;
  drill_id: string;
  drill_name: string;
  database_engine: string;
  snapshot_path: string;
  snapshot_hash: string;
  snapshot_bytes: number;
  executed_at: string;
  health_score: number;
  health_grade: HealthGrade;
  rto_ms: number;
  rto_sla_ms: number;
  sla_met: boolean;
  integrity_ok: boolean;
  audit_summary: string;
  signer_identity: string;
  signature: string;
}

export interface DrillRun {
  id: string;
  config_id: string;
  name: string;
  engine: string;
  status: DrillStatus;
  snapshot_source: string;
  snapshot_path: string;
  snapshot_hash: string;
  snapshot_bytes: number;
  started_at: string;
  completed_at?: string;
  timings: PhaseTimings;
  health_score: number;
  health_grade: HealthGrade;
  rto_sla_seconds: number;
  sla_met: boolean;
  error_message?: string;
  audit: AuditReport;
  certificate?: DisasterRecoveryCertificate;
}

export interface LogEntry {
  drill_id: string;
  timestamp: string;
  level: string;
  phase: string;
  message: string;
}

export interface SystemStatus {
  total_drills: number;
  successful_drills: number;
  failed_drills: number;
  average_rto_ms: number;
  latest_grade: HealthGrade;
  latest_score: number;
  active_drill_id?: string;
}
