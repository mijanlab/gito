package config

import (
	"os"
	"path/filepath"

	"gito/internal/ai"
	"gopkg.in/yaml.v3"
)

// AppConfig represents global user configuration for Gito.
type AppConfig struct {
	AI ai.Config `yaml:"ai"`
}

// DefaultAppConfig returns default application configuration.
func DefaultAppConfig() *AppConfig {
	return &AppConfig{
		AI: ai.DefaultConfig(),
	}
}

// GetConfigPath returns the absolute path to the gito config file.
func GetConfigPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		configDir = filepath.Join(home, ".config")
	}
	return filepath.Join(configDir, "gito", "config.yaml"), nil
}

// Load reads and parses the configuration file, returning default config if none exists.
func Load() (*AppConfig, error) {
	cfg := DefaultAppConfig()

	path, err := GetConfigPath()
	if err != nil {
		return cfg, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return cfg, err
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return cfg, err
	}

	return cfg, nil
}

// Save writes configuration to disk with secure 0600 file permissions.
func Save(cfg *AppConfig) error {
	path, err := GetConfigPath()
	if err != nil {
		return err
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600)
}
