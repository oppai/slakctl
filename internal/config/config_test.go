package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	t.Run("should return empty config when file doesn't exist", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "slakctl-test-*")
		if err != nil {
			t.Fatalf("failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tempDir)

		originalHome := os.Getenv("HOME")
		os.Setenv("HOME", tempDir)
		defer os.Setenv("HOME", originalHome)

		config, err := LoadConfig()
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		if config.Token != "" {
			t.Errorf("expected empty token, got: %s", config.Token)
		}
	})

	t.Run("should load config from file", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "slakctl-test-*")
		if err != nil {
			t.Fatalf("failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tempDir)

		originalHome := os.Getenv("HOME")
		os.Setenv("HOME", tempDir)
		defer os.Setenv("HOME", originalHome)

		configPath := filepath.Join(tempDir, ".slakctl")
		configContent := `{"token": "test-token"}`
		if err := os.WriteFile(configPath, []byte(configContent), 0600); err != nil {
			t.Fatalf("failed to write config file: %v", err)
		}

		config, err := LoadConfig()
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		if config.Token != "test-token" {
			t.Errorf("expected token 'test-token', got: %s", config.Token)
		}
	})
}

func TestSaveConfig(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "slakctl-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	config := &Config{
		Token: "test-token",
	}

	if err := SaveConfig(config); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	configPath := filepath.Join(tempDir, ".slakctl")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatal("config file was not created")
	}

	loadedConfig, err := LoadConfig()
	if err != nil {
		t.Fatalf("failed to load saved config: %v", err)
	}

	if loadedConfig.Token != "test-token" {
		t.Errorf("expected token 'test-token', got: %s", loadedConfig.Token)
	}
}