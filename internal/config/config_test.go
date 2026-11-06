package config

import (
	"os"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.Server.Port != 8080 {
		t.Fatalf("expected port 8080, got %d", cfg.Server.Port)
	}
	if len(cfg.Drills) == 0 {
		t.Fatalf("expected at least one default drill")
	}
}

func TestExpandEnvVars(t *testing.T) {
	os.Setenv("TEST_RESTORECRAFT_VAR", "test-value-123")
	defer os.Unsetenv("TEST_RESTORECRAFT_VAR")

	input := "secret: ${TEST_RESTORECRAFT_VAR}"
	out := expandEnvVars(input)
	if out != "secret: test-value-123" {
		t.Fatalf("expected secret: test-value-123, got %s", out)
	}
}
