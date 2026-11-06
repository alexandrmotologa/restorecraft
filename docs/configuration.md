# Configuration Guide

`restorecraft` is configured through a YAML file or equivalent environment variables. By default, it looks for `restorecraft.yaml` in the current working directory, or at a path supplied with `--config`.

## Example Configuration File

```yaml
version: "1"

server:
  host: "127.0.0.1"
  port: 8080
  data_dir: "./data"

security:
  signing_secret: "replace-with-a-high-entropy-secret-key-at-least-32-chars"

defaults:
  rto_sla_seconds: 300
  sandbox_memory_mb: 2048
  cleanup_on_failure: true

storage:
  local:
    enabled: true
    base_dir: "./backups"

  s3:
    enabled: false
    endpoint: "https://s3.amazonaws.com"
    region: "us-east-1"
    bucket: "production-backups"
    prefix: "daily/"
    access_key_id: "${AWS_ACCESS_KEY_ID}"
    secret_access_key: "${AWS_SECRET_ACCESS_KEY}"
    force_path_style: false

drills:
  - id: "production-sqlite-daily"
    name: "Production App SQLite Snapshot"
    engine: "sqlite"
    schedule: "0 3 * * *"
    storage_type: "local"
    snapshot_path: "app_production.sqlite.gz"
    rto_sla_seconds: 60
    parity:
      expected_tables:
        - "users"
        - "orders"
        - "order_items"
        - "audit_logs"
      min_row_counts:
        users: 100
        orders: 500

  - id: "warehouse-postgres-drill"
    name: "Analytics PostgreSQL Dump"
    engine: "postgres"
    schedule: "0 4 * * 0"
    storage_type: "s3"
    snapshot_path: "warehouse_latest.dump"
    rto_sla_seconds: 900
    docker:
      image: "postgres:16-alpine"
      memory_mb: 4096
      shm_size_mb: 1024
    parity:
      expected_tables:
        - "events"
        - "user_traits"
        - "metrics"

notifications:
  telegram:
    enabled: false
    bot_token: "${TELEGRAM_BOT_TOKEN}"
    chat_id: "${TELEGRAM_CHAT_ID}"
    notify_on: "all" # "all", "failures_only"

  slack:
    enabled: false
    webhook_url: "${SLACK_WEBHOOK_URL}"
    notify_on: "failures_only"

  discord:
    enabled: false
    webhook_url: "${DISCORD_WEBHOOK_URL}"
    notify_on: "all"
```

## Environment Variable Overrides

Any setting can be overridden via environment variables:

- `RESTORECRAFT_SERVER_PORT`: Overrides the HTTP web port (e.g. `8080`).
- `RESTORECRAFT_SERVER_HOST`: Overrides the HTTP listen interface (e.g. `0.0.0.0`).
- `RESTORECRAFT_SIGNING_SECRET`: Sets the HMAC key used to sign Disaster Recovery Certificates.
- `RESTORECRAFT_DATA_DIR`: Directory where drill history and SQLite state are stored.
- Standard AWS environment variables (`AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, `AWS_REGION`, `AWS_ENDPOINT_URL`) are read by the S3 storage connector when configured.
