package render

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"gitlens/internal/diff"
	"gitlens/internal/diff/theme"
)

func Sidebar(state *diff.AppState, th theme.Theme, width, height int) string {
	if state.SidebarCollapsed {
		return ""
	}
	var b strings.Builder

	for i, f := range state.Files {
		_, isViewed := state.ViewedFiles[f.Path]
		name := filepath.Base(f.Path)

		var statusColor string
		switch f.Status {
		case "A":
			statusColor = th.UI.StatusAdded
		case "D":
			statusColor = th.UI.StatusDeleted
		default:
			statusColor = th.UI.StatusModified
		}

		statusBadge := lipgloss.NewStyle().Foreground(lipgloss.Color(statusColor)).Render(f.Status)
		nameStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(th.UI.Text))
		if isViewed {
			nameStyle = nameStyle.Faint(true)
		}
		if i == state.CurrentFileIdx {
			nameStyle = nameStyle.Background(lipgloss.Color(th.UI.Selection)).Bold(true)
		}
		if i == state.SidebarSelected && state.Focus == diff.FocusSidebar {
			nameStyle = nameStyle.Underline(true)
		}

		dir := filepath.Dir(f.Path)
		indent := ""
		if dir != "." {
			indent = "  "
		}
		line := fmt.Sprintf("%s%s %s", indent, statusBadge, nameStyle.Render(name))
		b.WriteString(line + "\n")
	}
	return lipgloss.NewStyle().
		Width(width).
		Height(height).
		BorderRight(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color(th.UI.Border)).
		Render(b.String())
}
