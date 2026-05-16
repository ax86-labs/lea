package architecture

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

func LoadConfig(path string) (*Config, error) {
	if path == "" {
		return nil, fmt.Errorf("architecture config path is required")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	applyDefaults(&cfg)
	if err := validateConfig(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func applyDefaults(cfg *Config) {
	if cfg.Settings.AllowSelf == nil {
		value := true
		cfg.Settings.AllowSelf = &value
	}
	if cfg.Settings.AllowUnknown == nil {
		value := true
		cfg.Settings.AllowUnknown = &value
	}
	if cfg.Settings.DefaultAllowAll == nil {
		value := true
		cfg.Settings.DefaultAllowAll = &value
	}
}

func validateConfig(cfg *Config) error {
	seen := make(map[string]bool)
	for _, layer := range cfg.Layers {
		if layer.Name == "" {
			return fmt.Errorf("layer name cannot be empty")
		}
		if seen[layer.Name] {
			return fmt.Errorf("duplicate layer name: %s", layer.Name)
		}
		seen[layer.Name] = true
	}
	return nil
}
