package render

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"gitlens/internal/diff"
	"gitlens/internal/diff/theme"
	"gitlens/internal/git_entity"
)

func Footer(state *diff.AppState, th theme.Theme, width int, branchName string) string {
	textStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(th.UI.Text))
	borderStyle := lipgloss.NewStyle().
		Background(lipgloss.Color(th.UI.Border)).
		Foreground(lipgloss.Color(th.UI.Text))

	branch := borderStyle.Padding(0, 1).Render(" " + branchName + " ")

	added, deleted := countChanges(state)
	stats := textStyle.Render(fmt.Sprintf("+%d -%d", added, deleted))

	stacked := ""
	if state.StackedMode && len(state.StackedCommits) > 0 {
		current := state.StackedCommits[state.CurrentCommitIdx]
		shortSHA := current.Hash
		if len(shortSHA) > 7 {
			shortSHA = shortSHA[:7]
		}
		stacked = borderStyle.Padding(0, 1).Render(
			fmt.Sprintf("[%d/%d] %s", state.CurrentCommitIdx+1, len(state.StackedCommits), shortSHA),
		)
	}

	hints := textStyle.Faint(true).Render("q quit  ? help  tab sidebar  / search")

	parts := []string{branch, stats}
	if stacked != "" {
		parts = append(parts, stacked)
	}
	usedWidth := lipgloss.Width(strings.Join(parts, "  ")) + lipgloss.Width(hints) + 4
	padding := max(0, width-usedWidth)
	parts = append(parts, strings.Repeat(" ", padding))
	parts = append(parts, hints)

	return lipgloss.NewStyle().Width(width).Render(strings.Join(parts, "  "))
}

func countChanges(state *diff.AppState) (added, deleted int) {
	for _, l := range state.DiffLines {
		switch l.ChangeType {
		case git_entity.Insert, git_entity.Modified:
			added++
		case git_entity.Delete:
			deleted++
		}
	}
	return
}
