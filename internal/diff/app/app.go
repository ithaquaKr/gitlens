package app

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"gitlens/internal/diff"
	"gitlens/internal/diff/render"
	"gitlens/internal/diff/theme"
)

type modalKind int

const (
	modalNone modalKind = iota
	modalHelp
	modalAnnotations
	modalFilePicker
	modalAnnotationEditor
)

// Model is the Bubble Tea model for the diff TUI.
type Model struct {
	state       *diff.AppState
	theme       theme.Theme
	modal       modalKind
	branchName  string
	editInput   string
	cancelWatch func()
}

func NewModel(state *diff.AppState, th theme.Theme, branchName string) Model {
	return Model{
		state:      state,
		theme:      th,
		branchName: branchName,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.EnableMouseCellMotion
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.state.Width = msg.Width
		m.state.Height = msg.Height
		return m, nil
	case diff.ReloadMsg:
		return m, nil
	case tea.MouseMsg:
		return m.handleMouse(msg)
	case tea.KeyMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	s := m.state
	if m.modal != modalNone {
		return m.handleModalKey(msg)
	}
	switch msg.String() {
	case "q", "ctrl+c":
		if m.cancelWatch != nil {
			m.cancelWatch()
		}
		return m, tea.Quit
	case "?":
		m.modal = modalHelp
	case "I":
		m.modal = modalAnnotations
	case "ctrl+p":
		m.modal = modalFilePicker
	case "j", "down":
		s.ScrollY = clamp(s.ScrollY+1, 0, len(s.DiffLines)-1)
	case "k", "up":
		s.ScrollY = clamp(s.ScrollY-1, 0, len(s.DiffLines)-1)
	case "ctrl+d":
		s.ScrollY = clamp(s.ScrollY+s.Height/2, 0, len(s.DiffLines)-1)
	case "ctrl+u":
		s.ScrollY = clamp(s.ScrollY-s.Height/2, 0, len(s.DiffLines)-1)
	case "pgdown":
		s.ScrollY = clamp(s.ScrollY+s.Height, 0, len(s.DiffLines)-1)
	case "pgup":
		s.ScrollY = clamp(s.ScrollY-s.Height, 0, len(s.DiffLines)-1)
	case "G":
		s.ScrollY = len(s.DiffLines) - 1
	case "g":
		if s.PendingKey == diff.PendingKeyG {
			s.ScrollY = 0
			s.PendingKey = diff.PendingKeyNone
		} else {
			s.PendingKey = diff.PendingKeyG
		}
		return m, nil
	case "h":
		if s.Focus == diff.FocusDiff {
			s.ScrollX = clamp(s.ScrollX-4, 0, 1000)
		} else {
			s.SidebarSelected = clamp(s.SidebarSelected-1, 0, len(s.Files)-1)
		}
	case "l":
		if s.Focus == diff.FocusDiff {
			s.ScrollX += 4
		} else {
			s.SidebarSelected = clamp(s.SidebarSelected+1, 0, len(s.Files)-1)
		}
	case "{":
		jumpToHunk(s, -1)
	case "}":
		jumpToHunk(s, 1)
	case "ctrl+j":
		s.NavigateToFile(s.CurrentFileIdx + 1)
	case "ctrl+k":
		s.NavigateToFile(s.CurrentFileIdx - 1)
	case "[":
		s.Fullscreen = diff.FullscreenOld
	case "]":
		s.Fullscreen = diff.FullscreenNew
	case "=":
		s.Fullscreen = diff.FullscreenOff
	case "1":
		s.Focus = diff.FocusSidebar
	case "2":
		s.Focus = diff.FocusDiff
	case "tab":
		s.SidebarCollapsed = !s.SidebarCollapsed
	case "e":
		return m, openEditor(s)
	case "y":
		return m, copySelection(s)
	case "i":
		m.modal = modalAnnotationEditor
		m.editInput = ""
	case "space":
		return m.handleSpace()
	case "/":
		s.SearchActive = true
		s.SearchQuery = ""
	case "n":
		advanceSearch(s, 1)
	case "N":
		advanceSearch(s, -1)
	case "ctrl+l":
		if s.StackedMode {
			navigateStack(s, 1)
		}
	case "ctrl+h":
		if s.StackedMode {
			navigateStack(s, -1)
		}
	case "enter":
		if s.Focus == diff.FocusSidebar {
			s.NavigateToFile(s.SidebarSelected)
			s.Focus = diff.FocusDiff
		}
	}
	s.PendingKey = diff.PendingKeyNone
	return m, nil
}

func (m Model) handleModalKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "q", "?":
		m.modal = modalNone
	case "enter":
		if m.modal == modalAnnotationEditor && strings.TrimSpace(m.editInput) != "" {
			addAnnotation(m.state, m.editInput)
			m.modal = modalNone
			m.editInput = ""
		}
	default:
		if m.modal == modalAnnotationEditor {
			if msg.String() == "backspace" && len(m.editInput) > 0 {
				m.editInput = m.editInput[:len(m.editInput)-1]
			} else if len(msg.Runes) > 0 {
				m.editInput += string(msg.Runes)
			}
		}
	}
	return m, nil
}

func (m Model) handleMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	s := m.state
	switch msg.Button {
	case tea.MouseButtonWheelUp:
		s.ScrollY = clamp(s.ScrollY-3, 0, len(s.DiffLines)-1)
	case tea.MouseButtonWheelDown:
		s.ScrollY = clamp(s.ScrollY+3, 0, len(s.DiffLines)-1)
	case tea.MouseButtonLeft:
		switch msg.Action {
		case tea.MouseActionPress:
			s.Anchor = &diff.CursorPos{Row: msg.Y, Col: msg.X}
			s.Head = &diff.CursorPos{Row: msg.Y, Col: msg.X}
			s.ShowSelectionTooltip = false
			if msg.X < s.Layout.GutterWidth || (msg.X >= s.Layout.NewPanelStart && msg.X < s.Layout.NewPanelStart+s.Layout.GutterWidth) {
				s.SelectionMode = diff.SelectionLine
			} else {
				s.SelectionMode = diff.SelectionChar
			}
		case tea.MouseActionMotion:
			if s.Anchor != nil {
				s.Head = &diff.CursorPos{Row: msg.Y, Col: msg.X}
			}
		case tea.MouseActionRelease:
			if s.Anchor != nil && s.Head != nil {
				s.ShowSelectionTooltip = true
			}
		}
	}
	return m, nil
}

func (m Model) handleSpace() (tea.Model, tea.Cmd) {
	s := m.state
	if s.Focus == diff.FocusSidebar {
		if s.SidebarSelected < len(s.Files) {
			path := s.Files[s.SidebarSelected].Path
			if _, ok := s.ViewedFiles[path]; ok {
				delete(s.ViewedFiles, path)
			} else {
				s.ViewedFiles[path] = struct{}{}
			}
		}
	} else {
		s.ToggleViewed()
		s.NextUnviewed()
	}
	return m, nil
}

func (m Model) View() string {
	s := m.state
	if s.Width == 0 {
		return "Loading..."
	}

	sidebarWidth := 0
	sidebarView := ""
	if !s.SidebarCollapsed {
		sidebarWidth = clamp(s.Width/4, 20, 40)
		sidebarView = render.Sidebar(s, m.theme, sidebarWidth)
	}

	diffWidth := s.Width - sidebarWidth
	s.Layout = diff.PanelLayout{
		SidebarWidth:  sidebarWidth,
		GutterWidth:   5,
		OldPanelStart: sidebarWidth + 5,
		OldPanelEnd:   sidebarWidth + diffWidth/2,
		NewPanelStart: sidebarWidth + diffWidth/2 + 6,
		NewPanelEnd:   s.Width,
	}

	diffView := render.DiffView(s, m.theme, diffWidth, s.Height-2)
	footer := render.Footer(s, m.theme, s.Width, m.branchName)

	fileHeader := ""
	if s.CurrentFileIdx < len(s.Files) {
		f := s.Files[s.CurrentFileIdx]
		fileHeader = lipgloss.NewStyle().
			Width(diffWidth).
			Background(lipgloss.Color(m.theme.UI.Border)).
			Foreground(lipgloss.Color(m.theme.UI.Text)).
			Bold(true).
			Render(fmt.Sprintf(" %s [%s]", f.Path, f.Status))
	}

	mainContent := lipgloss.JoinVertical(lipgloss.Left, fileHeader, diffView)
	view := lipgloss.JoinHorizontal(lipgloss.Top, sidebarView, mainContent)
	view = lipgloss.JoinVertical(lipgloss.Left, view, footer)

	switch m.modal {
	case modalHelp:
		modal := render.HelpModal(m.theme, s.Width, s.Height)
		view = overlayCenter(view, modal, s.Width, s.Height)
	case modalAnnotations:
		modal := render.AnnotationsModal(s, m.theme)
		view = overlayCenter(view, modal, s.Width, s.Height)
	case modalAnnotationEditor:
		editor := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(m.theme.UI.Border)).
			Padding(1, 2).
			Width(60).
			Render("Add annotation:\n\n" + m.editInput + "█\n\nenter to save  esc to cancel")
		view = overlayCenter(view, editor, s.Width, s.Height)
	}

	return view
}

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func jumpToHunk(s *diff.AppState, dir int) {
	idx := s.CurrentHunkIdx()
	next := idx + dir
	if next >= 0 && next < len(s.Hunks) {
		s.ScrollY = s.Hunks[next].StartIdx
	}
}

func navigateStack(s *diff.AppState, dir int) {
	next := s.CurrentCommitIdx + dir
	if next >= 0 && next < len(s.StackedCommits) {
		s.CurrentCommitIdx = next
	}
}

func advanceSearch(s *diff.AppState, dir int) {
	if len(s.SearchMatches) == 0 {
		return
	}
	s.SearchIdx = (s.SearchIdx + dir + len(s.SearchMatches)) % len(s.SearchMatches)
	match := s.SearchMatches[s.SearchIdx]
	s.ScrollY = match.LineIdx
}

func addAnnotation(s *diff.AppState, content string) {
	if s.CurrentFileIdx >= len(s.Files) {
		return
	}
	ann := diff.Annotation{
		ID:        fmt.Sprintf("%d", len(s.Annotations)+1),
		Filename:  s.Files[s.CurrentFileIdx].Path,
		Content:   content,
		CreatedAt: timeNow(),
	}
	if s.Anchor != nil && s.Head != nil {
		ann.Target = diff.AnnotationTarget{
			Kind:      diff.TargetLineRange,
			Panel:     s.SelectionPanel,
			StartLine: s.Anchor.Row + s.ScrollY,
			EndLine:   s.Head.Row + s.ScrollY,
		}
	} else {
		ann.Target = diff.AnnotationTarget{Kind: diff.TargetFile}
	}
	s.Annotations = append(s.Annotations, ann)
}

func openEditor(s *diff.AppState) tea.Cmd {
	if s.CurrentFileIdx >= len(s.Files) {
		return nil
	}
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vi"
	}
	path := s.Files[s.CurrentFileIdx].Path
	return tea.ExecProcess(exec.Command(editor, path), nil)
}

func copySelection(s *diff.AppState) tea.Cmd {
	return func() tea.Msg {
		if s.Anchor == nil || s.Head == nil {
			return nil
		}
		var sb strings.Builder
		start := clamp(s.Anchor.Row+s.ScrollY, 0, len(s.DiffLines)-1)
		end := clamp(s.Head.Row+s.ScrollY, 0, len(s.DiffLines)-1)
		if start > end {
			start, end = end, start
		}
		for i := start; i <= end; i++ {
			dl := s.DiffLines[i]
			if s.SelectionPanel == diff.PanelOld && dl.OldLine != nil {
				sb.WriteString(dl.OldLine.Text + "\n")
			} else if s.SelectionPanel == diff.PanelNew && dl.NewLine != nil {
				sb.WriteString(dl.NewLine.Text + "\n")
			}
		}
		_ = clipboard.WriteAll(sb.String())
		return nil
	}
}

func overlayCenter(base, overlay string, w, h int) string {
	overlayLines := strings.Split(overlay, "\n")
	baseLines := strings.Split(base, "\n")
	overlayH := len(overlayLines)
	overlayW := lipgloss.Width(overlay)
	startY := (h - overlayH) / 2
	startX := (w - overlayW) / 2
	if startY < 0 {
		startY = 0
	}
	if startX < 0 {
		startX = 0
	}
	for i, line := range overlayLines {
		lineIdx := startY + i
		if lineIdx >= len(baseLines) {
			break
		}
		baseLine := baseLines[lineIdx]
		baseW := lipgloss.Width(baseLine)
		if startX+overlayW > baseW {
			baseLines[lineIdx] = baseLine + strings.Repeat(" ", startX+overlayW-baseW)
		}
		r := []rune(baseLines[lineIdx])
		ol := []rune(line)
		for j, ch := range ol {
			if startX+j < len(r) {
				r[startX+j] = ch
			}
		}
		baseLines[lineIdx] = string(r)
	}
	return strings.Join(baseLines, "\n")
}

var (
	timeNow        = func() time.Time { return time.Now() }
	clipboardWrite = func(s string) error { return clipboard.WriteAll(s) }
)
