package git_entity

import "time"

// --- Git layer models (returned by VCS backend) ---

type Commit struct {
	Hash    string
	Message string
	Author  string
	Email   string
	Date    time.Time
}

type Diff struct {
	Files []FileDiff
}

// FileDiff holds both sides of a file change for the TUI renderer.
type FileDiff struct {
	Path       string // current path
	Status     string // "A", "M", "D"
	OldContent string
	NewContent string
	IsBinary   bool
}

// --- Rendering layer models (output of diff_algo) ---

type ChangeType int

const (
	Equal    ChangeType = iota
	Delete
	Insert
	Modified
)

type Segment struct {
	Text      string
	Highlight bool // word-level diff highlight
}

type LineContent struct {
	LineNo int
	Text   string
}

type DiffLine struct {
	OldLine     *LineContent
	NewLine     *LineContent
	ChangeType  ChangeType
	OldSegments []Segment
	NewSegments []Segment
}

// Hunk is a contiguous block of non-Equal lines used for {/} navigation.
type Hunk struct {
	StartIdx int
	EndIdx   int
}
