package config

import (
	"os"
	"runtime"
	"testing"
)

func TestConfigLoadAndSave(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)
	t.Setenv("XDG_CONFIG_HOME", tempHome)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load config failed: %v", err)
	}

	if !cfg.AI.Enabled {
		t.Errorf("expected default AI to be enabled")
	}

	// Modify config
	cfg.AI.Model = "custom-coder:14b"
	cfg.AI.BaseURL = "http://localhost:8000/v1"

	err = Save(cfg)
	if err != nil {
		t.Fatalf("Save config failed: %v", err)
	}

	// Reload config
	reloaded, err := Load()
	if err != nil {
		t.Fatalf("Reload config failed: %v", err)
	}

	if reloaded.AI.Model != "custom-coder:14b" {
		t.Errorf("expected reloaded model 'custom-coder:14b', got '%s'", reloaded.AI.Model)
	}
	if reloaded.AI.BaseURL != "http://localhost:8000/v1" {
		t.Errorf("expected reloaded base URL 'http://localhost:8000/v1', got '%s'", reloaded.AI.BaseURL)
	}

	// Verify file permissions (0600 on POSIX platforms)
	if runtime.GOOS != "windows" {
		configPath, _ := GetConfigPath()
		info, err := os.Stat(configPath)
		if err != nil {
			t.Fatalf("stat on config file failed: %v", err)
		}
		perm := info.Mode().Perm()
		if perm != 0600 {
			t.Errorf("expected file permissions 0600, got %o", perm)
		}
	}
}
