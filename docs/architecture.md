# Architecture Specification

## Overview

`restorecraft` is an automated database disaster recovery drill daemon and verification engine. Its primary role is to run automated, isolated drills against real database backups to verify that snapshots can be restored without errors, satisfy schema and relational constraints, and recover within your designated Recovery Time Objective (RTO).

## Drill Execution Lifecycle

Every recovery drill follows a strict six-phase sequence:

1. **Snapshot Acquisition and Checksum Verification**:
   The storage fetcher pulls the configured database backup file from local storage, AWS S3, Cloudflare R2, or MinIO. As data streams in, `restorecraft` computes a running SHA-256 digest to detect storage bit-rot or transfer truncation before any execution begins.

2. **Ephemeral Sandbox Provisioning**:
   The orchestrator boots an isolated sandbox environment. Depending on the configuration and database engine:
   - For PostgreSQL and MySQL: `restorecraft` can connect to the Docker Engine API to launch an isolated container with an internal bridge network, temporary `tmpfs` volume mounts for high throughput, and strict memory limits.
   - For SQLite or lightweight environments: `restorecraft` uses a pure-Go in-memory/tempfile runner or isolated subprocess that requires no container daemon.

3. **Restoration and RTO Stopwatch**:
   The engine decompresses the backup on the fly (supporting raw `.sql`, gzip `.gz`, or zstandard `.zst`) and streams the data into the target database. A nanosecond-resolution stopwatch records the exact duration of each step.

4. **Deep Integrity and Constraint Audit**:
   Once the database is populated, the auditor executes verification queries:
   - Evaluates foreign key integrity and reports any orphaned child rows.
   - Runs database structural integrity checks (`PRAGMA integrity_check` for SQLite, `CHECK TABLE ... EXTENDED` for MySQL, catalog checks for PostgreSQL).
   - Validates index availability and scan health.

5. **Parity Verification**:
   If a parity manifest or expected assertions are configured, `restorecraft` compares table counts, individual row tallies, and schema column structures against the expected production baseline. Discrepancies reduce the overall drill health score.

6. **Certificate Issuance, Alerting, and Teardown**:
   - `restorecraft` generates a structured Disaster Recovery Certificate signed with HMAC-SHA256.
   - If the drill fails or exceeds the RTO SLA, notifications are dispatched to Telegram, Slack, Discord, or generic webhook targets.
   - The sandbox is pruned immediately, deleting temporary containers and volumes to prevent residual disk usage.

## Component Interactions

```
CLI / Scheduled Cron
       │
       ▼
Orchestrator ─────────► Storage Fetcher (S3 / R2 / Local Disk)
       │                        │
       │ (Streams dump stream)  ▼
       ├──────────────► Stream Decompressor (Gzip / Zstd)
       │                        │
       ▼                        ▼
Sandbox Manager ──────► Ephemeral DB Engine (Postgres / MySQL / SQLite)
       │                        │
       ▼                        ▼
Integrity Auditor ────► Foreign Key, Index & Parity Verifier
       │                        │
       ▼                        ▼
Certificate Signer ───► Signed JSON DR Certificate (HMAC-SHA256)
       │
       ├──────────────► Webhook Dispatcher (Telegram / Slack / Discord)
       │
       └──────────────► SQLite Drill History Store & React 19 Cockpit
```

## Security and Isolation

- Ephemeral containers run without host port publication, binding only to isolated private bridge networks or Unix domain sockets.
- Database passwords and S3 secrets are read from configuration or environment variables and never logged or exposed in drill certificates.
- Disaster Recovery Certificates are signed with a secret key so compliance auditors can verify that the certificate was produced by the authorized `restorecraft` runner.
