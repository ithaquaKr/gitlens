package config

import (
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// Config is the fully-resolved configuration after applying precedence rules.
type Config struct {
	Provider string
	APIKey   string
	Model    string
	Theme    ThemeConfig
	Draft    DraftConfig
}

type ThemeConfig struct {
	Base     string
	Override ThemeOverride
}

type ThemeOverride struct {
	// Syntax colors
	Keyword         string `toml:"keyword"`
	String          string `toml:"string"`
	Comment         string `toml:"comment"`
	Function        string `toml:"function"`
	FunctionMacro   string `toml:"function_macro"`
	Type            string `toml:"type"`
	Number          string `toml:"number"`
	Operator        string `toml:"operator"`
	Variable        string `toml:"variable"`
	VariableBuiltin string `toml:"variable_builtin"`
	VariableMember  string `toml:"variable_member"`
	Module          string `toml:"module"`
	Tag             string `toml:"tag"`
	Attribute       string `toml:"attribute"`
	Label           string `toml:"label"`
	Punctuation     string `toml:"punctuation"`
	// Diff colors
	AddedBg       string `toml:"added_bg"`
	DeletedBg     string `toml:"deleted_bg"`
	AddedWordBg   string `toml:"added_word_bg"`
	DeletedWordBg string `toml:"deleted_word_bg"`
	// UI colors
	Border    string `toml:"border"`
	Selection string `toml:"selection"`
}

type DraftConfig struct {
	CommitTypes map[string]string `toml:"commit_types"`
}

// fileConfig mirrors the TOML file structure for unmarshalling.
type fileConfig struct {
	Provider string      `toml:"provider"`
	APIKey   string      `toml:"api_key"`
	Model    string      `toml:"model"`
	Theme    ThemeConfig `toml:"theme"`
	Draft    DraftConfig `toml:"draft"`
}

// Overrides holds values passed via CLI flags.
type Overrides struct {
	Provider string
	APIKey   string
	Model    string
	Theme    string
}

// LoadOptions controls how config is loaded.
type LoadOptions struct {
	ExplicitPath string // --config flag
	Overrides    Overrides
}

// Load resolves config using the precedence chain:
// CLI flags > env vars > explicit file > project file > global file > defaults
func Load(opts LoadOptions) (*Config, error) {
	cfg := defaults()

	// Layer: global file
	if global, err := globalConfigPath(); err == nil {
		_ = mergeFile(cfg, global) // ignore missing file
	}

	// Layer: project file
	_ = mergeFile(cfg, "gitlens.config.toml")

	// Layer: explicit file
	if opts.ExplicitPath != "" {
		if err := mergeFile(cfg, opts.ExplicitPath); err != nil {
			return nil, err
		}
	}

	// Layer: env vars
	applyEnv(cfg)

	// Layer: CLI overrides
	if opts.Overrides.Provider != "" {
		cfg.Provider = opts.Overrides.Provider
	}
	if opts.Overrides.APIKey != "" {
		cfg.APIKey = opts.Overrides.APIKey
	}
	if opts.Overrides.Model != "" {
		cfg.Model = opts.Overrides.Model
	}
	if opts.Overrides.Theme != "" {
		cfg.Theme.Base = opts.Overrides.Theme
	}

	return cfg, nil
}

func defaults() *Config {
	return &Config{
		Provider: "claude",
		Model:    "claude-opus-4-6",
		Theme:    ThemeConfig{Base: "dark"},
		Draft: DraftConfig{
			CommitTypes: map[string]string{
				"feat":     "A new feature",
				"fix":      "A bug fix",
				"docs":     "Documentation changes",
				"refactor": "Code refactoring",
				"test":     "Adding tests",
				"chore":    "Maintenance tasks",
			},
		},
	}
}

func mergeFile(cfg *Config, path string) error {
	var fc fileConfig
	if _, err := toml.DecodeFile(path, &fc); err != nil {
		return err
	}
	if fc.Provider != "" {
		cfg.Provider = fc.Provider
	}
	if fc.APIKey != "" {
		cfg.APIKey = fc.APIKey
	}
	if fc.Model != "" {
		cfg.Model = fc.Model
	}
	if fc.Theme.Base != "" {
		cfg.Theme.Base = fc.Theme.Base
	}
	cfg.Theme.Override = fc.Theme.Override
	if len(fc.Draft.CommitTypes) > 0 {
		cfg.Draft.CommitTypes = fc.Draft.CommitTypes
	}
	return nil
}

func applyEnv(cfg *Config) {
	if v := os.Getenv("GITLENS_PROVIDER"); v != "" {
		cfg.Provider = v
	}
	if v := os.Getenv("GITLENS_API_KEY"); v != "" {
		cfg.APIKey = v
	}
	if v := os.Getenv("GITLENS_MODEL"); v != "" {
		cfg.Model = v
	}
	if v := os.Getenv("GITLENS_THEME"); v != "" {
		cfg.Theme.Base = v
	}
}

func globalConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "gitlens", "config.toml"), nil
}
