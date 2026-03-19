package render

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"gitlens/internal/diff"
	hl "gitlens/internal/diff/highlight"
	"gitlens/internal/diff/theme"
	"gitlens/internal/git_entity"
)

func DiffView(state *diff.AppState, th theme.Theme, width, height int) string {
	if len(state.Files) == 0 {
		return lipgloss.NewStyle().Foreground(lipgloss.Color(th.UI.Text)).
			Render("No changes to display")
	}

	currentFile := state.Files[state.CurrentFileIdx]
	highlighter := hl.New(th)
	oldHL := highlighter.HighlightFile(currentFile.Path, currentFile.OldContent)
	newHL := highlighter.HighlightFile(currentFile.Path, currentFile.NewContent)

	gutterWidth := state.Layout.GutterWidth
	if gutterWidth == 0 {
		gutterWidth = 5
	}

	fs := state.Fullscreen
	var panelWidth int
	switch fs {
	case diff.FullscreenOld, diff.FullscreenNew:
		panelWidth = width - gutterWidth // one gutter + one panel
	default:
		panelWidth = (width - gutterWidth*2 - 1) / 2 // two gutters + divider
	}

	var b strings.Builder

	for _, cl := range state.ContextLines {
		ctxStyle := lipgloss.NewStyle().
			Background(lipgloss.Color(th.UI.Border)).
			Foreground(lipgloss.Color(th.UI.Text)).
			Width(width)
		b.WriteString(ctxStyle.Render(fmt.Sprintf("  %d: %s", cl.LineNumber, cl.Content)) + "\n")
	}

	visibleLines := state.DiffLines
	start := state.ScrollY
	if start < 0 {
		start = 0
	}
	end := start + height - len(state.ContextLines)
	if end > len(visibleLines) {
		end = len(visibleLines)
	}
	if end < start {
		end = start
	}

	for i := start; i < end; i++ {
		line := visibleLines[i]
		b.WriteString(renderDiffLine(line, th, panelWidth, gutterWidth, oldHL, newHL, fs) + "\n")
	}

	return strings.TrimRight(b.String(), "\n")
}

func renderDiffLine(line git_entity.DiffLine, th theme.Theme, panelWidth, gutterWidth int,
	oldHL, newHL []hl.HighlightedLine, fs diff.DiffFullscreen) string {

	var oldBg, newBg lipgloss.Color
	switch line.ChangeType {
	case git_entity.Delete:
		oldBg = lipgloss.Color(th.Diff.DeletedBg)
	case git_entity.Insert:
		newBg = lipgloss.Color(th.Diff.AddedBg)
	case git_entity.Modified:
		oldBg = lipgloss.Color(th.Diff.DeletedBg)
		newBg = lipgloss.Color(th.Diff.AddedBg)
	}

	switch fs {
	case diff.FullscreenOld:
		gutter := renderGutter(line.OldLine, line.ChangeType, true, th, gutterWidth)
		panel := renderPanel(line.OldLine, line.OldSegments, oldBg, panelWidth, th)
		return gutter + panel

	case diff.FullscreenNew:
		gutter := renderGutter(line.NewLine, line.ChangeType, false, th, gutterWidth)
		panel := renderPanel(line.NewLine, line.NewSegments, newBg, panelWidth, th)
		return gutter + panel

	default:
		oldGutter := renderGutter(line.OldLine, line.ChangeType, true, th, gutterWidth)
		newGutter := renderGutter(line.NewLine, line.ChangeType, false, th, gutterWidth)
		oldPanel := renderPanel(line.OldLine, line.OldSegments, oldBg, panelWidth, th)
		newPanel := renderPanel(line.NewLine, line.NewSegments, newBg, panelWidth, th)
		divider := lipgloss.NewStyle().Foreground(lipgloss.Color(th.UI.Border)).Render("│")
		return oldGutter + oldPanel + divider + newGutter + newPanel
	}
}

func renderGutter(lc *git_entity.LineContent, ct git_entity.ChangeType, isOld bool, th theme.Theme, width int) string {
	var bg lipgloss.Color
	switch ct {
	case git_entity.Delete:
		if isOld {
			bg = lipgloss.Color(th.Diff.DeletedGutterBg)
		}
	case git_entity.Insert:
		if !isOld {
			bg = lipgloss.Color(th.Diff.AddedGutterBg)
		}
	case git_entity.Modified:
		if isOld {
			bg = lipgloss.Color(th.Diff.DeletedGutterBg)
		} else {
			bg = lipgloss.Color(th.Diff.AddedGutterBg)
		}
	}
	style := lipgloss.NewStyle().Width(width).Background(bg).
		Foreground(lipgloss.Color(th.UI.Text)).Faint(true)
	if lc != nil {
		return style.Render(fmt.Sprintf("%*d", width-1, lc.LineNo))
	}
	return style.Render(strings.Repeat(" ", width))
}

func renderPanel(lc *git_entity.LineContent, segs []git_entity.Segment, bg lipgloss.Color, width int, th theme.Theme) string {
	style := lipgloss.NewStyle().Width(width).Background(bg)
	if lc == nil {
		return style.Render("")
	}
	if len(segs) > 0 {
		var sb strings.Builder
		for _, seg := range segs {
			segStyle := lipgloss.NewStyle().Background(bg)
			if seg.Highlight {
				if bg == lipgloss.Color(th.Diff.DeletedBg) {
					segStyle = segStyle.Background(lipgloss.Color(th.Diff.DeletedWordBg))
				} else {
					segStyle = segStyle.Background(lipgloss.Color(th.Diff.AddedWordBg))
				}
			}
			sb.WriteString(segStyle.Render(seg.Text))
		}
		return style.Render(sb.String())
	}
	return style.Render(lc.Text)
}
