//go:build !cgo

package highlight

import (
	"strings"

	"gitlens/internal/diff/theme"
)

// Token is one highlighted span within a line.
type Token struct {
	Text  string
	Color string // hex color or "" for default
}

// HighlightedLine is a single source line broken into colored tokens.
type HighlightedLine struct {
	Tokens []Token
}

// Highlighter holds theme state for syntax coloring.
type Highlighter struct {
	theme theme.Theme
}

func New(th theme.Theme) *Highlighter {
	return &Highlighter{theme: th}
}

// HighlightFile returns plain-text lines when CGo is not available.
func (h *Highlighter) HighlightFile(path, content string) []HighlightedLine {
	return plainLines(content)
}

func plainLines(content string) []HighlightedLine {
	lines := strings.Split(content, "\n")
	result := make([]HighlightedLine, len(lines))
	for i, l := range lines {
		result[i] = HighlightedLine{Tokens: []Token{{Text: l}}}
	}
	return result
}
