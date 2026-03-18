//go:build cgo

package highlight

import (
	"path/filepath"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/csharp"
	"github.com/smacker/go-tree-sitter/golang"
	"github.com/smacker/go-tree-sitter/javascript"
	"github.com/smacker/go-tree-sitter/python"
	"github.com/smacker/go-tree-sitter/rust"
	"github.com/smacker/go-tree-sitter/typescript/tsx"
	"github.com/smacker/go-tree-sitter/typescript/typescript"
)

// LanguageForFile returns the tree-sitter language for the given file path.
// Returns nil for unsupported files.
func LanguageForFile(path string) *sitter.Language {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".go":
		return golang.GetLanguage()
	case ".rs":
		return rust.GetLanguage()
	case ".ts":
		return typescript.GetLanguage()
	case ".tsx":
		return tsx.GetLanguage()
	case ".js", ".mjs", ".cjs":
		return javascript.GetLanguage()
	case ".py":
		return python.GetLanguage()
	case ".cs":
		return csharp.GetLanguage()
	default:
		return nil
	}
}
