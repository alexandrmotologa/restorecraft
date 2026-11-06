# `restorecraft` — Architecture & Engineering Master Plan

> **Automated Database Disaster Recovery Drill, Backup Verification & Parity Engine (PostgreSQL, MySQL, SQLite)**  
> *Solving "An Untested Backup is Not a Backup" in a Single Binary (Go + React 19)*

---

## 1. Executive Summary & Problem Space

### The Core Problem
The most dangerous misconception in IT operations is assuming that creating backup files (`pg_dump`, `mysqldump`, `litestream`, `xtrabackup`) means you can restore them. In reality:
- Dump files frequently suffer from silent storage bit-rot, truncated transfers, or incompatible server major version changes.
- PostgreSQL database extensions (e.g., `pgvector`, `postgis`, `timescaledb`) might be missing on the target host.
- Corrupted foreign keys or uncommitted transactions can cause the restore process to abort halfway through.
- When an actual disaster strikes, teams face disastrous downtime because their Recovery Time Objective (RTO) was never benchmarked or tested under realistic conditions.

### The Solution: `restorecraft`
`restorecraft` is an automated, single-binary disaster recovery drill daemon. On a configured schedule (e.g., daily at 03:00 UTC), it pulls the latest database backup from cloud storage (AWS S3, Cloudflare R2, MinIO, or local disk), spins up an isolated ephemeral sandbox environment, executes a complete restore from scratch, runs exhaustive structural and logical integrity audits, verifies row-count parity against production metadata, benchmarks RTO, and issues a cryptographically signed **Disaster Recovery Health Certificate**.

If a backup fails to restore or exceeds your RTO threshold, `restorecraft` fires immediate alerts to Telegram, Slack, or PagerDuty.

---

## 2. System Architecture & Drill Lifecycle

```
[ Cloud Storage (S3 / R2 / GCS) ]
               │
               │  1. Scheduled Pull (Daily / Weekly)
               ▼
┌────────────────────────────────────────────────────────────────────────┐
│                        `restorecraft` Daemon                           │
│                                                                        │
│  ┌──────────────────────────────────────────────────────────────────┐  │
│  │ Snapshot Fetcher & Checksum Verifier (SHA-256 / Gzip / Zstd)     │  │
│  └──────────────────────────────────┬───────────────────────────────┘  │
│                                     │                                  │
│                                     ▼                                  │
│  ┌──────────────────────────────────────────────────────────────────┐  │
│  │ Ephemeral Sandbox Orchestrator (Docker Engine API / Subprocess)  │  │
│  │ - Boots isolated ephemeral Postgres / MySQL / SQLite instance    │  │
│  │ - Configures memory limits, isolated networks, temp volumes      │  │
│  └──────────────────────────────────┬───────────────────────────────┘  │
│                                     │                                  │
│                                     ▼                                  │
│  ┌──────────────────────────────────────────────────────────────────┐  │
│  │ Automated Restoration Runner & Timer                             │  │
│  │ - Streams dump into sandbox (`pg_restore`, `mysql`, `.restore`)  │  │
│  │ - High-resolution stopwatch records RTO (Recovery Time Objective)│  │
│  └──────────────────────────────────┬───────────────────────────────┘  │
│                                     │                                  │
│                                     ▼                                  │
│  ┌──────────────────────────────────────────────────────────────────┐  │
│  │ Deep Integrity & Parity Auditor                                  │  │
│  │ - Foreign key constraint validation across all relational tables │  │
│  │ - Index corruption scan & `ANALYZE` / `PRAGMA integrity_check`   │  │
│  │ - Row count and schema parity check against production metadata  │  │
│  └──────────────────────────────────┬───────────────────────────────┘  │
│                                     │                                  │
│         ┌───────────────────────────┴──────────────────────────┐       │
│         ▼                                                      ▼       │
│  ┌───────────────────────────────┐           ┌──────────────────────┐  │
│  │ Teardown & Ephemeral Cleanup  │           │ Issue DR Certificate │  │
│  │ (Zero persistent disk waste)  │           │ & Dispatch Alerts    │  │
│  └───────────────────────────────┘           └──────────────────────┘  │
│                                                                        │
│  ┌──────────────────────────────────────────────────────────────────┐  │
│  │ Embedded React 19 Recovery Cockpit                               │  │
│  │ - Backup Health Score (A+ to F), RTO trend charts, drill logs    │  │
│  │ - 1-Click manual on-demand Disaster Recovery Drill trigger       │  │
│  └──────────────────────────────────────────────────────────────────┘  │
└────────────────────────────────────────────────────────────────────────┘
```

---

## 3. Technology Stack

| Layer | Component | Specification |
| :--- | :--- | :--- |
| **Language & Runtime** | Go 1.23+ | Native cross-platform daemon, concurrent sandbox orchestration, minimal RAM footprint |
| **Container & Sandbox Driver** | Docker API SDK + Subprocess | Manages ephemeral containers with isolated virtual networks and auto-pruning |
| **Database Engines Supported** | PostgreSQL (12–17), MySQL (8.0+), MariaDB, SQLite 3 | Native dump format parsers (`.sql`, `.dump`, `.tar`, `.gz`, `.zst`) |
| **Storage Connectors** | S3 API, Cloudflare R2, MinIO, Local Disk | Multi-part streaming downloads with resume and SHA-256 integrity validation |
| **State Persistence** | Embedded SQLite 3.45+ | Stores drill historical metrics, RTO tracking, and health certificates |
| **Notification Engine** | Multi-channel Webhooks | Telegram Bot, Slack incoming webhooks, Discord, and PagerDuty |
| **Embedded Web UI** | React 19 + TypeScript + Vite | Built as a single bundle and served via Go `//go:embed` |
| **UI Styling** | Tailwind CSS + Lucide Icons | High-contrast dark-mode reliability engineering cockpit |

---

## 4. Repository & Directory Structure

```
B:\workgit\restorecraft\
├── cmd\
│   └── restorecraft\
│       └── main.go                 # Daemon entrypoint, CLI commands (`drill`, `serve`, `verify`)
├── internal\
│   ├── config\
│   │   └── config.go               # YAML configuration (S3 credentials, databases, drill crons)
│   ├── orchestrator\
│   │   ├── runner.go               # Disaster recovery drill pipeline orchestrator
│   │   ├── sandbox.go              # Ephemeral Docker / local process sandbox manager
│   │   └── timer.go                # High-precision RTO timer and phase profiler
│   ├── database\
│   │   ├── postgres.go             # PostgreSQL restore runner and integrity auditor
│   │   ├── mysql.go                # MySQL / MariaDB restore runner and validation queries
│   │   └── sqlite.go               # SQLite memory/file restore and `PRAGMA` auditor
│   ├── auditor\
│   │   ├── constraints.go          # Foreign key and orphan row detector
│   │   ├── indexes.go              # Index corruption and accessibility scanner
│   │   └── parity.go               # Row count and schema diff comparison engine
│   ├── storage\
│   │   ├── s3.go                   # S3 / R2 / MinIO snapshot downloader
│   │   └── local.go                # Local disk snapshot manager
│   ├── certificate\
│   │   ├── generator.go            # Cryptographically signed Disaster Recovery Certificate
│   │   └── schema.go               # Certificate JSON data schema
│   ├── notifier\
│   │   ├── telegram.go             # Telegram Bot notification dispatcher
│   │   └── slack.go                # Slack webhook alert dispatcher
│   └── store\
│       ├── sqlite.go               # Historical drill database connector
│       └── models.go               # Drill run records, logs, and RTO data structures
├── web\
│   ├── src\
│   │   ├── components\
│   │   │   ├── HealthScoreBadge.tsx# Overall backup grade (A+, A, B, C, F)
│   │   │   ├── RtoChart.tsx        # Recovery Time Objective history and trendline
│   │   │   ├── DrillHistory.tsx    # Chronological table of drill runs with deep log modal
│   │   │   └── TriggerDrillBtn.tsx # 1-Click manual drill launcher with live log stream
│   │   ├── App.tsx                 # Root dashboard cockpit
│   │   ├── main.tsx                # React 19 entrypoint
│   │   └── index.css               # Tailwind CSS declarations
│   ├── package.json
│   ├── tsconfig.json
│   └── vite.config.ts
├── Makefile                        # Build, test, and container packaging
├── Dockerfile                      # Self-hosted single container image with Docker socket support
├── README.md                       # Documentation, architecture, and configuration examples
└── MASTER_PLAN.md                  # This master specification
```

---

## 5. Audit Matrix & Health Scoring Formula

### Deep Integrity Audit Checklists

| Engine | Phase 1: Structural Integrity | Phase 2: Relational Integrity | Phase 3: Parity Checks |
| :--- | :--- | :--- | :--- |
| **PostgreSQL** | `pg_restore --clean --if-exists`, check exit code, `pg_checksums` | `SELECT conrelid::regclass, confrelid::regclass FROM pg_constraint WHERE NOT convalidated;` | Validate table counts against production snapshot metadata |
| **MySQL** | `mysql < dump.sql`, verify warnings & errors | `CHECK TABLE tablename EXTENDED;`, foreign key check mode enabled | Row count validation per table, schema hash comparison |
| **SQLite** | Restore via online backup API | `PRAGMA integrity_check;`, `PRAGMA foreign_key_check;` | Compare `dbstat` page counts and total row sums |

### Backup Health Grade Formula
$$\text{Health Score} = 100 - (\text{DataLossPenalty} \times 50) - (\text{ConstraintViolations} \times 20) - (\text{RTO Overrun Penalty})$$
- **Grade A+ (95–100%)**: Zero errors, all constraints intact, RTO under SLA threshold.
- **Grade B (75–94%)**: Restores cleanly, but minor non-critical warnings or RTO approaching threshold.
- **Grade F (< 50%)**: Restoration failed, truncated file, broken foreign keys, or unrecoverable error.

---

## 6. Implementation Roadmap

### Phase 1: Sandbox Orchestration & Engine Drivers
- [ ] Initialize Go module `github.com/alexandrmotologa/restorecraft`.
- [ ] Implement Docker Engine API client to launch ephemeral PostgreSQL and MySQL containers.
- [ ] Implement SQLite sandbox runner without Docker requirements.
- [ ] Build auto-pruning logic ensuring zero orphaned containers or volumes remain on disk.

### Phase 2: Storage Fetcher & Database Restoration Runners
- [ ] Implement S3/R2 snapshot puller with streaming un-gzip and un-zstd decompression.
- [ ] Implement PostgreSQL restore driver (`pg_restore` and `psql` pipelines).
- [ ] Implement MySQL restore driver (`mysql` streaming pipeline).
- [ ] Capture millisecond-resolution timestamps across each drill phase (fetch, unpack, restore, audit).

### Phase 3: Integrity Auditor & Parity Verifier
- [ ] Implement relational foreign key constraint validation queries.
- [ ] Implement table row-count and schema hash parity validator.
- [ ] Implement JSON Disaster Recovery Certificate generator with HMAC-SHA256 digital signature.
- [ ] Implement Telegram and Slack alert dispatchers for drill outcomes.

### Phase 4: Embedded React 19 Recovery Cockpit
- [ ] Scaffold Vite + React 19 frontend in `web/`.
- [ ] Build interactive drill history with real-time log streaming.
- [ ] Build RTO historical performance line charts.
- [ ] Build 1-click on-demand drill execution button.
- [ ] Embed static assets into the Go binary using `//go:embed`.

### Phase 5: Production Hardening, Testing & Documentation
- [ ] Build end-to-end integration tests using deliberately corrupted dump files to verify failure detection.
- [ ] Write comprehensive documentation on configuring S3 lifecycles and cron schedules.
- [ ] Release single-binary binaries for Linux, macOS, and Windows.
