# restorecraft: Architecture and Engineering Master Plan

> Automated Database Disaster Recovery Drill, Backup Verification and Parity Engine (PostgreSQL, MySQL, SQLite)
> Single binary daemon and CLI built in Go with an embedded React 19 Recovery Cockpit.

---

## 1. Executive Summary and Problem Space

### The Problem
Organizations often assume that scheduled database dumps (`pg_dump`, `mysqldump`, `litestream`, or cloud snapshots) guarantee disaster recovery. In practice:
- Dump files silently suffer from storage bit-rot, truncated uploads, or incompatible database major version upgrades.
- Missing database extensions or changed collation configurations cause restores to fail unexpectedly.
- Corrupted foreign key references or uncommitted transaction logs abort restoration halfway.
- Teams have no empirical measurement of their Recovery Time Objective (RTO) under realistic restore conditions until an actual outage occurs.

### The Solution: restorecraft
`restorecraft` is an automated disaster recovery drill daemon and CLI tool. On a scheduled cron or on-demand trigger:
1. Pulls database backup snapshots from local storage, AWS S3, Cloudflare R2, or MinIO.
2. Boots an isolated ephemeral sandbox environment (Docker container or native local process runner).
3. Executes complete database restoration from the snapshot while benchmarking RTO at millisecond resolution.
4. Performs deep structural integrity checks, foreign key constraint validation, and row-count parity verification.
5. Emits a cryptographically signed Disaster Recovery Certificate (HMAC-SHA256).
6. Alerts incident teams via Telegram, Slack, Discord, or generic webhooks when health scores drop or RTO thresholds are exceeded.
7. Serves a real-time web recovery cockpit built in React 19 embedded directly in the Go binary.

---

## 2. System Architecture

```
[ Cloud Storage: S3 / R2 / MinIO / Local Disk ]
                     │
                     │  1. Scheduled or On-Demand Pull
                     ▼
┌────────────────────────────────────────────────────────────────────────┐
│                          restorecraft Core                             │
│                                                                        │
│  ┌──────────────────────────────────────────────────────────────────┐  │
│  │ Snapshot Fetcher & Stream Decompressor (Gzip / Zstd / Raw)       │  │
│  │ - Computes streaming SHA-256 checksum against metadata           │  │
│  └──────────────────────────────────┬───────────────────────────────┘  │
│                                     │                                  │
│                                     ▼                                  │
│  ┌──────────────────────────────────────────────────────────────────┐  │
│  │ Ephemeral Sandbox Orchestrator                                   │  │
│  │ - Docker Engine API driver (isolated containers, tmpfs, bridge)  │  │
│  │ - Native process driver (pure-Go SQLite / local process testbed) │  │
│  │ - Guaranteed auto-cleanup on drill completion or cancellation    │  │
│  └──────────────────────────────────┬───────────────────────────────┘  │
│                                     │                                  │
│                                     ▼                                  │
│  ┌──────────────────────────────────────────────────────────────────┐  │
│  │ Restoration Engine & Phase Stopwatch                             │  │
│  │ - Streams dump into sandbox (`pg_restore`, `mysql`, SQLite API)  │  │
│  │ - Measures individual phase durations and total RTO              │  │
│  └──────────────────────────────────┬───────────────────────────────┘  │
│                                     │                                  │
│                                     ▼                                  │
│  ┌──────────────────────────────────────────────────────────────────┐  │
│  │ Integrity & Parity Auditor                                       │  │
│  │ - Relational foreign key and orphan row scans                    │  │
│  │ - Page checksum and index validity checks                        │  │
│  │ - Row count and schema parity comparisons                        │  │
│  └──────────────────────────────────┬───────────────────────────────┘  │
│                                     │                                  │
│         ┌───────────────────────────┴──────────────────────────┐       │
│         ▼                                                      ▼       │
│  ┌───────────────────────────────┐           ┌──────────────────────┐  │
│  │ Ephemeral Sandbox Teardown    │           │ Issue DR Certificate │  │
│  │ - Clean container/temp prune  │           │ & Dispatch Alerts    │  │
│  └───────────────────────────────┘           └──────────────────────┘  │
│                                                                        │
│  ┌──────────────────────────────────────────────────────────────────┐  │
│  │ Embedded React 19 Recovery Cockpit                               │  │
│  │ - Health score grade (A+ through F) with deduction breakdown     │  │
│  │ - RTO trendline charts and live SSE log stream for active drills │  │
│  │ - Certificate verification inspector and manual drill launcher   │  │
│  └──────────────────────────────────────────────────────────────────┘  │
└────────────────────────────────────────────────────────────────────────┘
```

---

## 3. Technology Stack

| Layer | Component | Specification |
| :--- | :--- | :--- |
| Core Language | Go 1.23+ | Clean concurrent daemon, minimal memory footprint, cross-platform compilation |
| Sandbox Drivers | Docker Engine API SDK + Native Process | Ephemeral containers for Postgres/MySQL; pure-Go runner for SQLite |
| Database Engines | PostgreSQL 12-17, MySQL 8.0+, SQLite 3 | Native dump streaming (`.sql`, `.dump`, `.gz`, `.zst`, `.sqlite`) |
| Storage Sources | AWS S3, Cloudflare R2, MinIO, Local Disk | Streaming pull with on-the-fly checksum validation |
| Persistence Store | Embedded SQLite | Stores drill run history, phase metrics, and recovery certificates |
| Cryptography | HMAC-SHA256 | Cryptographically signed disaster recovery certificates |
| Web Cockpit | React 19, TypeScript, Vite, Tailwind CSS | High-contrast reliability engineering UI embedded via `embed.FS` |
| Alerting | Webhooks | Telegram Bot API, Slack Webhooks, Discord Webhooks, Generic Webhooks |

---

## 4. Repository and Directory Structure

```
B:\workgit\restorecraft\
├── cmd\
│   └── restorecraft\
│       └── main.go                     # CLI entrypoint (serve, drill, verify-cert, fixtures)
├── internal\
│   ├── config\
│   │   ├── config.go                   # YAML configuration loader and defaults
│   │   └── config_test.go              # Configuration unit tests
│   ├── models\
│   │   └── models.go                   # Drill run, phase timings, health grade, and certificate types
│   ├── storage\
│   │   ├── storage.go                  # Storage provider interface
│   │   ├── local.go                    # Local directory snapshot manager
│   │   ├── s3.go                       # S3, R2, and MinIO streaming reader
│   │   └── decompress.go               # Gzip, Zstd, and raw stream filters
│   ├── sandbox\
│   │   ├── sandbox.go                  # Sandbox interface and lifecycle management
│   │   ├── docker.go                   # Ephemeral Docker container orchestrator
│   │   └── native.go                   # Native process and SQLite in-memory sandbox
│   ├── engine\
│   │   ├── engine.go                   # Database engine restore and audit interface
│   │   ├── sqlite.go                   # SQLite restore, PRAGMA checks, and parity audits
│   │   ├── postgres.go                 # PostgreSQL pg_restore/psql runner and constraint auditor
│   │   └── mysql.go                    # MySQL dump runner and table integrity auditor
│   ├── auditor\
│   │   ├── auditor.go                  # Audit runner orchestrator
│   │   ├── constraints.go              # Foreign key and referential constraint checks
│   │   ├── parity.go                   # Schema and table row count parity comparisons
│   │   └── health.go                   # Health score formula and letter grade calculator
│   ├── certificate\
│   │   ├── signer.go                   # HMAC-SHA256 certificate generation and signing
│   │   └── verifier.go                 # Standalone certificate signature validation
│   ├── notifier\
│   │   ├── notifier.go                 # Alert dispatcher interface
│   │   ├── telegram.go                 # Telegram bot notifier
│   │   ├── slack.go                    # Slack webhook notifier
│   │   └── discord.go                  # Discord webhook notifier
│   ├── store\
│   │   ├── store.go                    # Drill history SQLite database repository
│   │   └── store_test.go               # History persistence unit tests
│   ├── orchestrator\
│   │   ├── orchestrator.go             # Drill execution coordinator and log hub
│   │   └── timer.go                    # High-precision phase stopwatch
│   ├── fixtures\
│   │   └── fixtures.go                 # Synthetic test database and corrupted snapshot generator
│   └── server\
│       ├── server.go                   # HTTP REST server and SSE live log streamer
│       └── handlers.go                 # API endpoints for drills, history, certificates, and config
├── web\
│   ├── src\
│   │   ├── components\
│   │   │   ├── HealthScoreBadge.tsx    # Health score grade visual badge
│   │   │   ├── RtoChart.tsx            # Recovery Time Objective trendline chart
│   │   │   ├── DrillHistoryTable.tsx   # Chronological drill runs with status pills
│   │   │   ├── DrillDetailModal.tsx    # Detailed phase timings, constraints, and audit logs
│   │   │   ├── RunDrillDialog.tsx      # On-demand manual drill modal with engine picker
│   │   │   ├── LiveLogStream.tsx       # Real-time console log viewer with SSE connection
│   │   │   ├── CertificateViewer.tsx   # Cryptographic certificate modal with signature check
│   │   │   └── Navbar.tsx              # Cockpit navigation and system status indicator
│   │   ├── types\
│   │   │   └── api.ts                  # TypeScript definitions matching Go models
│   │   ├── App.tsx                     # Main cockpit view and state management
│   │   ├── main.tsx                    # React 19 root
│   │   └── index.css                   # Tailwind CSS styling and theme definitions
│   ├── package.json
│   ├── tsconfig.json
│   └── vite.config.ts
├── docs\
│   ├── architecture.md                 # Detailed architecture documentation
│   ├── configuration.md                # Configuration options and environment variables
│   └── disaster_recovery_playbook.md   # Production DR drill best practices
├── Makefile                            # Build, test, and release targets
├── Dockerfile                          # Production container image
├── README.md                           # Documentation, quickstart, and configuration guides
└── MASTER_PLAN.md                      # This specification
```

---

## 5. Audit Matrix and Health Scoring Formula

### Deep Integrity Audit Checklists

| Database Engine | Structural Integrity | Relational Integrity | Parity Checks |
| :--- | :--- | :--- | :--- |
| PostgreSQL | `pg_restore --clean --if-exists`, verify non-zero exit codes, detect parse warnings | Validate unvalidated constraints in `pg_constraint` where `NOT convalidated` | Compare table counts and schema column hashes against manifest |
| MySQL | `mysql < dump.sql`, verify warnings, SQL errors, and table locks | `CHECK TABLE tablename EXTENDED`, verify foreign key check state | Verify row count per table against production snapshot metadata |
| SQLite | Restore via online backup API or file copy | `PRAGMA integrity_check`, `PRAGMA foreign_key_check` | Compare page counts, schema table counts, and row counts |

### Health Grade Calculation

$$\text{Health Score} = 100 - (\text{DataLossPenalty} \times 50) - (\text{ConstraintViolations} \times 20) - (\text{RTO Overrun Penalty})$$

- **Grade A+ (95 to 100)**: Restoration succeeded cleanly, all foreign keys and constraints intact, RTO strictly below SLA threshold.
- **Grade A (85 to 94)**: Restoration succeeded cleanly, all constraints intact, RTO within SLA threshold.
- **Grade B (70 to 84)**: Restoration succeeded, non-critical warnings detected or RTO exceeded SLA threshold by less than 25%.
- **Grade C (50 to 69)**: Restoration succeeded but foreign key violations or schema discrepancies detected.
- **Grade F (0 to 49)**: Restoration failed, corrupted archive, broken file header, unrecoverable query error, or total timeout.

---

## 6. Implementation Roadmap

### Phase 1: Core Foundation and Models
- Initialize Go module `github.com/alexandrmotologa/restorecraft`.
- Define core data models for drill runs, phase timings, audit reports, and certificates.
- Implement configuration parser (`internal/config`) with YAML loading, environment overrides, and SLA thresholds.
- Implement embedded SQLite persistence store (`internal/store`) for recording drill runs and metrics.

### Phase 2: Storage Connectors and Decompression Engine
- Implement local file snapshot provider (`internal/storage/local.go`).
- Implement S3, Cloudflare R2, and MinIO snapshot provider (`internal/storage/s3.go`).
- Implement streaming decompressors for gzip, zstd, and raw SQL/tar streams (`internal/storage/decompress.go`).
- Implement on-the-fly SHA-256 streaming checksum calculator.

### Phase 3: Ephemeral Sandbox and Database Restore Runners
- Implement sandbox interface and lifecycle management (`internal/sandbox/sandbox.go`).
- Implement ephemeral Docker container orchestrator (`internal/sandbox/docker.go`) with tmpfs and isolated networks.
- Implement native isolated process and pure-Go SQLite sandbox runner (`internal/sandbox/native.go`).
- Implement database restoration engines for SQLite, PostgreSQL, and MySQL (`internal/engine/`).
- Implement high-resolution phase stopwatch for accurate RTO profiling (`internal/orchestrator/timer.go`).

### Phase 4: Integrity Auditor, Parity Engine and Certificate Signer
- Implement relational constraint auditor for SQLite, PostgreSQL, and MySQL (`internal/auditor/constraints.go`).
- Implement table row count and schema parity comparator (`internal/auditor/parity.go`).
- Implement health score formula and grading calculator (`internal/auditor/health.go`).
- Implement HMAC-SHA256 Disaster Recovery Certificate generator and standalone verifier (`internal/certificate/`).
- Implement alert dispatchers for Telegram, Slack, and Discord (`internal/notifier/`).

### Phase 5: Drill Orchestrator, Live Log Hub and Synthetic Fixtures
- Implement central drill orchestrator coordinating fetch, sandbox boot, restore, audit, and teardown (`internal/orchestrator/orchestrator.go`).
- Implement real-time log event hub supporting subscription and SSE broadcasting.
- Implement synthetic testbed generator (`internal/fixtures/fixtures.go`) providing realistic clean backups as well as deliberately corrupted backups (bit-rot, missing foreign keys, truncated archives).

### Phase 6: Embedded React 19 Recovery Cockpit
- Scaffold React 19, TypeScript, and Vite application in `web/`.
- Configure Tailwind CSS with high-contrast reliability cockpit theme.
- Build Health Score summary cards with grade breakdown.
- Build interactive RTO history trendline chart.
- Build Drill History table with detailed modal viewer.
- Build Live Log Stream component with Server-Sent Events (SSE) integration.
- Build On-Demand Drill Trigger dialog with engine and snapshot selection.
- Build Disaster Recovery Certificate viewer with signature copy and verification badge.
- Build HTTP API server in Go (`internal/server/`) serving REST endpoints, SSE streams, and embedded React static assets via Go `embed.FS`.

### Phase 7: Verification, Testing and Production Packaging
- Write comprehensive unit tests for configuration, storage, sandboxes, engines, auditors, and certificates.
- Write end-to-end integration tests using synthetic valid and corrupt backups to verify failure grading.
- Compile and test standalone binary `restorecraft.exe` on Windows and verify clean startup.
- Verify web cockpit in browser.
- Write clear, factual documentation (`README.md`, `docs/architecture.md`, `docs/configuration.md`, `docs/disaster_recovery_playbook.md`).
