package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"gitlens/internal/ai"
	"gitlens/internal/git_entity"
	"gitlens/internal/vcs"
)

var draftContext string

var draftCmd = &cobra.Command{
	Use:   "draft",
	Short: "Generate a conventional commit message for staged changes",
	RunE: func(cmd *cobra.Command, args []string) error {
		backend, err := vcs.NewGitBackend(".")
		if err != nil {
			return err
		}
		diff, err := backend.GetWorkingTreeDiff(true) // staged only
		if err != nil {
			return err
		}
		if len(diff.Files) == 0 {
			return fmt.Errorf("no staged changes found")
		}

		diffText := buildDiffText(diff)

		provider, err := ai.New(Cfg)
		if err != nil {
			return err
		}

		prompt := ai.DraftPrompt(diffText, draftContext, Cfg.Draft.CommitTypes)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		ch, err := provider.Stream(ctx, prompt)
		if err != nil {
			fmt.Fprintf(os.Stderr, "stream unavailable, falling back: %v\n", err)
			result, err := provider.Complete(ctx, prompt)
			if err != nil {
				return err
			}
			fmt.Print(result)
			return nil
		}
		for chunk := range ch {
			if chunk.Err != nil {
				return chunk.Err
			}
			fmt.Print(chunk.Text)
		}
		if isTerminal(os.Stdout) {
			fmt.Println()
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(draftCmd)
	draftCmd.Flags().StringVar(&draftContext, "context", "", "additional context about the intent of these changes")
}

func buildDiffText(diff *git_entity.Diff) string {
	var sb strings.Builder
	for _, f := range diff.Files {
		sb.WriteString(fmt.Sprintf("--- a/%s\n+++ b/%s\n", f.Path, f.Path))
		switch f.Status {
		case "A":
			// New file: all lines are additions
			for _, line := range strings.Split(strings.TrimRight(f.NewContent, "\n"), "\n") {
				sb.WriteString("+" + line + "\n")
			}
		case "D":
			// Deleted file: all lines are removals
			for _, line := range strings.Split(strings.TrimRight(f.OldContent, "\n"), "\n") {
				sb.WriteString("-" + line + "\n")
			}
		default:
			// Modified: show removed old lines then added new lines
			oldLines := strings.Split(strings.TrimRight(f.OldContent, "\n"), "\n")
			newLines := strings.Split(strings.TrimRight(f.NewContent, "\n"), "\n")
			sb.WriteString(fmt.Sprintf("@@ -%d,%d +%d,%d @@\n",
				1, len(oldLines), 1, len(newLines)))
			for _, line := range oldLines {
				sb.WriteString("-" + line + "\n")
			}
			for _, line := range newLines {
				sb.WriteString("+" + line + "\n")
			}
		}
	}
	return sb.String()
}

func isTerminal(f *os.File) bool {
	fi, err := f.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}
