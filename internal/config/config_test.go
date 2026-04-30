package config

import (
	"os"
	"testing"
	"time"
)

func writeTempConfig(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "cronaudit-*.yaml")
	if err != nil {
		t.Fatalf("creating temp file: %v", err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("writing temp file: %v", err)
	}
	f.Close()
	return f.Name()
}

func TestLoad_ValidConfig(t *testing.T) {
	content := `
log_level: info
jobs:
  - name: backup
    schedule: "0 2 * * *"
    timeout: 30m
    command: /usr/local/bin/backup.sh
alerts:
  slack_webhook: https://hooks.slack.com/xxx
  email: ops@example.com
`
	path := writeTempConfig(t, content)
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.LogLevel != "info" {
		t.Errorf("expected log_level=info, got %q", cfg.LogLevel)
	}
	if len(cfg.Jobs) != 1 {
		t.Fatalf("expected 1 job, got %d", len(cfg.Jobs))
	}
	if cfg.Jobs[0].Name != "backup" {
		t.Errorf("expected job name=backup, got %q", cfg.Jobs[0].Name)
	}
	if cfg.Jobs[0].Timeout != 30*time.Minute {
		t.Errorf("expected timeout=30m, got %v", cfg.Jobs[0].Timeout)
	}
}

func TestLoad_MissingName(t *testing.T) {
	content := `
jobs:
  - schedule: "0 2 * * *"
    command: /usr/local/bin/backup.sh
`
	path := writeTempConfig(t, content)
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected validation error for missing name")
	}
}

func TestLoad_FileNotFound(t *testing.T) {
	_, err := Load("/nonexistent/path/config.yaml")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestLoad_InvalidYAML(t *testing.T) {
	path := writeTempConfig(t, ": invalid: yaml: [")
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
}
