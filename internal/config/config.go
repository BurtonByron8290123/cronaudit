package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Job describes a single monitored cron job.
type Job struct {
	Name     string `yaml:"name"`
	Schedule string `yaml:"schedule"`
	Command  string `yaml:"command"`
	DriftSec int    `yaml:"drift_sec"`
}

// Webhook holds alert destination configuration.
type Webhook struct {
	URL string `yaml:"url"`
}

// Config is the top-level configuration structure.
type Config struct {
	Jobs    []Job   `yaml:"jobs"`
	Webhook Webhook `yaml:"webhook"`
	LogFile string  `yaml:"log_file"`
}

// Load reads and validates a YAML config file from the given path.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("config: read file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("config: parse yaml: %w", err)
	}

	for i, j := range cfg.Jobs {
		if j.Name == "" {
			return nil, fmt.Errorf("config: job[%d] missing required field 'name'", i)
		}
	}

	return &cfg, nil
}
