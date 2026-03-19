package render

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"gitlens/internal/diff"
	"gitlens/internal/diff/theme"
)

func HelpModal(th theme.Theme, width, height int) string {
	help := [][]string{
		{"j/k, ↓/↑", "scroll diff / move in sidebar"},
		{"ctrl+d/u", "half-page scroll"},
		{"pgdn/pgup", "full-page scroll"},
		{"g g / G", "top / bottom"},
		{"h/l (diff)", "scroll left / right"},
		{"h (sidebar)", "collapse dir"},
		{"l / enter (sidebar)", "expand / open"},
		{"{/}", "prev/next hunk"},
		{"J/K", "next/prev file (macOS+Linux)"},
		{"ctrl+j/k", "next/prev file (Linux only)"},
		{"ctrl+p", "file picker"},
		{"[/]/=", "fullscreen old/new/reset"},
		{"tab", "toggle sidebar"},
		{"1/2", "focus sidebar/diff"},
		{"e", "open in $EDITOR"},
		{"y", "copy selection"},
		{"i", "annotate"},
		{"I", "view annotations"},
		{"space", "mark file viewed"},
		{"/", "search"},
		{"n/N", "next/prev match"},
		{"</> (stacked)", "prev/next commit"},
		{"?", "toggle help"},
		{"q / ctrl+c", "quit"},
	}
	boxWidth := 50
	var b strings.Builder
	b.WriteString(lipgloss.NewStyle().Bold(true).Render("Keybindings") + "\n\n")
	for _, row := range help {
		key := lipgloss.NewStyle().Foreground(lipgloss.Color(th.Syntax.Keyword)).
			Width(14).Render(row[0])
		desc := lipgloss.NewStyle().Foreground(lipgloss.Color(th.UI.Text)).Render(row[1])
		b.WriteString(key + "  " + desc + "\n")
	}
	return lipgloss.NewStyle().
		Width(boxWidth).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(th.UI.Border)).
		Padding(1, 2).
		Render(b.String())
}

func AnnotationsModal(state *diff.AppState, th theme.Theme) string {
	if len(state.Annotations) == 0 {
		return lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(th.UI.Border)).
			Padding(1, 2).
			Render("No annotations yet.\n\nPress i to annotate a selection.")
	}

	var b strings.Builder
	b.WriteString(lipgloss.NewStyle().Bold(true).Render("Annotations") + "\n\n")
	for i, ann := range state.Annotations {
		loc := ann.Filename
		if ann.Target.Kind == diff.TargetLineRange {
			panel := "left"
			if ann.Target.Panel == diff.PanelNew {
				panel = "right"
			}
			loc = fmt.Sprintf("%s:%s L%d-%d", ann.Filename, panel, ann.Target.StartLine, ann.Target.EndLine)
		}
		header := lipgloss.NewStyle().
			Foreground(lipgloss.Color(th.Syntax.Function)).
			Render(fmt.Sprintf("[%d] %s", i+1, loc))
		preview := ann.Content
		if len(preview) > 60 {
			preview = preview[:60] + "..."
		}
		b.WriteString(header + "\n")
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(th.UI.Text)).Faint(true).
			Render("  "+preview) + "\n\n")
	}
	b.WriteString(lipgloss.NewStyle().Faint(true).Render("d delete  y copy  enter jump  esc close"))
	return lipgloss.NewStyle().
		Width(70).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(th.UI.Border)).
		Padding(1, 2).
		Render(b.String())
}
