package diff

import (
	"time"

	"gitlens/internal/git_entity"
)

type PendingKey int

const (
	PendingKeyNone PendingKey = iota
	PendingKeyG
)

type DiffFullscreen int

const (
	FullscreenOff DiffFullscreen = iota
	FullscreenOld
	FullscreenNew
)

type SelectionMode int

const (
	SelectionChar SelectionMode = iota
	SelectionLine
)

type PanelFocus int

const (
	PanelOld PanelFocus = iota
	PanelNew
)

type FocusArea int

const (
	FocusDiff FocusArea = iota
	FocusSidebar
)

type CursorPos struct {
	Row, Col int
}

type MatchPos struct {
	FileIdx int
	LineIdx int
	Col     int
}

type TargetKind int

const (
	TargetFile TargetKind = iota
	TargetLineRange
)

type AnnotationTarget struct {
	Kind      TargetKind
	Panel     PanelFocus
	StartLine int
	EndLine   int
}

type Annotation struct {
	ID        string
	Filename  string
	Target    AnnotationTarget
	Content   string
	CreatedAt time.Time
}

type PanelLayout struct {
	SidebarWidth  int
	OldPanelStart int
	OldPanelEnd   int
	NewPanelStart int
	NewPanelEnd   int
	GutterWidth   int
}

// AppState holds all mutable state for the diff TUI.
type AppState struct {
	Files          []git_entity.FileDiff
	CurrentFileIdx int

	DiffLines []git_entity.DiffLine
	Hunks     []git_entity.Hunk

	ScrollY        int
	ScrollX        int
	SidebarScrollX int

	ContextLines []ContextLine
	PendingKey   PendingKey
	Fullscreen   DiffFullscreen
	Focus        FocusArea

	SidebarCollapsed bool
	SidebarSelected  int
	CollapsedDirs    map[string]bool

	Anchor               *CursorPos
	Head                 *CursorPos
	SelectionMode        SelectionMode
	SelectionPanel       PanelFocus
	ShowSelectionTooltip bool

	Annotations []Annotation

	SearchActive  bool
	SearchQuery   string
	SearchMatches []MatchPos
	SearchIdx     int

	ViewedFiles map[string]struct{}

	StackedMode        bool
	StackedCommits     []*git_entity.Commit
	CurrentCommitIdx   int
	StackedViewedFiles map[string]map[string]struct{}

	Layout PanelLayout

	Width, Height int
}

// NewAppState initializes AppState with empty collections.
func NewAppState(files []git_entity.FileDiff) *AppState {
	s := &AppState{
		Files:              files,
		CollapsedDirs:      make(map[string]bool),
		ViewedFiles:        make(map[string]struct{}),
		StackedViewedFiles: make(map[string]map[string]struct{}),
	}
	if len(files) > 0 {
		s.recompute()
	}
	return s
}

func (s *AppState) recompute() {
	if s.CurrentFileIdx >= len(s.Files) {
		return
	}
	f := s.Files[s.CurrentFileIdx]
	s.DiffLines = ComputeDiffLines(f.OldContent, f.NewContent)
	s.Hunks = ComputeHunks(s.DiffLines)
	s.ScrollY = 0
	s.ScrollX = 0
}

func (s *AppState) NavigateToFile(idx int) {
	if idx < 0 || idx >= len(s.Files) {
		return
	}
	s.CurrentFileIdx = idx
	s.recompute()
}

func (s *AppState) ToggleViewed() {
	if s.CurrentFileIdx >= len(s.Files) {
		return
	}
	path := s.Files[s.CurrentFileIdx].Path
	if _, ok := s.ViewedFiles[path]; ok {
		delete(s.ViewedFiles, path)
	} else {
		s.ViewedFiles[path] = struct{}{}
	}
}

func (s *AppState) NextUnviewed() {
	start := s.CurrentFileIdx
	for i := 1; i <= len(s.Files); i++ {
		idx := (start + i) % len(s.Files)
		path := s.Files[idx].Path
		if _, viewed := s.ViewedFiles[path]; !viewed {
			s.NavigateToFile(idx)
			return
		}
	}
}

func (s *AppState) CurrentHunkIdx() int {
	for i, h := range s.Hunks {
		if s.ScrollY >= h.StartIdx && s.ScrollY <= h.EndIdx {
			return i
		}
	}
	return -1
}
