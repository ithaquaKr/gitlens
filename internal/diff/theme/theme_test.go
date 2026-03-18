package theme_test

import (
	"testing"
	"gitlens/internal/config"
	"gitlens/internal/diff/theme"
)

func TestLoadDefaultPreset(t *testing.T) {
	cfg := &config.Config{Theme: config.ThemeConfig{Base: "dark"}}
	th := theme.Load(cfg)
	if th.Syntax.Keyword == "" {
		t.Error("expected non-empty keyword color")
	}
}

func TestLoadWithOverride(t *testing.T) {
	cfg := &config.Config{
		Theme: config.ThemeConfig{
			Base:     "dark",
			Override: config.ThemeOverride{Keyword: "#ff0000"},
		},
	}
	th := theme.Load(cfg)
	if th.Syntax.Keyword != "#ff0000" {
		t.Errorf("override not applied: got %q", th.Syntax.Keyword)
	}
}

func TestAllPresetsLoad(t *testing.T) {
	presets := []string{"dark", "light", "catppuccin-mocha", "catppuccin-latte",
		"dracula", "nord", "gruvbox-dark", "gruvbox-light", "one-dark",
		"solarized-dark", "solarized-light"}
	for _, name := range presets {
		cfg := &config.Config{Theme: config.ThemeConfig{Base: name}}
		th := theme.Load(cfg)
		if th.Syntax.Keyword == "" {
			t.Errorf("preset %q: empty keyword color", name)
		}
	}
}
