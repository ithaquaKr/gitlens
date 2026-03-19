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

	// Determine the visible window.  SidebarScrollY is maintained by the key
	// handler (clampSidebarScroll), but we defensively recompute it here so
	// that the first render before any key press is also correct.
	scrollY := state.SidebarScrollY
	if state.SidebarSelected < scrollY {
		scrollY = state.SidebarSelected
	}
	if state.SidebarSelected >= scrollY+height {
		scrollY = state.SidebarSelected - height + 1
	}
	if scrollY < 0 {
		scrollY = 0
	}

	// Slice to exactly height rows so lipgloss Height() has nothing to clip.
	end := scrollY + height
	if end > len(items) {
		end = len(items)
	}
	visible := items
	if scrollY > 0 || end < len(items) {
		visible = items[scrollY:end]
	}

	var lines []string
	for j, item := range visible {
		fullIdx := scrollY + j // index into the full items slice
		indent := strings.Repeat("  ", item.Depth)

		var line string
		switch item.Kind {
		case diff.TreeItemRepo:
			repoStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color(th.UI.Border)).
				Background(lipgloss.Color(th.UI.Selection)).
				Bold(true)
			if fullIdx == state.SidebarSelected && state.Focus == diff.FocusSidebar {
				repoStyle = repoStyle.Underline(true)
			}
			line = repoStyle.Render("◉ " + item.Name)

		case diff.TreeItemDir:
			icon := "▾"
			if state.CollapsedDirs[item.DirPath] {
				icon = "▸"
			}
			nameStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(th.UI.Text)).Bold(true)
			if fullIdx == state.SidebarSelected && state.Focus == diff.FocusSidebar {
				nameStyle = nameStyle.Underline(true)
			}
			line = fmt.Sprintf("%s%s %s", indent, icon, nameStyle.Render(item.Name+"/"))

		case diff.TreeItemFile:
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
			if fullIdx == state.SidebarSelected && state.Focus == diff.FocusSidebar {
				nameStyle = nameStyle.Underline(true)
			}

			stats := ""
			if item.FileIdx < len(state.FileStats) {
				st := state.FileStats[item.FileIdx]
				if st.Added > 0 || st.Deleted > 0 {
					added := lipgloss.NewStyle().Foreground(lipgloss.Color(th.UI.StatusAdded)).Faint(true).
						Render(fmt.Sprintf("+%d", st.Added))
					deleted := lipgloss.NewStyle().Foreground(lipgloss.Color(th.UI.StatusDeleted)).Faint(true).
						Render(fmt.Sprintf("-%d", st.Deleted))
					stats = " " + added + " " + deleted
				}
			}

			line = fmt.Sprintf("%s%s %s%s", indent, statusBadge, nameStyle.Render(item.Name), stats)
		}
		lines = append(lines, line)
	}

	content := strings.Join(lines, "\n")
	return lipgloss.NewStyle().
		Width(width).
		Height(height).
		BorderRight(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color(th.UI.Border)).
		Render(content)
}
