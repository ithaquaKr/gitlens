//go:build cgo

package highlight

import (
	"context"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
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

// HighlightFile returns highlighted lines for the given source content and file path.
// Falls back to plain text if the language is unsupported.
func (h *Highlighter) HighlightFile(path, content string) []HighlightedLine {
	lang := LanguageForFile(path)
	if lang == nil {
		return plainLines(content)
	}
	return h.highlightWithTreeSitter(lang, content)
}

func (h *Highlighter) highlightWithTreeSitter(lang *sitter.Language, content string) []HighlightedLine {
	parser := sitter.NewParser()
	parser.SetLanguage(lang)
	tree, err := parser.ParseCtx(context.Background(), nil, []byte(content))
	if err != nil || tree == nil {
		return plainLines(content)
	}

	sourceBytes := []byte(content)
	_ = sourceBytes
	root := tree.RootNode()
	var tokens []struct {
		start, end uint32
		nodeType   string
	}
	var walk func(node *sitter.Node)
	walk = func(node *sitter.Node) {
		if node.ChildCount() == 0 && node.EndByte() > node.StartByte() {
			tokens = append(tokens, struct {
				start, end uint32
				nodeType   string
			}{node.StartByte(), node.EndByte(), node.Type()})
		}
		for i := 0; i < int(node.ChildCount()); i++ {
			walk(node.Child(i))
		}
	}
	walk(root)

	colorFor := func(nodeType string) string {
		switch nodeType {
		case "identifier":
			return h.theme.Syntax.Variable
		case "string_literal", "string", "interpreted_string_literal", "raw_string_literal":
			return h.theme.Syntax.String
		case "comment", "line_comment", "block_comment":
			return h.theme.Syntax.Comment
		case "int_literal", "float_literal", "imaginary_literal", "integer_literal":
			return h.theme.Syntax.Number
		case "func_literal", "function_declaration", "method_declaration":
			return h.theme.Syntax.Function
		case "type_identifier", "type_spec":
			return h.theme.Syntax.Type
		default:
			if isKeyword(nodeType) {
				return h.theme.Syntax.Keyword
			}
			return ""
		}
	}

	rawLines := strings.Split(content, "\n")
	result := make([]HighlightedLine, len(rawLines))
	tokenIdx := 0
	offset := uint32(0)
	for lineIdx, line := range rawLines {
		lineEnd := offset + uint32(len(line))
		var lineTokens []Token
		for tokenIdx < len(tokens) && tokens[tokenIdx].start < lineEnd {
			tok := tokens[tokenIdx]
			start := tok.start
			end := tok.end
			if end > lineEnd {
				end = lineEnd
			}
			relStart := start - offset
			relEnd := end - offset
			if relEnd > uint32(len(line)) {
				relEnd = uint32(len(line))
			}
			text := line[relStart:relEnd]
			lineTokens = append(lineTokens, Token{Text: text, Color: colorFor(tok.nodeType)})
			if tok.end >= lineEnd {
				break
			}
			tokenIdx++
		}
		if len(lineTokens) == 0 {
			lineTokens = []Token{{Text: line}}
		}
		result[lineIdx] = HighlightedLine{Tokens: lineTokens}
		offset = lineEnd + 1
	}
	return result
}

func isKeyword(nodeType string) bool {
	keywords := map[string]bool{
		"if": true, "else": true, "for": true, "return": true,
		"func": true, "var": true, "const": true, "type": true,
		"package": true, "import": true, "struct": true, "interface": true,
		"switch": true, "case": true, "default": true, "break": true,
		"continue": true, "goto": true, "defer": true, "go": true,
		"select": true, "chan": true, "map": true, "range": true,
		"fn": true, "let": true, "mut": true, "pub": true, "use": true,
		"impl": true, "trait": true, "enum": true, "match": true,
		"def": true, "class": true, "from": true,
		"and": true, "or": true, "not": true, "in": true, "is": true,
		"async": true, "await": true, "yield": true,
	}
	return keywords[nodeType]
}

func plainLines(content string) []HighlightedLine {
	lines := strings.Split(content, "\n")
	result := make([]HighlightedLine, len(lines))
	for i, l := range lines {
		result[i] = HighlightedLine{Tokens: []Token{{Text: l}}}
	}
	return result
}
