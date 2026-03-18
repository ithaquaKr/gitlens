//go:build !cgo

package diff

// ContextLine is one sticky header line shown at the top of the diff view.
type ContextLine struct {
	LineNumber int
	Content    string
}

// ComputeContextLines returns nil when CGo is unavailable (no tree-sitter).
func ComputeContextLines(path, content string, currentLine int) []ContextLine {
	return nil
}
