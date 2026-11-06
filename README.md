<p align="center">
  <img src="docs/images/logo.png?raw=true" alt="RestoreCraft Logo" width="130" style="border-radius: 24px;" />
</p>

<h1 align="center">restorecraft</h1>

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.23+-00ADD8?style=flat-square&logo=go&logoColor=white" alt="Go 1.23+" />
  <img src="https://img.shields.io/badge/React-19-61DAFB?style=flat-square&logo=react&logoColor=black" alt="React 19" />
  <img src="https://img.shields.io/badge/Tailwind-CSS-38B2AC?style=flat-square&logo=tailwind-css&logoColor=white" alt="Tailwind CSS" />
  <img src="https://img.shields.io/badge/Compliance-SOC%202%20%7C%20ISO%2027001%20%7C%20HIPAA-blue?style=flat-square" alt="Compliance Standards" />
  <img src="https://img.shields.io/badge/Prometheus-%2Fmetrics-E6522C?style=flat-square&logo=prometheus&logoColor=white" alt="Prometheus Metrics" />
  <a href="LICENSE"><img src="https://img.shields.io/badge/License-MIT-green?style=flat-square" alt="License: MIT" /></a>
</p>

<p align="center">
  Automated database disaster recovery drill, backup verification, and parity engine for PostgreSQL, MySQL, and SQLite.
</p>

<p align="center">
  <img src="docs/images/restorecraft_demo.gif?raw=true" alt="RestoreCraft Recovery Cockpit Demo" width="100%" style="border-radius: 12px; border: 1px solid #1e293b;" />
</p>

---

## The Core Problem

An untested backup is not a verified backup. Organizations frequently discover that backup snapshots are damaged only during a production outage. Database dumps fail due to storage bit-rot, truncated network transfers, unvalidated foreign key violations, or missing extensions.

`restorecraft` runs automated disaster recovery drills against real backup snapshots, verifies relational constraints, checks schema and row count parity, benchmarks your Recovery Time Objective (RTO), detects unmasked sensitive data, and issues cryptographically signed recovery certificates.

## Key Capabilities

- **Multi-Engine Restoration**: Tests backups for PostgreSQL (12-17), MySQL (8.0+), and SQLite 3.
- **Ephemeral Sandbox Isolation**: Provisions isolated ephemeral Docker containers or runs pure-Go local runners, automatically cleaning up temporary files and containers upon completion.
- **Deep Integrity Auditing**: Scans for foreign key constraint violations, broken referential links, index corruption, and page-level damage.
- **Schema & Drift Auditing**: Compares restored schemas, table structures, and columns against production baseline manifests.
- **Privacy & PII Leak Detection**: Audits restored tables for unmasked sensitive records (raw email addresses, payment cards, social security numbers) to verify data sanitization compliance before staging deployment.
- **Time-Travel Snapshot Bisect**: Runs chronological bisect searches across historical backup archives to pinpoint the exact date and snapshot where data corruption or structural regression was introduced.
- **Interactive Sandbox SQL Console**: Allows holding restored sandboxes open for live, read-only SQL inspection without touching production workloads.
- **Background Cron Scheduler**: Executes unattended recurring disaster recovery drills on custom cron schedules (`@daily`, `@hourly`, or standard 5-part cron syntax).
- **Prometheus Telemetry**: Exposes recovery metrics, health scores, and RTO gauges in Prometheus exposition format on `/metrics`.
- **Compliance Evidence Pack**: Exports printable, audit-grade Disaster Recovery Compliance Reports with HMAC-SHA256 digital seals for SOC 2 Type II, ISO/IEC 27001, and HIPAA compliance reviews.
- **Multi-Channel Alerting**: Dispatches instant notifications to Telegram, Slack, and Discord when drills fail or exceed SLA thresholds.
- **Embedded React 19 Cockpit**: Single-binary deployment with an embedded dark-mode web dashboard displaying live logs, health scores, and RTO trendlines.

## Visual Evidence & Interface

### Live Disaster Recovery Cockpit
Full overview showing health grades, mathematical RTO trendlines, live Server-Sent Events telemetry, and complete drill execution history.

<p align="center">
  <img src="docs/images/restorecraft_overview.png?raw=true" alt="RestoreCraft Desktop Overview" width="100%" style="border-radius: 10px; border: 1px solid #1e293b;" />
</p>

### Detailed Drill Inspection & Phase Stopwatch
Millisecond-resolution breakdown across snapshot fetch, decompression, sandbox boot, restore injection, and integrity auditing.

<p align="center">
  <img src="docs/images/restorecraft_drill_detail.png?raw=true" alt="RestoreCraft Drill Detail Modal" width="100%" style="border-radius: 10px; border: 1px solid #1e293b;" />
</p>

### Audit-Grade Compliance Evidence Pack
Printable report designed for SOC 2 Type II (CC9.1/CC9.2), ISO/IEC 27001:2022 (A.12.3.1), and HIPAA verification, complete with HMAC-SHA256 digital signatures and executive sign-off blocks.

<p align="center">
  <img src="docs/images/restorecraft_compliance.png?raw=true" alt="RestoreCraft Compliance Pack" width="100%" style="border-radius: 10px; border: 1px solid #1e293b;" />
</p>

### Interactive Ephemeral Sandbox SQL Console
Query restored databases directly inside ephemeral sandboxes using read-only SQL queries to inspect table data before teardown.

<p align="center">
  <img src="docs/images/restorecraft_sql_console.png?raw=true" alt="RestoreCraft Interactive SQL Console" width="100%" style="border-radius: 10px; border: 1px solid #1e293b;" />
</p>

### Time-Travel Snapshot Bisect Engine
Chronological evaluation across backup directories to locate the exact snapshot where data corruption was introduced.

<p align="center">
  <img src="docs/images/restorecraft_bisect.png?raw=true" alt="RestoreCraft Time-Travel Bisect" width="100%" style="border-radius: 10px; border: 1px solid #1e293b;" />
</p>

### Multi-Device Responsiveness (Mobile View)
Full responsive design adapted for smartphones and tablets with touch-friendly quick action toolbars.

<p align="center">
  <img src="docs/images/restorecraft_mobile.png?raw=true" alt="RestoreCraft Mobile View" width="360" style="border-radius: 16px; border: 1px solid #1e293b;" />
</p>

## Quickstart

### Installation

Clone the repository and build the binary:

```bash
git clone https://github.com/alexandrmotologa/restorecraft.git
cd restorecraft
go build -ldflags="-s -w -X 'main.Version=1.0.0'" -o restorecraft ./cmd/restorecraft
```

### Running the Web Cockpit & Daemon

Start the background daemon, cron scheduler, and web interface on port 8080:

```bash
./restorecraft serve --port 8080
```

Open `http://localhost:8080` in your browser to view the Recovery Cockpit.

### Running a Standalone Drill via CLI

Run an on-demand drill against an SQLite backup file:

```bash
./restorecraft drill --engine sqlite --snapshot ./backups/app.sqlite.gz --sla 60
```

Run a drill against a PostgreSQL dump using ephemeral Docker isolation:

```bash
./restorecraft drill --engine postgres --snapshot ./backups/prod.dump --docker
```

### Time-Travel Snapshot Bisect

Search through a folder of chronological backups to find the first broken snapshot:

```bash
./restorecraft drill-bisect --dir ./test-backups --engine sqlite
```

### Generating Test Fixtures

`restorecraft` includes a built-in generator to produce realistic valid and deliberately corrupted backups for chaos testing:

```bash
./restorecraft fixtures generate --dir ./test-backups
```

This generates:
1. `valid_ecommerce.sqlite`: Clean database with users, products, orders, and foreign keys.
2. `valid_ecommerce.sql.gz`: Gzipped streaming SQL dump verifying decompression pipelines.
3. `corrupt_header.sqlite`: SQLite file with damaged header bytes.
4. `fk_violation.sqlite`: Database with orphan rows violating foreign key constraints.
5. `truncated.sql.gz`: Interrupted gzip stream simulating a broken network transfer.

### Verifying a Recovery Certificate

Verify the authenticity of a signed Disaster Recovery Certificate using either the JSON file or its database ID:

```bash
# Verify via exported certificate JSON file
./restorecraft verify-cert --cert ./data/certificates/cert-drill-001.json

# Or verify directly by database certificate ID
./restorecraft verify-cert --id cert-drill-cli-drill-1793999059-1793999059
```

## Architecture

```
[ Storage: S3 / R2 / Local ]
             │
             ▼
[ Stream Decompressor (Gzip / Zstd / SHA-256) ]
             │
             ▼
[ Ephemeral Sandbox (Docker / Native) ] ◄───► [ Interactive SQL Console ]
             │
             ▼
[ Restoration Engine ] ──► [ High-Precision RTO Stopwatch ]
             │
             ▼
[ Integrity, Parity & PII Leak Auditor ]
             │
      ┌──────┴──────────────────────────┐
      ▼                                 ▼
[ DR Certificate (HMAC) ]      [ Alerts (Telegram/Slack/Discord) ]
      │                                 │
      ▼                                 ▼
[ Compliance Evidence Pack ]     [ Prometheus Metrics (/metrics) ]
      │
      ▼
[ React 19 Cockpit ]
```

## Configuration

`restorecraft` reads `restorecraft.yaml` by default. See `docs/configuration.md` for full schema details.

```yaml
version: "1"

server:
  host: "0.0.0.0"
  port: 8080
  data_dir: "./data"

security:
  signing_secret: "secret-key-for-hmac-signatures"
  signer_identity: "restorecraft-orchestrator-01"

defaults:
  rto_sla_seconds: 60
  auto_prune_sandboxes: true

drills:
  - id: "production-sqlite-drill"
    name: "Production SQLite Disaster Recovery Drill"
    engine: "sqlite"
    schedule: "0 2 * * *"
    storage_type: "local"
    snapshot_path: "./test-backups/valid_ecommerce.sqlite"
    rto_sla_seconds: 45
    audit_pii: true
    parity:
      expected_tables:
        - "users"
        - "orders"
        - "order_items"
      min_row_counts:
        users: 5
        orders: 5

notifications:
  telegram:
    enabled: false
    bot_token: "${TELEGRAM_BOT_TOKEN}"
    chat_id: "${TELEGRAM_CHAT_ID}"
    notify_on: "failures_only"
  slack:
    enabled: false
    webhook_url: "${SLACK_WEBHOOK_URL}"
    notify_on: "all"
```

## Health Grade Formula

Drills receive an objective grade based on data safety and SLA compliance:

$$\text{Health Score} = 100 - (\text{DataLossPenalty} \times 55) - (\text{ConstraintViolations} \times 20) - (\text{RTO Overrun Penalty})$$

- **Grade A+ (95-100%)**: Zero errors, intact relational constraints, RTO under SLA.
- **Grade A (85-94%)**: Restores cleanly, all constraints valid, RTO within SLA.
- **Grade B (70-84%)**: Non-critical warnings or RTO approaching threshold.
- **Grade C (50-69%)**: Restores with foreign key violations or schema mismatch.
- **Grade F (< 50%)**: Restoration failed, file corrupted, or operation timed out.

## License

MIT License. See [LICENSE](LICENSE) for details.
