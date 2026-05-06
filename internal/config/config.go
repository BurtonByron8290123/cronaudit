package config

import (
	"errors"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Job describes a single cron job to monitor.
type Job struct {
	Name     string `yaml:"name"`
	Schedule string `yaml:"schedule"`
	Command  string `yaml:"command"`
	Timeout  int    `yaml:"timeout_seconds"`
}

// Config is the top-level configuration structure.
type Config struct {
	SyslogPath   string `yaml:"syslog_path"`
	PollInterval int    `yaml:"poll_interval_ms"`
	WebhookURL   string `yaml:"webhook_url"`
	Jobs         []Job  `yaml:"jobs"`
}

// Load reads and validates a YAML config file at the given path.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("config: read file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("config: parse yaml: %w", err)
	}

	if err := validate(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func validate(cfg *Config) error {
	for i, j := range cfg.Jobs {
		if j.Name == "" {
			return errors.New(fmt.Sprintf("config: job[%d] missing name", i))
		}
	}
	return nil
}
