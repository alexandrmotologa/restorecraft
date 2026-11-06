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

export interface SchemaDiffReport {
  matched: boolean;
  missing_tables: string[];
  unexpected_tables: string[];
  missing_columns: Record<string, string[]>;
  altered_columns: Record<string, string[]>;
}

export interface PiiFinding {
  table: string;
  column: string;
  pii_type: string;
  sample_count: number;
  description: string;
}

export interface PiiAuditReport {
  sanitized_ok: boolean;
  checked: boolean;
  findings: PiiFinding[];
}

export interface AuditReport {
  integrity_ok: boolean;
  structural_errors: string[];
  constraint_violations: ConstraintViolation[];
  parity: ParityReport;
  schema_diff?: SchemaDiffReport;
  pii_audit?: PiiAuditReport;
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

export interface ActiveSandboxInfo {
  id: string;
  drill_id: string;
  engine: string;
  type: string;
  hold_until: string;
  file_path?: string;
}

export interface QueryResult {
  columns: string[];
  rows: string[][];
  row_count: number;
  duration_ms: number;
}

export interface ScheduledJob {
  drill_config: {
    id: string;
    name: string;
    engine: string;
    schedule: string;
  };
  schedule: string;
  last_run_at?: string;
  next_run_at: string;
  running: boolean;
}

export interface BisectPoint {
  snapshot_path: string;
  health_grade: HealthGrade;
  health_score: number;
  status: DrillStatus;
  rto_ms: number;
  integrity_ok: boolean;
  error_message?: string;
}

export interface BisectReport {
  total_snapshots: number;
  tested_snapshots: number;
  first_broken_snapshot?: string;
  first_failure_reason?: string;
  timeline: BisectPoint[];
  healthy_count: number;
  broken_count: number;
}
