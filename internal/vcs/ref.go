package vcs

import "strings"

// CommitReference represents a parsed git ref argument.
type CommitReference struct {
	Single   string // HEAD, sha
	From, To string // range or three-dot
	ThreeDot bool
}

// ParseRef parses strings like "HEAD", "abc123", "main..feature", "main...feature".
func ParseRef(s string) CommitReference {
	if strings.Contains(s, "...") {
		parts := strings.SplitN(s, "...", 2)
		return CommitReference{From: parts[0], To: parts[1], ThreeDot: true}
	}
	if strings.Contains(s, "..") {
		parts := strings.SplitN(s, "..", 2)
		return CommitReference{From: parts[0], To: parts[1]}
	}
	return CommitReference{Single: s}
}
