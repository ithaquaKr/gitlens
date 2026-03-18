package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"gitlens/internal/config"
)

var (
	cfgFile      string
	providerFlag string
	apiKeyFlag   string
	modelFlag    string
	themeFlag    string

	// Cfg is loaded once at PersistentPreRunE and shared across subcommands.
	Cfg *config.Config
)

var rootCmd = &cobra.Command{
	Use:   "gitlens",
	Short: "AI-powered git CLI tool",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		var err error
		Cfg, err = config.Load(config.LoadOptions{
			ExplicitPath: cfgFile,
			Overrides: config.Overrides{
				Provider: providerFlag,
				APIKey:   apiKeyFlag,
				Model:    modelFlag,
				Theme:    themeFlag,
			},
		})
		return err
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file path")
	rootCmd.PersistentFlags().StringVarP(&providerFlag, "provider", "p", "", "AI provider (claude, gemini)")
	rootCmd.PersistentFlags().StringVarP(&apiKeyFlag, "api-key", "k", "", "API key override")
	rootCmd.PersistentFlags().StringVarP(&modelFlag, "model", "m", "", "model name override")
	rootCmd.PersistentFlags().StringVar(&themeFlag, "theme", "", "color theme override")
}
