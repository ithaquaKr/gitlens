package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

var configureCmd = &cobra.Command{
	Use:   "configure",
	Short: "Interactive setup wizard",
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			provider string
			apiKey   string
			model    string
			theme    string
		)

		providerDefaults := map[string]string{
			"claude": "claude-opus-4-6",
			"gemini": "gemini-2.0-flash",
		}

		form := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("AI Provider").
					Options(
						huh.NewOption("Claude (Anthropic)", "claude"),
						huh.NewOption("Gemini (Google)", "gemini"),
					).
					Value(&provider),
			),
			huh.NewGroup(
				huh.NewInput().
					Title("API Key").
					Description("Leave empty to use environment variable (GITLENS_API_KEY)").
					EchoMode(huh.EchoModePassword).
					Value(&apiKey),
				huh.NewInput().
					Title("Model name").
					Description("Leave empty for default").
					Value(&model),
			),
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Color theme").
					Options(
						huh.NewOption("Dark", "dark"),
						huh.NewOption("Light", "light"),
						huh.NewOption("Catppuccin Mocha", "catppuccin-mocha"),
						huh.NewOption("Catppuccin Latte", "catppuccin-latte"),
						huh.NewOption("Dracula", "dracula"),
						huh.NewOption("Nord", "nord"),
						huh.NewOption("Gruvbox Dark", "gruvbox-dark"),
						huh.NewOption("Gruvbox Light", "gruvbox-light"),
						huh.NewOption("One Dark", "one-dark"),
						huh.NewOption("Solarized Dark", "solarized-dark"),
						huh.NewOption("Solarized Light", "solarized-light"),
					).
					Value(&theme),
			),
		)

		if err := form.Run(); err != nil {
			return err
		}

		if model == "" {
			model = providerDefaults[provider]
		}

		type themeSection struct {
			Base string `toml:"base"`
		}
		type cfg struct {
			Provider string       `toml:"provider"`
			APIKey   string       `toml:"api_key,omitempty"`
			Model    string       `toml:"model"`
			Theme    themeSection `toml:"theme"`
		}

		out := cfg{
			Provider: provider,
			Model:    model,
			Theme:    themeSection{Base: theme},
		}
		if apiKey != "" {
			out.APIKey = apiKey
		}

		configDir := filepath.Join(mustHomeDir(), ".config", "gitlens")
		if err := os.MkdirAll(configDir, 0o755); err != nil {
			return fmt.Errorf("creating config dir: %w", err)
		}
		configPath := filepath.Join(configDir, "config.toml")
		f, err := os.Create(configPath)
		if err != nil {
			return fmt.Errorf("creating config file: %w", err)
		}
		defer f.Close()

		if err := toml.NewEncoder(f).Encode(out); err != nil {
			return fmt.Errorf("writing config: %w", err)
		}

		fmt.Printf("\nConfig saved to %s\n", configPath)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(configureCmd)
}

func mustHomeDir() string {
	h, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	return h
}
