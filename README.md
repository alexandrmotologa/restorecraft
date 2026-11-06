# restorecraft

Automated database disaster recovery drill, backup verification, and parity engine for PostgreSQL, MySQL, and SQLite.

An untested backup is not a verified backup. `restorecraft` runs automated disaster recovery drills against real backup snapshots, verifies relational constraints, checks schema and row count parity, benchmarks your Recovery Time Objective (RTO), and issues cryptographically signed recovery certificates.

## Key Capabilities

- **Multi-Engine Restoration**: Tests backups for PostgreSQL (12-17), MySQL (8.0+), and SQLite 3.
- **Ephemeral Sandbox Isolation**: Provisions isolated ephemeral Docker containers or runs pure-Go local runners, automatically cleaning up temporary files and containers upon completion.
- **Deep Integrity Auditing**: Scans for foreign key constraint violations, broken referential links, and index corruption.
- **Parity Checking**: Compares restored schemas, table structures, and row counts against production baseline manifests.
- **Precise RTO Benchmarking**: Records millisecond-level duration across download, decompression, restoration, and auditing phases.
- **Cryptographic Certificates**: Generates auditable JSON Disaster Recovery Certificates signed with HMAC-SHA256 for compliance audits (SOC 2, ISO 27001).
- **Multi-Channel Alerting**: Dispatches instant notifications to Telegram, Slack, and Discord when drills fail or exceed SLA thresholds.
- **Embedded React 19 Cockpit**: Single-binary deployment with an embedded dark-mode web dashboard displaying live logs, health scores, and RTO trendlines.

## Quickstart

### Installation

Clone the repository and build the binary:

```bash
git clone https://github.com/alexandrmotologa/restorecraft.git
cd restorecraft
go build -o restorecraft ./cmd/restorecraft
```

### Running the Web Cockpit

Start the daemon and web interface on port 8080:

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

### Generating Test Fixtures

`restorecraft` includes a built-in generator to produce realistic valid and deliberately corrupted backups for chaos testing:

```bash
./restorecraft fixtures generate --dir ./test-backups
```

This generates:
1. `valid_ecommerce.sqlite`: Clean database with users, products, orders, and foreign keys.
2. `corrupt_header.sqlite`: SQLite file with damaged header bytes.
3. `fk_violation.sqlite`: Database with orphan rows violating foreign key constraints.
4. `truncated.sql.gz`: Interrupted gzip stream simulating a broken network transfer.

### Verifying a Recovery Certificate

Verify the authenticity of a signed Disaster Recovery Certificate:

```bash
./restorecraft verify-cert --cert ./certificates/cert-drill-001.json --secret my-signing-key
```

## Architecture

```
[ Storage: S3 / R2 / Local ]
             │
             ▼
[ Stream Decompressor (Gzip / Zstd) ]
             │
             ▼
[ Ephemeral Sandbox (Docker / Native) ]
             │
             ▼
[ Restoration Engine ] ──► [ RTO Phase Stopwatch ]
             │
             ▼
[ Integrity & Parity Auditor ]
             │
     ┌───────┴───────┐
     ▼               ▼
[ DR Certificate ]  [ Alerts (Telegram/Slack) ]
     │
     ▼
[ React 19 Cockpit ]
```

## Configuration

`restorecraft` reads `restorecraft.yaml` by default. See `docs/configuration.md` for full schema details.

```yaml
version: "1"

server:
  port: 8080
  data_dir: "./data"

security:
  signing_secret: "secret-key-for-hmac-signatures"

drills:
  - id: "main-sqlite-drill"
    engine: "sqlite"
    schedule: "0 2 * * *"
    storage_type: "local"
    snapshot_path: "./test-backups/valid_ecommerce.sqlite"
    rto_sla_seconds: 45
    parity:
      expected_tables:
        - "users"
        - "orders"
        - "order_items"
      min_row_counts:
        users: 10
        orders: 20
```

## Health Grade Formula

Drills receive an objective grade based on data safety and SLA compliance:

$$\text{Health Score} = 100 - (\text{DataLossPenalty} \times 50) - (\text{ConstraintViolations} \times 20) - (\text{RTO Overrun Penalty})$$

- **Grade A+ (95-100%)**: Zero errors, intact relational constraints, RTO under SLA.
- **Grade A (85-94%)**: Restores cleanly, all constraints valid, RTO within SLA.
- **Grade B (70-84%)**: Non-critical warnings or RTO approaching threshold.
- **Grade C (50-69%)**: Restores with foreign key violations or schema mismatch.
- **Grade F (< 50%)**: Restoration failed, file corrupted, or operation timed out.

## License

MIT License. See [LICENSE](LICENSE) for details.
