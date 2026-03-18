//go:build cgo

package diff

import (
	"context"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	"gitlens/internal/diff/highlight"
)

// ContextLine is one sticky header line shown at the top of the diff view.
type ContextLine struct {
	LineNumber int
	Content    string
}

// ComputeContextLines returns the enclosing scope headers (up to 5)
// for the current scroll position in the file.
func ComputeContextLines(path, content string, currentLine int) []ContextLine {
	lang := highlight.LanguageForFile(path)
	if lang == nil {
		return nil
	}
	parser := sitter.NewParser()
	parser.SetLanguage(lang)
	tree, err := parser.ParseCtx(context.Background(), nil, []byte(content))
	if err != nil || tree == nil {
		return nil
	}

	lines := strings.Split(content, "\n")
	if currentLine >= len(lines) {
		return nil
	}

	offset := 0
	for i := 0; i < currentLine && i < len(lines); i++ {
		offset += len(lines[i]) + 1
	}

	root := tree.RootNode()
	scopeTypes := map[string]bool{
		"function_declaration": true, "method_declaration": true,
		"func_literal": true, "function_definition": true,
		"class_declaration": true, "class_definition": true,
		"impl_item": true, "trait_item": true,
		"struct_item": true, "enum_item": true,
	}

	var enclosing []ContextLine
	var walk func(node *sitter.Node)
	walk = func(node *sitter.Node) {
		if node.StartByte() <= uint32(offset) && node.EndByte() > uint32(offset) {
			if scopeTypes[node.Type()] {
				startByte := node.StartByte()
				lineNo := strings.Count(content[:startByte], "\n")
				if lineNo < currentLine {
					enclosing = append(enclosing, ContextLine{
						LineNumber: lineNo + 1,
						Content:    lines[lineNo],
					})
				}
			}
			for i := 0; i < int(node.ChildCount()); i++ {
				walk(node.Child(i))
			}
		}
	}
	walk(root)

	if len(enclosing) > 5 {
		enclosing = enclosing[len(enclosing)-5:]
	}
	return enclosing
}
