package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadUsesDefaultsWhenConfigFileIsMissing(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("change working directory: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Fatalf("restore working directory: %v", err)
		}
	})

	cfg, err := Load()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if cfg.App.Environment != "development" {
		t.Fatalf("expected default environment development, got %q", cfg.App.Environment)
	}
	if cfg.Server.Host != "0.0.0.0" {
		t.Fatalf("expected default host 0.0.0.0, got %q", cfg.Server.Host)
	}
	if cfg.Server.Port != 8080 {
		t.Fatalf("expected default port 8080, got %d", cfg.Server.Port)
	}
}

func TestLoadReturnsErrorForInvalidConfig(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}

	if err := os.WriteFile(filepath.Join(tmpDir, "config.yaml"), []byte("app: [invalid"), 0o600); err != nil {
		t.Fatalf("write invalid config: %v", err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("change working directory: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Fatalf("restore working directory: %v", err)
		}
	})

	if _, err := Load(); err == nil {
		t.Fatal("expected invalid config to return an error")
	}
}

func TestServerAddress(t *testing.T) {
	cfg := ServerConfig{Host: "127.0.0.1", Port: 3000}

	if got := cfg.Address(); got != "127.0.0.1:3000" {
		t.Fatalf("expected address 127.0.0.1:3000, got %q", got)
	}
}
