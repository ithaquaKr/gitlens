package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
	"gitlens/internal/ai"
	"gitlens/internal/git_entity"
	"gitlens/internal/vcs"
)

var (
	explainStaged bool
	explainQuery  string
	explainList   bool
)

var explainCmd = &cobra.Command{
	Use:   "explain [ref|-]",
	Short: "Explain git changes using AI",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		backend, err := vcs.NewGitBackend(".")
		if err != nil {
			return err
		}

		ref := ""
		if len(args) > 0 {
			ref = args[0]
		}

		// Resolve ref from --list or stdin "-"
		if explainList {
			ref, err = fzfSelectCommit()
			if err != nil {
				return err
			}
		} else if ref == "-" {
			scanner := bufio.NewScanner(os.Stdin)
			scanner.Scan()
			if err := scanner.Err(); err != nil {
				return fmt.Errorf("reading stdin: %w", err)
			}
			ref = strings.TrimSpace(scanner.Text())
			if ref == "" {
				return fmt.Errorf("no ref received from stdin")
			}
		}

		var diff *git_entity.Diff
		if ref != "" {
			parsed := vcs.ParseRef(ref)
			if parsed.Single != "" {
				diff, err = backend.GetRangeDiff(parsed.Single+"^", parsed.Single, false)
				if err != nil {
					// Initial commit has no parent — return empty diff
					diff = &git_entity.Diff{}
				}
			} else {
				diff, err = backend.GetRangeDiff(parsed.From, parsed.To, parsed.ThreeDot)
				if err != nil {
					return err
				}
			}
		} else {
			diff, err = backend.GetWorkingTreeDiff(explainStaged)
			if err != nil {
				return err
			}
		}

		diffText := diffToText(diff)
		if diffText == "" {
			return fmt.Errorf("no changes to explain")
		}

		provider, err := ai.New(Cfg)
		if err != nil {
			return err
		}

		prompt := ai.ExplainPrompt(diffText, explainQuery)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		ch, err := provider.Stream(ctx, prompt)
		if err != nil {
			fmt.Fprintf(os.Stderr, "stream unavailable, falling back: %v\n", err)
			result, err := provider.Complete(ctx, prompt)
			if err != nil {
				return err
			}
			return printWithMdcat(result)
		}

		var sb strings.Builder
		for chunk := range ch {
			if chunk.Err != nil {
				return chunk.Err
			}
			sb.WriteString(chunk.Text)
		}
		return printWithMdcat(sb.String())
	},
}

func init() {
	rootCmd.AddCommand(explainCmd)
	explainCmd.Flags().BoolVar(&explainStaged, "staged", false, "explain staged changes only")
	explainCmd.Flags().StringVar(&explainQuery, "query", "", "ask a specific question about the changes")
	explainCmd.Flags().BoolVar(&explainList, "list", false, "interactively select a commit via fzf")
}

// fzfSelectCommit shells out to fzf with git log output.
func fzfSelectCommit() (string, error) {
	if _, err := exec.LookPath("fzf"); err != nil {
		return "", fmt.Errorf("explain --list requires fzf. Install it from https://github.com/junegunn/fzf")
	}
	logCmd := exec.Command("git", "log", "--oneline", "--color=always")
	fzfCmd := exec.Command("fzf", "--ansi", "--reverse")
	fzfCmd.Stdin, _ = logCmd.StdoutPipe()
	fzfCmd.Stderr = os.Stderr
	if err := logCmd.Start(); err != nil {
		return "", fmt.Errorf("git log: %w", err)
	}
	out, err := fzfCmd.Output()
	waitErr := logCmd.Wait()
	if err != nil {
		return "", fmt.Errorf("fzf: %w", err)
	}
	if waitErr != nil {
		fmt.Fprintf(os.Stderr, "git log: %v\n", waitErr)
	}
	fields := strings.Fields(strings.TrimSpace(string(out)))
	if len(fields) == 0 {
		return "", fmt.Errorf("no commit selected")
	}
	return fields[0], nil // first field is the short SHA
}

// printWithMdcat renders markdown via mdcat if available, else plain stdout.
func printWithMdcat(text string) error {
	if _, err := exec.LookPath("mdcat"); err != nil {
		fmt.Println(text)
		return nil
	}
	mdCmd := exec.Command("mdcat")
	mdCmd.Stdin = strings.NewReader(text)
	mdCmd.Stdout = os.Stdout
	mdCmd.Stderr = os.Stderr
	return mdCmd.Run()
}

// diffToText converts a Diff to a text representation for AI prompts.
func diffToText(diff *git_entity.Diff) string {
	var sb strings.Builder
	for _, f := range diff.Files {
		oldLines := strings.Split(strings.TrimRight(f.OldContent, "\n"), "\n")
		newLines := strings.Split(strings.TrimRight(f.NewContent, "\n"), "\n")
		sb.WriteString(fmt.Sprintf("--- a/%s\n+++ b/%s\n", f.Path, f.Path))
		switch f.Status {
		case "A":
			sb.WriteString(fmt.Sprintf("@@ -0,0 +1,%d @@\n", len(newLines)))
			for _, line := range newLines {
				sb.WriteString("+" + line + "\n")
			}
		case "D":
			sb.WriteString(fmt.Sprintf("@@ -1,%d +0,0 @@\n", len(oldLines)))
			for _, line := range oldLines {
				sb.WriteString("-" + line + "\n")
			}
		default:
			sb.WriteString(fmt.Sprintf("@@ -1,%d +1,%d @@\n", len(oldLines), len(newLines)))
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
