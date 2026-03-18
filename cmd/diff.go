package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	diff "gitlens/internal/diff"
	diffapp "gitlens/internal/diff/app"
	"gitlens/internal/diff/theme"
	"gitlens/internal/git_entity"
	"gitlens/internal/vcs"
)

var (
	diffWatch   bool
	diffTheme   string
	diffStacked bool
	diffFocus   string
	diffFiles   []string
)

var diffCmd = &cobra.Command{
	Use:   "diff [ref]",
	Short: "Interactive side-by-side diff viewer",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		backend, err := vcs.NewGitBackend(".")
		if err != nil {
			return err
		}

		// Load diff
		var gitDiff *git_entity.Diff
		if len(args) > 0 {
			ref := vcs.ParseRef(args[0])
			if ref.Single != "" {
				gitDiff, err = backend.GetRangeDiff(ref.Single+"^", ref.Single, false)
				if err != nil {
					gitDiff, err = backend.GetRangeDiff("", ref.Single, false)
				}
			} else {
				gitDiff, err = backend.GetRangeDiff(ref.From, ref.To, ref.ThreeDot)
			}
		} else {
			gitDiff, err = backend.GetWorkingTreeDiff(false)
		}
		if err != nil {
			return err
		}
		if len(gitDiff.Files) == 0 {
			return fmt.Errorf("no changes to display")
		}

		// Filter files if --file flags provided
		files := gitDiff.Files
		if len(diffFiles) > 0 {
			files = filterFiles(files, diffFiles)
		}

		// Load theme
		themeName := Cfg.Theme.Base
		if diffTheme != "" {
			themeName = diffTheme
			Cfg.Theme.Base = themeName
		}
		_ = themeName
		th := theme.Load(Cfg)

		// Get branch name for footer
		branchName := getBranchName(backend)

		// Build app state
		state := diff.NewAppState(files)

		// Focus on specific file if --focus provided
		if diffFocus != "" {
			for i, f := range files {
				if f.Path == diffFocus || filepath.Base(f.Path) == diffFocus {
					state.NavigateToFile(i)
					break
				}
			}
		}

		// Stacked mode
		if diffStacked && len(args) > 0 {
			ref := vcs.ParseRef(args[0])
			if ref.From != "" {
				commits, err := backend.GetCommitsInRange(ref.From, ref.To)
				if err == nil && len(commits) > 0 {
					state.StackedMode = true
					state.StackedCommits = commits
				}
			}
		}

		model := diffapp.NewModel(state, th, branchName)

		p := tea.NewProgram(model,
			tea.WithAltScreen(),
			tea.WithMouseCellMotion(),
		)

		// Watch mode
		if diffWatch {
			watchPaths := watchPathsFor(args, backend)
			cancel, err := diff.WatchFiles(p, watchPaths)
			if err != nil {
				fmt.Fprintf(os.Stderr, "watch: %v\n", err)
			} else {
				defer cancel()
			}
		}

		_, err = p.Run()
		return err
	},
}

func init() {
	rootCmd.AddCommand(diffCmd)
	diffCmd.Flags().BoolVar(&diffWatch, "watch", false, "auto-reload on file changes")
	diffCmd.Flags().StringVar(&diffTheme, "theme", "", "color theme override")
	diffCmd.Flags().BoolVar(&diffStacked, "stacked", false, "commit-by-commit navigation")
	diffCmd.Flags().StringVar(&diffFocus, "focus", "", "start at this file")
	diffCmd.Flags().StringArrayVar(&diffFiles, "file", nil, "filter to specific files")
}

func filterFiles(files []git_entity.FileDiff, filter []string) []git_entity.FileDiff {
	filterSet := make(map[string]bool)
	for _, f := range filter {
		filterSet[f] = true
	}
	var result []git_entity.FileDiff
	for _, f := range files {
		if filterSet[f.Path] || filterSet[filepath.Base(f.Path)] {
			result = append(result, f)
		}
	}
	return result
}

func getBranchName(backend *vcs.GitBackend) string {
	repo := backend.Repo()
	head, err := repo.Head()
	if err != nil {
		return "HEAD"
	}
	if head.Name().IsBranch() {
		return head.Name().Short()
	}
	return head.Hash().String()[:7]
}

func watchPathsFor(args []string, backend vcs.Backend) []string {
	if len(args) == 0 {
		return []string{".git/index"}
	}
	return []string{".git/refs", ".git/HEAD"}
}
