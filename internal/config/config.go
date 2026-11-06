package config

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/alexandrmotologa/restorecraft/internal/models"
	"gopkg.in/yaml.v3"
)

// Config is the root application configuration.
type Config struct {
	Version       string               `yaml:"version"`
	Server        ServerConfig         `yaml:"server"`
	Security      SecurityConfig       `yaml:"security"`
	Defaults      DefaultsConfig       `yaml:"defaults"`
	Storage       StorageConfig        `yaml:"storage"`
	Drills        []DrillConfig        `yaml:"drills"`
	Notifications NotificationConfig   `yaml:"notifications"`
}

type ServerConfig struct {
	Host    string `yaml:"host"`
	Port    int    `yaml:"port"`
	DataDir string `yaml:"data_dir"`
}

type SecurityConfig struct {
	SigningSecret  string `yaml:"signing_secret"`
	SignerIdentity string `yaml:"signer_identity"`
}

type DefaultsConfig struct {
	RtoSlaSeconds    int  `yaml:"rto_sla_seconds"`
	SandboxMemoryMb  int  `yaml:"sandbox_memory_mb"`
	CleanupOnFailure bool `yaml:"cleanup_on_failure"`
}

type StorageConfig struct {
	Local LocalStorageConfig `yaml:"local"`
	S3    S3StorageConfig    `yaml:"s3"`
}

type LocalStorageConfig struct {
	Enabled bool   `yaml:"enabled"`
	BaseDir string `yaml:"base_dir"`
}

type S3StorageConfig struct {
	Enabled         bool   `yaml:"enabled"`
	Endpoint        string `yaml:"endpoint"`
	Region          string `yaml:"region"`
	Bucket          string `yaml:"bucket"`
	Prefix          string `yaml:"prefix"`
	AccessKeyID     string `yaml:"access_key_id"`
	SecretAccessKey string `yaml:"secret_access_key"`
	ForcePathStyle  bool   `yaml:"force_path_style"`
}

type ParityConfig struct {
	ExpectedTables []string                           `yaml:"expected_tables"`
	MinRowCounts   map[string]int64                   `yaml:"min_row_counts"`
	ExpectedSchema map[string]models.TableSchema      `yaml:"expected_schema"`
}

type DockerSandboxConfig struct {
	Image     string `yaml:"image"`
	MemoryMb  int    `yaml:"memory_mb"`
	ShmSizeMb int    `yaml:"shm_size_mb"`
}

type DrillConfig struct {
	ID             string              `yaml:"id"`
	Name           string              `yaml:"name"`
	Engine         string              `yaml:"engine"` // "sqlite", "postgres", "mysql"
	Schedule       string              `yaml:"schedule"` // cron expression e.g. "0 3 * * *"
	StorageType    string              `yaml:"storage_type"` // "local", "s3"
	SnapshotPath   string              `yaml:"snapshot_path"`
	RtoSlaSeconds  int                 `yaml:"rto_sla_seconds"`
	HoldMinutes    int                 `yaml:"hold_minutes"`    // Keep sandbox alive for debugging
	HoldOnFailure  bool                `yaml:"hold_on_failure"` // Hold sandbox if drill fails
	AuditPii       bool                `yaml:"audit_pii"`       // Assert data sanitization / masking
	Docker         DockerSandboxConfig `yaml:"docker"`
	Parity         ParityConfig        `yaml:"parity"`
}

type NotificationConfig struct {
	Telegram TelegramNotification `yaml:"telegram"`
	Slack    SlackNotification    `yaml:"slack"`
	Discord  DiscordNotification  `yaml:"discord"`
}

type TelegramNotification struct {
	Enabled  bool   `yaml:"enabled"`
	BotToken string `yaml:"bot_token"`
	ChatID   string `yaml:"chat_id"`
	NotifyOn string `yaml:"notify_on"` // "all", "failures_only"
}

type SlackNotification struct {
	Enabled    bool   `yaml:"enabled"`
	WebhookURL string `yaml:"webhook_url"`
	NotifyOn   string `yaml:"notify_on"`
}

type DiscordNotification struct {
	Enabled    bool   `yaml:"enabled"`
	WebhookURL string `yaml:"webhook_url"`
	NotifyOn   string `yaml:"notify_on"`
}

// DefaultConfig returns a sane base configuration.
func DefaultConfig() *Config {
	return &Config{
		Version: "1",
		Server: ServerConfig{
			Host:    "127.0.0.1",
			Port:    8080,
			DataDir: "./data",
		},
		Security: SecurityConfig{
			SigningSecret:  "restorecraft-default-development-secret-key-32b",
			SignerIdentity: "restorecraft-drill-authority",
		},
		Defaults: DefaultsConfig{
			RtoSlaSeconds:    300,
			SandboxMemoryMb:  2048,
			CleanupOnFailure: true,
		},
		Storage: StorageConfig{
			Local: LocalStorageConfig{
				Enabled: true,
				BaseDir: "./backups",
			},
			S3: S3StorageConfig{
				Enabled: false,
				Region:  "us-east-1",
			},
		},
		Drills: []DrillConfig{
			{
				ID:            "demo-sqlite-drill",
				Name:          "Default SQLite Snapshot Verification",
				Engine:        "sqlite",
				StorageType:   "local",
				SnapshotPath:  "./test-backups/valid_ecommerce.sqlite",
				RtoSlaSeconds: 60,
				Parity: ParityConfig{
					ExpectedTables: []string{"users", "orders", "order_items"},
					MinRowCounts: map[string]int64{
						"users":  5,
						"orders": 5,
					},
				},
			},
		},
	}
}

// LoadConfig loads and parses a YAML configuration file with env substitution.
func LoadConfig(path string) (*Config, error) {
	cfg := DefaultConfig()

	if path == "" {
		// Look for default file
		candidates := []string{"restorecraft.yaml", "restorecraft.yml", "config.yaml", "config.yml"}
		for _, c := range candidates {
			if _, err := os.Stat(c); err == nil {
				path = c
				break
			}
		}
	}

	if path != "" {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read config file: %w", err)
		}

		interpolated := expandEnvVars(string(data))
		if err := yaml.Unmarshal([]byte(interpolated), cfg); err != nil {
			return nil, fmt.Errorf("parse config yaml: %w", err)
		}
	}

	// Apply environment variable overrides
	if val := os.Getenv("RESTORECRAFT_SERVER_HOST"); val != "" {
		cfg.Server.Host = val
	}
	if val := os.Getenv("RESTORECRAFT_SIGNING_SECRET"); val != "" {
		cfg.Security.SigningSecret = val
	}
	if val := os.Getenv("RESTORECRAFT_DATA_DIR"); val != "" {
		cfg.Server.DataDir = val
	}

	return cfg, nil
}

// expandEnvVars expands ${ENV_VAR} or $ENV_VAR inside text.
func expandEnvVars(text string) string {
	re := regexp.MustCompile(`\$\{([A-Za-z0-9_]+)\}`)
	return re.ReplaceAllStringFunc(text, func(m string) string {
		varName := strings.TrimSuffix(strings.TrimPrefix(m, "${"), "}")
		if val, exists := os.LookupEnv(varName); exists {
			return val
		}
		return m
	})
}
