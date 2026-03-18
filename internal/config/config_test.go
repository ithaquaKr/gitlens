package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"gitlens/internal/config"
)

func TestLoadDefaults(t *testing.T) {
	cfg, err := config.Load(config.LoadOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Provider != "claude" {
		t.Errorf("default provider: got %q, want %q", cfg.Provider, "claude")
	}
}

func TestLoadFromFile(t *testing.T) {
	dir := t.TempDir()
	content := `
provider = "gemini"
api_key  = "test-key"
model    = "gemini-pro"
`
	path := filepath.Join(dir, "config.toml")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	cfg, err := config.Load(config.LoadOptions{ExplicitPath: path})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Provider != "gemini" {
		t.Errorf("provider: got %q, want %q", cfg.Provider, "gemini")
	}
	if cfg.APIKey != "test-key" {
		t.Errorf("api_key: got %q, want %q", cfg.APIKey, "test-key")
	}
}

func TestEnvVarOverridesFileConfig(t *testing.T) {
	dir := t.TempDir()
	content := "provider = \"gemini\"\napi_key = \"file-key\"\n"
	path := filepath.Join(dir, "config.toml")
	os.WriteFile(path, []byte(content), 0o600)

	t.Setenv("GITLENS_API_KEY", "env-key")

	cfg, err := config.Load(config.LoadOptions{ExplicitPath: path})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.APIKey != "env-key" {
		t.Errorf("env var should override file: got %q, want %q", cfg.APIKey, "env-key")
	}
}

func TestCLIOverrideOverridesAll(t *testing.T) {
	t.Setenv("GITLENS_PROVIDER", "gemini")
	cfg, err := config.Load(config.LoadOptions{
		Overrides: config.Overrides{Provider: "claude"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Provider != "claude" {
		t.Errorf("CLI override: got %q, want %q", cfg.Provider, "claude")
	}
}
