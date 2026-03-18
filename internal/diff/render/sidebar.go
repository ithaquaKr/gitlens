package render

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"gitlens/internal/diff"
	"gitlens/internal/diff/theme"
)

func Sidebar(state *diff.AppState, th theme.Theme, width, height int) string {
	if state.SidebarCollapsed {
		return ""
	}

	items := diff.VisibleTreeItems(state.Files, state.CollapsedDirs)

	var b strings.Builder
	for i, item := range items {
		indent := strings.Repeat("  ", item.Depth)

		var line string
		if item.Kind == diff.TreeItemDir {
			icon := "▾"
			if state.CollapsedDirs[item.DirPath] {
				icon = "▸"
			}
			nameStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(th.UI.Text)).Bold(true)
			if i == state.SidebarSelected && state.Focus == diff.FocusSidebar {
				nameStyle = nameStyle.Underline(true)
			}
			line = fmt.Sprintf("%s%s %s", indent, icon, nameStyle.Render(item.Name+"/"))
		} else {
			f := state.Files[item.FileIdx]
			_, isViewed := state.ViewedFiles[f.Path]

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
			if item.FileIdx == state.CurrentFileIdx {
				nameStyle = nameStyle.Background(lipgloss.Color(th.UI.Selection)).Bold(true)
			}
			if i == state.SidebarSelected && state.Focus == diff.FocusSidebar {
				nameStyle = nameStyle.Underline(true)
			}
			line = fmt.Sprintf("%s%s %s", indent, statusBadge, nameStyle.Render(item.Name))
		}
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
