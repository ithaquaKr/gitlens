package theme

import "gitlens/internal/config"

// Theme holds all colors for the diff TUI.
type Theme struct {
	Syntax SyntaxColors
	Diff   DiffColors
	UI     UIColors
}

// SyntaxColors maps to tree-sitter highlight names (same 16 as lumen).
type SyntaxColors struct {
	Keyword         string
	String          string
	Comment         string
	Function        string
	FunctionMacro   string
	Type            string
	Number          string
	Operator        string
	Variable        string
	VariableBuiltin string
	VariableMember  string
	Module          string
	Tag             string
	Attribute       string
	Label           string
	Punctuation     string
}

type DiffColors struct {
	AddedBg         string
	DeletedBg       string
	AddedWordBg     string
	DeletedWordBg   string
	AddedGutterBg   string
	DeletedGutterBg string
}

type UIColors struct {
	Border          string
	Text            string
	Selection       string
	SearchHighlight string
	StatusAdded     string
	StatusModified  string
	StatusDeleted   string
}

// Load builds a Theme from config: starts with preset, applies overrides.
func Load(cfg *config.Config) Theme {
	base := preset(cfg.Theme.Base)
	o := cfg.Theme.Override

	apply := func(dst *string, src string) {
		if src != "" {
			*dst = src
		}
	}
	apply(&base.Syntax.Keyword, o.Keyword)
	apply(&base.Syntax.String, o.String)
	apply(&base.Syntax.Comment, o.Comment)
	apply(&base.Syntax.Function, o.Function)
	apply(&base.Syntax.FunctionMacro, o.FunctionMacro)
	apply(&base.Syntax.Type, o.Type)
	apply(&base.Syntax.Number, o.Number)
	apply(&base.Syntax.Operator, o.Operator)
	apply(&base.Syntax.Variable, o.Variable)
	apply(&base.Syntax.VariableBuiltin, o.VariableBuiltin)
	apply(&base.Syntax.VariableMember, o.VariableMember)
	apply(&base.Syntax.Module, o.Module)
	apply(&base.Syntax.Tag, o.Tag)
	apply(&base.Syntax.Attribute, o.Attribute)
	apply(&base.Syntax.Label, o.Label)
	apply(&base.Syntax.Punctuation, o.Punctuation)
	apply(&base.Diff.AddedBg, o.AddedBg)
	apply(&base.Diff.DeletedBg, o.DeletedBg)
	apply(&base.Diff.AddedWordBg, o.AddedWordBg)
	apply(&base.Diff.DeletedWordBg, o.DeletedWordBg)
	apply(&base.UI.Border, o.Border)
	apply(&base.UI.Selection, o.Selection)
	return base
}
