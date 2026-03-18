# GitLens Diff TUI — Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.
>
> **Prerequisite:** Plan 1 (`2026-03-18-gitlens-foundation.md`) must be complete. This plan assumes `internal/git_entity`, `internal/vcs`, `internal/config`, and all `cmd/` commands are working.

**Goal:** Build the full interactive TUI diff viewer (`gitlens diff`) with side-by-side rendering, syntax highlighting, file tree sidebar, search, annotations, stacked mode, watch mode, and all keybindings matching lumen's layout exactly.

**Architecture:** Bubble Tea Elm-architecture (`Init` / `Update` / `View`). A single `AppState` holds all mutable state. Rendering functions in `internal/diff/render/` are pure functions of `AppState` — they take state and return styled strings via lipgloss. All side effects (file watching, clipboard, editor launch) go through `tea.Cmd`.

**Tech Stack:** Bubble Tea (event loop), Lipgloss (styling), Bubbles (viewport/textinput), go-tree-sitter + CGo grammars (syntax highlighting + context lines), sergi/go-diff (diff algorithm), fsnotify (watch mode), atotto/clipboard.

> **Build note:** Syntax highlighting and sticky context lines require CGo. Build with `CGO_ENABLED=1` and a C compiler (clang or gcc) present.

---

## File Map

```
internal/diff/
├── app.go              # Bubble Tea Model struct, Init/Update/View, keybinding dispatch
├── state.go            # AppState struct and all associated types
├── diff_algo.go        # Side-by-side diff computation: []DiffLine, []Hunk
├── context.go          # Sticky context lines via tree-sitter
├── watcher.go          # --watch mode: fsnotify goroutine, ReloadMsg
├── highlight/
│   ├── highlight.go    # Syntax highlighter: file → []HighlightedLine
│   └── languages.go    # File extension → tree-sitter language mapping
├── theme/
│   ├── theme.go        # Theme struct (SyntaxColors, DiffColors, UiColors)
│   └── presets.go      # 11 preset themes
└── render/
    ├── diffview.go     # Side-by-side panels, context header, gutters, highlights
    ├── sidebar.go      # File tree with collapse, status colors, viewed dimming
    ├── footer.go       # Branch badge, stats, keybinding hints, stacked position
    └── modal.go        # Help modal, file picker modal, annotations modal

cmd/diff.go             # gitlens diff command (wires everything together)
```

---

## Task 1: Theme System

**Files:**
- Create: `internal/diff/theme/theme.go`
- Create: `internal/diff/theme/presets.go`

- [ ] **Step 1: Write `internal/diff/theme/theme.go`**

```go
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
    AddedBg        string
    DeletedBg      string
    AddedWordBg    string
    DeletedWordBg  string
    AddedGutterBg  string
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

    // Apply non-empty overrides
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
```

- [ ] **Step 2: Write `internal/diff/theme/presets.go`** (define all 11 presets)

```go
package theme

func preset(name string) Theme {
    switch name {
    case "catppuccin-mocha":
        return catppuccinMocha()
    case "catppuccin-latte":
        return catppuccinLatte()
    case "dracula":
        return dracula()
    case "nord":
        return nord()
    case "gruvbox-dark":
        return gruvboxDark()
    case "gruvbox-light":
        return gruvboxLight()
    case "one-dark":
        return oneDark()
    case "solarized-dark":
        return solarizedDark()
    case "solarized-light":
        return solarizedLight()
    case "light":
        return lightTheme()
    default: // "dark" and fallback
        return darkTheme()
    }
}

func darkTheme() Theme {
    return Theme{
        Syntax: SyntaxColors{
            Keyword: "#569cd6", String: "#ce9178", Comment: "#6a9955",
            Function: "#dcdcaa", FunctionMacro: "#c586c0", Type: "#4ec9b0",
            Number: "#b5cea8", Operator: "#d4d4d4", Variable: "#9cdcfe",
            VariableBuiltin: "#569cd6", VariableMember: "#9cdcfe",
            Module: "#4ec9b0", Tag: "#569cd6", Attribute: "#9cdcfe",
            Label: "#c586c0", Punctuation: "#d4d4d4",
        },
        Diff: DiffColors{
            AddedBg: "#1a2e1a", DeletedBg: "#2e1a1a",
            AddedWordBg: "#2d4a2d", DeletedWordBg: "#4a2d2d",
            AddedGutterBg: "#163016", DeletedGutterBg: "#301616",
        },
        UI: UIColors{
            Border: "#444444", Text: "#d4d4d4", Selection: "#264f78",
            SearchHighlight: "#515c6a", StatusAdded: "#6a9955",
            StatusModified: "#dcdcaa", StatusDeleted: "#f44747",
        },
    }
}

func lightTheme() Theme {
    return Theme{
        Syntax: SyntaxColors{
            Keyword: "#0000ff", String: "#a31515", Comment: "#008000",
            Function: "#795e26", FunctionMacro: "#af00db", Type: "#267f99",
            Number: "#098658", Operator: "#000000", Variable: "#001080",
            VariableBuiltin: "#0000ff", VariableMember: "#001080",
            Module: "#267f99", Tag: "#800000", Attribute: "#ff0000",
            Label: "#af00db", Punctuation: "#000000",
        },
        Diff: DiffColors{
            AddedBg: "#dafada", DeletedBg: "#fadada",
            AddedWordBg: "#b5f0b5", DeletedWordBg: "#f0b5b5",
            AddedGutterBg: "#c8f0c8", DeletedGutterBg: "#f0c8c8",
        },
        UI: UIColors{
            Border: "#cccccc", Text: "#000000", Selection: "#add6ff",
            SearchHighlight: "#d6ebff", StatusAdded: "#008000",
            StatusModified: "#795e26", StatusDeleted: "#cd3131",
        },
    }
}

func catppuccinMocha() Theme {
    return Theme{
        Syntax: SyntaxColors{
            Keyword: "#cba6f7", String: "#a6e3a1", Comment: "#585b70",
            Function: "#89b4fa", FunctionMacro: "#cba6f7", Type: "#f38ba8",
            Number: "#fab387", Operator: "#cdd6f4", Variable: "#cdd6f4",
            VariableBuiltin: "#f38ba8", VariableMember: "#cdd6f4",
            Module: "#89b4fa", Tag: "#f38ba8", Attribute: "#fab387",
            Label: "#f38ba8", Punctuation: "#cdd6f4",
        },
        Diff: DiffColors{
            AddedBg: "#1e3a2a", DeletedBg: "#3a1e1e",
            AddedWordBg: "#2d5a3d", DeletedWordBg: "#5a2d2d",
            AddedGutterBg: "#182e20", DeletedGutterBg: "#2e1818",
        },
        UI: UIColors{
            Border: "#313244", Text: "#cdd6f4", Selection: "#45475a",
            SearchHighlight: "#494d64", StatusAdded: "#a6e3a1",
            StatusModified: "#f9e2af", StatusDeleted: "#f38ba8",
        },
    }
}

func catppuccinLatte() Theme {
    return Theme{
        Syntax: SyntaxColors{
            Keyword: "#8839ef", String: "#40a02b", Comment: "#9ca0b0",
            Function: "#1e66f5", FunctionMacro: "#8839ef", Type: "#d20f39",
            Number: "#fe640b", Operator: "#4c4f69", Variable: "#4c4f69",
            VariableBuiltin: "#d20f39", VariableMember: "#4c4f69",
            Module: "#1e66f5", Tag: "#d20f39", Attribute: "#fe640b",
            Label: "#8839ef", Punctuation: "#4c4f69",
        },
        Diff: DiffColors{
            AddedBg: "#d8f0de", DeletedBg: "#f0d8d8",
            AddedWordBg: "#b8e8c0", DeletedWordBg: "#e8b8b8",
            AddedGutterBg: "#c8e8d0", DeletedGutterBg: "#e8c8c8",
        },
        UI: UIColors{
            Border: "#ccd0da", Text: "#4c4f69", Selection: "#bcc0cc",
            SearchHighlight: "#dce0e8", StatusAdded: "#40a02b",
            StatusModified: "#df8e1d", StatusDeleted: "#d20f39",
        },
    }
}

func dracula() Theme {
    return Theme{
        Syntax: SyntaxColors{
            Keyword: "#ff79c6", String: "#f1fa8c", Comment: "#6272a4",
            Function: "#50fa7b", FunctionMacro: "#ff79c6", Type: "#8be9fd",
            Number: "#bd93f9", Operator: "#f8f8f2", Variable: "#f8f8f2",
            VariableBuiltin: "#ff79c6", VariableMember: "#f8f8f2",
            Module: "#8be9fd", Tag: "#ff79c6", Attribute: "#50fa7b",
            Label: "#ff79c6", Punctuation: "#f8f8f2",
        },
        Diff: DiffColors{
            AddedBg: "#1a3020", DeletedBg: "#30201a",
            AddedWordBg: "#2a4830", DeletedWordBg: "#48302a",
            AddedGutterBg: "#162818", DeletedGutterBg: "#281816",
        },
        UI: UIColors{
            Border: "#44475a", Text: "#f8f8f2", Selection: "#44475a",
            SearchHighlight: "#6272a4", StatusAdded: "#50fa7b",
            StatusModified: "#f1fa8c", StatusDeleted: "#ff5555",
        },
    }
}

func nord() Theme {
    return Theme{
        Syntax: SyntaxColors{
            Keyword: "#81a1c1", String: "#a3be8c", Comment: "#616e88",
            Function: "#88c0d0", FunctionMacro: "#b48ead", Type: "#8fbcbb",
            Number: "#b48ead", Operator: "#eceff4", Variable: "#d8dee9",
            VariableBuiltin: "#81a1c1", VariableMember: "#d8dee9",
            Module: "#8fbcbb", Tag: "#81a1c1", Attribute: "#88c0d0",
            Label: "#b48ead", Punctuation: "#eceff4",
        },
        Diff: DiffColors{
            AddedBg: "#1e2e1e", DeletedBg: "#2e1e1e",
            AddedWordBg: "#2d422d", DeletedWordBg: "#42302d",
            AddedGutterBg: "#192619", DeletedGutterBg: "#261919",
        },
        UI: UIColors{
            Border: "#3b4252", Text: "#eceff4", Selection: "#434c5e",
            SearchHighlight: "#4c566a", StatusAdded: "#a3be8c",
            StatusModified: "#ebcb8b", StatusDeleted: "#bf616a",
        },
    }
}

func gruvboxDark() Theme {
    return Theme{
        Syntax: SyntaxColors{
            Keyword: "#fb4934", String: "#b8bb26", Comment: "#928374",
            Function: "#fabd2f", FunctionMacro: "#fe8019", Type: "#8ec07c",
            Number: "#d3869b", Operator: "#ebdbb2", Variable: "#ebdbb2",
            VariableBuiltin: "#fb4934", VariableMember: "#ebdbb2",
            Module: "#83a598", Tag: "#fb4934", Attribute: "#fabd2f",
            Label: "#fe8019", Punctuation: "#ebdbb2",
        },
        Diff: DiffColors{
            AddedBg: "#1d2b1d", DeletedBg: "#2b1d1d",
            AddedWordBg: "#2d402d", DeletedWordBg: "#40302d",
            AddedGutterBg: "#192319", DeletedGutterBg: "#231919",
        },
        UI: UIColors{
            Border: "#504945", Text: "#ebdbb2", Selection: "#3c3836",
            SearchHighlight: "#504945", StatusAdded: "#b8bb26",
            StatusModified: "#fabd2f", StatusDeleted: "#fb4934",
        },
    }
}

func gruvboxLight() Theme {
    return Theme{
        Syntax: SyntaxColors{
            Keyword: "#9d0006", String: "#79740e", Comment: "#928374",
            Function: "#b57614", FunctionMacro: "#af3a03", Type: "#427b58",
            Number: "#8f3f71", Operator: "#3c3836", Variable: "#3c3836",
            VariableBuiltin: "#9d0006", VariableMember: "#3c3836",
            Module: "#076678", Tag: "#9d0006", Attribute: "#b57614",
            Label: "#af3a03", Punctuation: "#3c3836",
        },
        Diff: DiffColors{
            AddedBg: "#daeada", DeletedBg: "#eadada",
            AddedWordBg: "#b8d8b8", DeletedWordBg: "#d8b8b8",
            AddedGutterBg: "#c8d8c8", DeletedGutterBg: "#d8c8c8",
        },
        UI: UIColors{
            Border: "#d5c4a1", Text: "#3c3836", Selection: "#ebdbb2",
            SearchHighlight: "#f2e5bc", StatusAdded: "#79740e",
            StatusModified: "#b57614", StatusDeleted: "#9d0006",
        },
    }
}

func oneDark() Theme {
    return Theme{
        Syntax: SyntaxColors{
            Keyword: "#c678dd", String: "#98c379", Comment: "#5c6370",
            Function: "#61afef", FunctionMacro: "#c678dd", Type: "#e5c07b",
            Number: "#d19a66", Operator: "#abb2bf", Variable: "#abb2bf",
            VariableBuiltin: "#e06c75", VariableMember: "#abb2bf",
            Module: "#61afef", Tag: "#e06c75", Attribute: "#d19a66",
            Label: "#c678dd", Punctuation: "#abb2bf",
        },
        Diff: DiffColors{
            AddedBg: "#1d2b1d", DeletedBg: "#2b1d1d",
            AddedWordBg: "#2d3f2d", DeletedWordBg: "#3f2d2d",
            AddedGutterBg: "#192519", DeletedGutterBg: "#251919",
        },
        UI: UIColors{
            Border: "#3e4452", Text: "#abb2bf", Selection: "#3e4452",
            SearchHighlight: "#3e4452", StatusAdded: "#98c379",
            StatusModified: "#e5c07b", StatusDeleted: "#e06c75",
        },
    }
}

func solarizedDark() Theme {
    return Theme{
        Syntax: SyntaxColors{
            Keyword: "#859900", String: "#2aa198", Comment: "#586e75",
            Function: "#268bd2", FunctionMacro: "#d33682", Type: "#b58900",
            Number: "#d33682", Operator: "#839496", Variable: "#839496",
            VariableBuiltin: "#859900", VariableMember: "#839496",
            Module: "#268bd2", Tag: "#859900", Attribute: "#cb4b16",
            Label: "#d33682", Punctuation: "#839496",
        },
        Diff: DiffColors{
            AddedBg: "#1a2b1a", DeletedBg: "#2b1a1a",
            AddedWordBg: "#253c25", DeletedWordBg: "#3c2525",
            AddedGutterBg: "#162416", DeletedGutterBg: "#241616",
        },
        UI: UIColors{
            Border: "#073642", Text: "#839496", Selection: "#073642",
            SearchHighlight: "#073642", StatusAdded: "#859900",
            StatusModified: "#b58900", StatusDeleted: "#dc322f",
        },
    }
}

func solarizedLight() Theme {
    return Theme{
        Syntax: SyntaxColors{
            Keyword: "#859900", String: "#2aa198", Comment: "#93a1a1",
            Function: "#268bd2", FunctionMacro: "#d33682", Type: "#b58900",
            Number: "#d33682", Operator: "#657b83", Variable: "#657b83",
            VariableBuiltin: "#859900", VariableMember: "#657b83",
            Module: "#268bd2", Tag: "#859900", Attribute: "#cb4b16",
            Label: "#d33682", Punctuation: "#657b83",
        },
        Diff: DiffColors{
            AddedBg: "#daeada", DeletedBg: "#eadada",
            AddedWordBg: "#b8d8b8", DeletedWordBg: "#d8b8b8",
            AddedGutterBg: "#c8d8c8", DeletedGutterBg: "#d8c8c8",
        },
        UI: UIColors{
            Border: "#eee8d5", Text: "#657b83", Selection: "#eee8d5",
            SearchHighlight: "#fdf6e3", StatusAdded: "#859900",
            StatusModified: "#b58900", StatusDeleted: "#dc322f",
        },
    }
}
```

- [ ] **Step 3: Write theme tests**

```go
// internal/diff/theme/theme_test.go
package theme_test

import (
    "testing"
    "gitlens/internal/config"
    "gitlens/internal/diff/theme"
)

func TestLoadDefaultPreset(t *testing.T) {
    cfg := &config.Config{Theme: config.ThemeConfig{Base: "dark"}}
    th := theme.Load(cfg)
    if th.Syntax.Keyword == "" {
        t.Error("expected non-empty keyword color")
    }
}

func TestLoadWithOverride(t *testing.T) {
    cfg := &config.Config{
        Theme: config.ThemeConfig{
            Base: "dark",
            Override: config.ThemeOverride{Keyword: "#ff0000"},
        },
    }
    th := theme.Load(cfg)
    if th.Syntax.Keyword != "#ff0000" {
        t.Errorf("override not applied: got %q", th.Syntax.Keyword)
    }
}

func TestAllPresetsLoad(t *testing.T) {
    presets := []string{"dark", "light", "catppuccin-mocha", "catppuccin-latte",
        "dracula", "nord", "gruvbox-dark", "gruvbox-light", "one-dark",
        "solarized-dark", "solarized-light"}
    for _, name := range presets {
        cfg := &config.Config{Theme: config.ThemeConfig{Base: name}}
        th := theme.Load(cfg)
        if th.Syntax.Keyword == "" {
            t.Errorf("preset %q: empty keyword color", name)
        }
    }
}
```

- [ ] **Step 4: Run tests**

```bash
go test ./internal/diff/theme/... -v
```

Expected: all 3 tests PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/diff/theme/
git commit -m "feat: add theme system with 11 presets and config overrides"
```

---

## Task 2: Diff Algorithm

**Files:**
- Create: `internal/diff/diff_algo.go`
- Create: `internal/diff/diff_algo_test.go`

- [ ] **Step 1: Write failing tests**

```go
// internal/diff/diff_algo_test.go
package diff_test

import (
    "testing"
    "gitlens/internal/diff"
    "gitlens/internal/git_entity"
)

func TestComputeAdded(t *testing.T) {
    lines := diff.ComputeDiffLines("", "hello\nworld\n")
    if len(lines) == 0 {
        t.Fatal("expected diff lines")
    }
    for _, l := range lines {
        if l.ChangeType != git_entity.Insert {
            t.Errorf("expected Insert, got %v", l.ChangeType)
        }
    }
}

func TestComputeDeleted(t *testing.T) {
    lines := diff.ComputeDiffLines("hello\nworld\n", "")
    for _, l := range lines {
        if l.ChangeType != git_entity.Delete {
            t.Errorf("expected Delete, got %v", l.ChangeType)
        }
    }
}

func TestComputeEqual(t *testing.T) {
    lines := diff.ComputeDiffLines("hello\n", "hello\n")
    for _, l := range lines {
        if l.ChangeType != git_entity.Equal {
            t.Errorf("expected Equal, got %v", l.ChangeType)
        }
    }
}

func TestComputeModified(t *testing.T) {
    lines := diff.ComputeDiffLines("hello world\n", "hello earth\n")
    found := false
    for _, l := range lines {
        if l.ChangeType == git_entity.Modified {
            found = true
            // Word-level segments should be populated
            if len(l.OldSegments) == 0 && len(l.NewSegments) == 0 {
                t.Error("expected word-level segments on Modified line")
            }
        }
    }
    if !found {
        t.Error("expected at least one Modified line")
    }
}

func TestComputeHunks(t *testing.T) {
    old := "a\nb\nc\nd\ne\nf\ng\n"
    new := "a\nb\nX\nd\ne\nf\ng\n" // line 3 changed
    lines := diff.ComputeDiffLines(old, new)
    hunks := diff.ComputeHunks(lines)
    if len(hunks) == 0 {
        t.Error("expected at least one hunk")
    }
}
```

- [ ] **Step 2: Run to confirm failure**

```bash
go test ./internal/diff/... -run TestCompute -v
```

Expected: compilation error.

- [ ] **Step 3: Write `internal/diff/diff_algo.go`**

```go
package diff

import (
    "strings"

    "github.com/sergi/go-diff/diffmatchpatch"
    "gitlens/internal/git_entity"
)

// ComputeDiffLines produces a side-by-side []DiffLine from two file content strings.
func ComputeDiffLines(oldContent, newContent string) []git_entity.DiffLine {
    dmp := diffmatchpatch.New()
    oldLines := splitLines(oldContent)
    newLines := splitLines(newContent)

    // Line-level diff
    chars1, chars2, lineArray := dmp.DiffLinesToChars(
        strings.Join(oldLines, ""),
        strings.Join(newLines, ""),
    )
    diffs := dmp.DiffMain(chars1, chars2, false)
    diffs = dmp.DiffCharsToLines(diffs, lineArray)

    var result []git_entity.DiffLine
    oldIdx, newIdx := 1, 1

    for _, d := range diffs {
        lines := splitLines(d.Text)
        switch d.Type {
        case diffmatchpatch.DiffEqual:
            for _, l := range lines {
                if l == "" {
                    continue
                }
                result = append(result, git_entity.DiffLine{
                    OldLine:    &git_entity.LineContent{LineNo: oldIdx, Text: l},
                    NewLine:    &git_entity.LineContent{LineNo: newIdx, Text: l},
                    ChangeType: git_entity.Equal,
                })
                oldIdx++
                newIdx++
            }
        case diffmatchpatch.DiffDelete:
            for _, l := range lines {
                if l == "" {
                    continue
                }
                result = append(result, git_entity.DiffLine{
                    OldLine:    &git_entity.LineContent{LineNo: oldIdx, Text: l},
                    ChangeType: git_entity.Delete,
                })
                oldIdx++
            }
        case diffmatchpatch.DiffInsert:
            for _, l := range lines {
                if l == "" {
                    continue
                }
                result = append(result, git_entity.DiffLine{
                    NewLine:    &git_entity.LineContent{LineNo: newIdx, Text: l},
                    ChangeType: git_entity.Insert,
                })
                newIdx++
            }
        }
    }

    // Pair adjacent Delete+Insert into Modified
    result = pairModified(result)
    return result
}

// pairModified pairs consecutive Delete and Insert lines into Modified,
// adding word-level diff segments.
func pairModified(lines []git_entity.DiffLine) []git_entity.DiffLine {
    dmp := diffmatchpatch.New()
    result := make([]git_entity.DiffLine, 0, len(lines))
    i := 0
    for i < len(lines) {
        if i+1 < len(lines) &&
            lines[i].ChangeType == git_entity.Delete &&
            lines[i+1].ChangeType == git_entity.Insert {
            old := lines[i].OldLine.Text
            new := lines[i+1].NewLine.Text
            // Apply word-level diff only when >20% content is unchanged
            oldSegs, newSegs := wordDiff(dmp, old, new)
            result = append(result, git_entity.DiffLine{
                OldLine:     lines[i].OldLine,
                NewLine:     lines[i+1].NewLine,
                ChangeType:  git_entity.Modified,
                OldSegments: oldSegs,
                NewSegments: newSegs,
            })
            i += 2
            continue
        }
        result = append(result, lines[i])
        i++
    }
    return result
}

// wordDiff computes character-level diff segments for a modified line pair.
// Returns empty segments if less than 20% of content is unchanged.
func wordDiff(dmp *diffmatchpatch.DiffMatchPatch, old, new string) ([]git_entity.Segment, []git_entity.Segment) {
    diffs := dmp.DiffMain(old, new, false)
    diffs = dmp.DiffCleanupSemantic(diffs)

    // Count unchanged chars
    unchanged := 0
    total := len(old) + len(new)
    for _, d := range diffs {
        if d.Type == diffmatchpatch.DiffEqual {
            unchanged += len(d.Text) * 2
        }
    }
    if total > 0 && float64(unchanged)/float64(total) < 0.2 {
        // Less than 20% unchanged — skip word highlights
        return []git_entity.Segment{{Text: old}}, []git_entity.Segment{{Text: new}}
    }

    var oldSegs, newSegs []git_entity.Segment
    for _, d := range diffs {
        switch d.Type {
        case diffmatchpatch.DiffEqual:
            oldSegs = append(oldSegs, git_entity.Segment{Text: d.Text})
            newSegs = append(newSegs, git_entity.Segment{Text: d.Text})
        case diffmatchpatch.DiffDelete:
            oldSegs = append(oldSegs, git_entity.Segment{Text: d.Text, Highlight: true})
        case diffmatchpatch.DiffInsert:
            newSegs = append(newSegs, git_entity.Segment{Text: d.Text, Highlight: true})
        }
    }
    return oldSegs, newSegs
}

// ComputeHunks identifies contiguous blocks of non-Equal lines.
func ComputeHunks(lines []git_entity.DiffLine) []git_entity.Hunk {
    var hunks []git_entity.Hunk
    inHunk := false
    start := 0
    for i, l := range lines {
        if l.ChangeType != git_entity.Equal {
            if !inHunk {
                start = i
                inHunk = true
            }
        } else {
            if inHunk {
                hunks = append(hunks, git_entity.Hunk{StartIdx: start, EndIdx: i - 1})
                inHunk = false
            }
        }
    }
    if inHunk {
        hunks = append(hunks, git_entity.Hunk{StartIdx: start, EndIdx: len(lines) - 1})
    }
    return hunks
}

func splitLines(s string) []string {
    if s == "" {
        return nil
    }
    lines := strings.Split(s, "\n")
    // Re-add newlines (dmp works with newline-terminated lines)
    result := make([]string, 0, len(lines))
    for _, l := range lines {
        if l != "" {
            result = append(result, l+"\n")
        }
    }
    return result
}
```

- [ ] **Step 4: Run tests**

```bash
go test ./internal/diff/... -run TestCompute -v
```

Expected: all tests PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/diff/diff_algo.go internal/diff/diff_algo_test.go
git commit -m "feat: add side-by-side diff algorithm with word-level highlighting"
```

---

## Task 3: Syntax Highlighting

**Files:**
- Create: `internal/diff/highlight/languages.go`
- Create: `internal/diff/highlight/highlight.go`

> **CGo note:** This task requires `CGO_ENABLED=1`. If CGo is not available, stub out the highlighter to return plain-text lines and add a TODO comment.

- [ ] **Step 1: Install tree-sitter language packages**

```bash
go get github.com/smacker/go-tree-sitter@latest
go get github.com/smacker/go-tree-sitter/golang@latest
go get github.com/smacker/go-tree-sitter/rust@latest
go get github.com/smacker/go-tree-sitter/typescript/typescript@latest
go get github.com/smacker/go-tree-sitter/typescript/tsx@latest
go get github.com/smacker/go-tree-sitter/javascript@latest
go get github.com/smacker/go-tree-sitter/python@latest
go get github.com/smacker/go-tree-sitter/c_sharp@latest
go mod tidy
```

- [ ] **Step 2: Write `internal/diff/highlight/languages.go`**

```go
package highlight

import (
    "path/filepath"
    "strings"

    sitter "github.com/smacker/go-tree-sitter"
    "github.com/smacker/go-tree-sitter/c_sharp"
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
        return c_sharp.GetLanguage()
    default:
        return nil
    }
}
```

- [ ] **Step 3: Write `internal/diff/highlight/highlight.go`**

```go
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

// Highlighter caches parsed trees per file.
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

    // Walk the tree and collect leaf nodes with their types
    sourceBytes := []byte(content)
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

    // Map node types to colors
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
            // Keywords: Go, Rust, Python etc. keywords share "keyword" node type
            if isKeyword(nodeType) {
                return h.theme.Syntax.Keyword
            }
            return ""
        }
    }

    // Reconstruct highlighted lines
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
        offset = lineEnd + 1 // +1 for \n
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
        "def": true, "class": true, "import": true, "from": true,
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
```

- [ ] **Step 4: Verify build with CGo**

```bash
CGO_ENABLED=1 go build ./...
```

- [ ] **Step 5: Commit**

```bash
git add internal/diff/highlight/
git commit -m "feat: add tree-sitter syntax highlighting"
```

---

## Task 4: Sticky Context Lines

**Files:**
- Create: `internal/diff/context.go`

- [ ] **Step 1: Write `internal/diff/context.go`**

```go
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

// ComputeContextLines returns the enclosing scope headers (up to maxLines=5)
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

    // Find byte offset of currentLine
    offset := 0
    for i := 0; i < currentLine && i < len(lines); i++ {
        offset += len(lines[i]) + 1
    }

    // Walk up the tree to find enclosing scope nodes
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
                // Get the first line of this node
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

    // Return up to 5, innermost last
    if len(enclosing) > 5 {
        enclosing = enclosing[len(enclosing)-5:]
    }
    return enclosing
}
```

- [ ] **Step 2: Verify `LanguageForFile` is exported**

The function is already named `LanguageForFile` (exported) in `languages.go` from Task 3. No rename needed.

- [ ] **Step 3: Verify build**

```bash
CGO_ENABLED=1 go build ./...
```

- [ ] **Step 4: Commit**

```bash
git add internal/diff/context.go
git commit -m "feat: add sticky context lines via tree-sitter"
```

---

## Task 5: AppState

**Files:**
- Create: `internal/diff/state.go`

- [ ] **Step 1: Write `internal/diff/state.go`**

```go
package diff

import (
    "time"

    "gitlens/internal/git_entity"
)

// --- AppState types ---

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
    // File list
    Files          []git_entity.FileDiff
    CurrentFileIdx int

    // Computed rendering data for current file
    DiffLines []git_entity.DiffLine
    Hunks     []git_entity.Hunk

    // Scroll
    ScrollY        int
    ScrollX        int
    SidebarScrollX int

    // Navigation
    ContextLines  []ContextLine
    PendingKey    PendingKey
    Fullscreen    DiffFullscreen
    Focus         FocusArea

    // Sidebar
    SidebarCollapsed bool
    SidebarSelected  int
    CollapsedDirs    map[string]bool

    // Selection
    Anchor               *CursorPos
    Head                 *CursorPos
    SelectionMode        SelectionMode
    SelectionPanel       PanelFocus
    ShowSelectionTooltip bool

    // Annotations
    Annotations []Annotation

    // Search
    SearchActive  bool
    SearchQuery   string
    SearchMatches []MatchPos
    SearchIdx     int

    // Viewed files
    ViewedFiles map[string]struct{} // keyed by filename

    // Stacked mode
    StackedMode         bool
    StackedCommits      []*git_entity.Commit
    CurrentCommitIdx    int
    StackedViewedFiles  map[string]map[string]struct{} // SHA → filename → viewed

    // Layout (recomputed each View, used by Update for mouse coords)
    Layout PanelLayout

    // Terminal size
    Width, Height int
}

// NewAppState initializes AppState with empty collections.
func NewAppState(files []git_entity.FileDiff) *AppState {
    s := &AppState{
        Files:          files,
        CollapsedDirs:  make(map[string]bool),
        ViewedFiles:    make(map[string]struct{}),
        StackedViewedFiles: make(map[string]map[string]struct{}),
    }
    if len(files) > 0 {
        s.recompute()
    }
    return s
}

// recompute recalculates DiffLines and Hunks for the current file.
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

// NavigateToFile changes the current file and recomputes diff.
func (s *AppState) NavigateToFile(idx int) {
    if idx < 0 || idx >= len(s.Files) {
        return
    }
    s.CurrentFileIdx = idx
    s.recompute()
}

// ToggleViewed marks/unmarks the current file as viewed.
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

// NextUnviewed advances to the next unviewed file, wrapping around.
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

// CurrentHunk returns the hunk index containing the current scroll position.
func (s *AppState) CurrentHunkIdx() int {
    for i, h := range s.Hunks {
        if s.ScrollY >= h.StartIdx && s.ScrollY <= h.EndIdx {
            return i
        }
    }
    return -1
}
```

- [ ] **Step 2: Verify build**

```bash
CGO_ENABLED=1 go build ./...
```

- [ ] **Step 3: Commit**

```bash
git add internal/diff/state.go
git commit -m "feat: add AppState with all TUI state fields"
```

---

## Task 6: Watcher

**Files:**
- Create: `internal/diff/watcher.go`

- [ ] **Step 1: Write `internal/diff/watcher.go`**

```go
package diff

import (
    "time"

    "github.com/fsnotify/fsnotify"
    tea "github.com/charmbracelet/bubbletea"
)

// ReloadMsg is sent to the Bubble Tea program when a watched file changes.
type ReloadMsg struct{}

// WatchFiles starts a goroutine that watches the given paths and sends
// ReloadMsg to the program when any of them change. Debounced at 200ms.
// Call cancel() to stop the watcher.
func WatchFiles(p *tea.Program, paths []string) (cancel func(), err error) {
    watcher, err := fsnotify.NewWatcher()
    if err != nil {
        return nil, err
    }
    for _, path := range paths {
        watcher.Add(path)
    }
    done := make(chan struct{})
    go func() {
        defer watcher.Close()
        timer := time.NewTimer(0)
        timer.Stop()
        for {
            select {
            case <-done:
                return
            case _, ok := <-watcher.Events:
                if !ok {
                    return
                }
                timer.Reset(200 * time.Millisecond)
            case err, ok := <-watcher.Errors:
                _ = err
                if !ok {
                    return
                }
            case <-timer.C:
                p.Send(ReloadMsg{})
            }
        }
    }()
    return func() { close(done) }, nil
}
```

- [ ] **Step 2: Verify build**

```bash
CGO_ENABLED=1 go build ./...
```

- [ ] **Step 3: Commit**

```bash
git add internal/diff/watcher.go
git commit -m "feat: add file watcher for --watch mode"
```

---

## Task 7: Renderers

**Files:**
- Create: `internal/diff/render/sidebar.go`
- Create: `internal/diff/render/footer.go`
- Create: `internal/diff/render/diffview.go`
- Create: `internal/diff/render/modal.go`

- [ ] **Step 1: Write `internal/diff/render/sidebar.go`**

```go
package render

import (
    "fmt"
    "path/filepath"
    "strings"

    "github.com/charmbracelet/lipgloss"
    "gitlens/internal/diff"
    "gitlens/internal/diff/theme"
)

// Sidebar renders the file tree panel.
func Sidebar(state *diff.AppState, th theme.Theme, width int) string {
    if state.SidebarCollapsed {
        return ""
    }
    var b strings.Builder
    borderStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(th.UI.Border))

    for i, f := range state.Files {
        _, isViewed := state.ViewedFiles[f.Path]
        dir := filepath.Dir(f.Path)
        name := filepath.Base(f.Path)

        // Status color
        var statusColor string
        switch f.Status {
        case "A":
            statusColor = th.UI.StatusAdded
        case "D":
            statusColor = th.UI.StatusDeleted
        default:
            statusColor = th.UI.StatusModified
        }

        statusBadge := lipgloss.NewStyle().Foreground(lipgloss.Color(statusColor)).Render(f.Status)
        nameStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(th.UI.Text))
        if isViewed {
            nameStyle = nameStyle.Faint(true)
        }
        if i == state.CurrentFileIdx {
            nameStyle = nameStyle.Background(lipgloss.Color(th.UI.Selection)).Bold(true)
        }
        if i == state.SidebarSelected && state.Focus == diff.FocusSidebar {
            nameStyle = nameStyle.Underline(true)
        }

        indent := ""
        if dir != "." {
            indent = "  "
        }
        line := fmt.Sprintf("%s%s %s%s",
            indent,
            statusBadge,
            nameStyle.Render(name),
            borderStyle.Render(""),
        )
        b.WriteString(line + "\n")
        _ = dir
    }
    return lipgloss.NewStyle().
        Width(width).
        BorderRight(true).
        BorderStyle(lipgloss.NormalBorder()).
        BorderForeground(lipgloss.Color(th.UI.Border)).
        Render(b.String())
}
```

- [ ] **Step 2: Write `internal/diff/render/footer.go`**

```go
package render

import (
    "fmt"
    "strings"

    "github.com/charmbracelet/lipgloss"
    "gitlens/internal/diff"
    "gitlens/internal/diff/theme"
)

// Footer renders the bottom status bar.
func Footer(state *diff.AppState, th theme.Theme, width int, branchName string) string {
    textStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(th.UI.Text))
    borderStyle := lipgloss.NewStyle().
        Background(lipgloss.Color(th.UI.Border)).
        Foreground(lipgloss.Color(th.UI.Text))

    // Branch badge
    branch := borderStyle.Padding(0, 1).Render(" " + branchName + " ")

    // File stats
    added, deleted := countChanges(state)
    stats := textStyle.Render(fmt.Sprintf("+%d -%d", added, deleted))

    // Stacked mode badge
    stacked := ""
    if state.StackedMode && len(state.StackedCommits) > 0 {
        current := state.StackedCommits[state.CurrentCommitIdx]
        shortSHA := current.Hash[:7]
        stacked = borderStyle.Padding(0, 1).Render(
            fmt.Sprintf("[%d/%d] %s", state.CurrentCommitIdx+1, len(state.StackedCommits), shortSHA),
        )
    }

    // Keybinding hints
    hints := textStyle.Faint(true).Render("q quit  ? help  tab sidebar  / search")

    parts := []string{branch, stats}
    if stacked != "" {
        parts = append(parts, stacked)
    }
    parts = append(parts, strings.Repeat(" ", max(0, width-lipgloss.Width(strings.Join(parts, " "))-lipgloss.Width(hints)-4)))
    parts = append(parts, hints)

    return lipgloss.NewStyle().Width(width).Render(strings.Join(parts, "  "))
}

func countChanges(state *diff.AppState) (added, deleted int) {
    for _, l := range state.DiffLines {
        switch l.ChangeType {
        case git_entity.Insert, git_entity.Modified:
            added++
        case git_entity.Delete:
            deleted++
        }
    }
    return
}

func max(a, b int) int {
    if a > b { return a }
    return b
}
```

> **Note:** `countChanges` needs `git_entity` import: `"gitlens/internal/git_entity"`

- [ ] **Step 3: Write `internal/diff/render/diffview.go`**

```go
package render

import (
    "fmt"
    "strings"

    "github.com/charmbracelet/lipgloss"
    "gitlens/internal/diff"
    "gitlens/internal/diff/theme"
    "gitlens/internal/git_entity"
    hl "gitlens/internal/diff/highlight"
)

// DiffView renders the main side-by-side diff area.
func DiffView(state *diff.AppState, th theme.Theme, width, height int) string {
    if len(state.Files) == 0 {
        return lipgloss.NewStyle().Foreground(lipgloss.Color(th.UI.Text)).
            Render("No changes to display")
    }

    currentFile := state.Files[state.CurrentFileIdx]
    highlighter := hl.New(th)
    oldHL := highlighter.HighlightFile(currentFile.Path, currentFile.OldContent)
    newHL := highlighter.HighlightFile(currentFile.Path, currentFile.NewContent)

    panelWidth := (width - state.Layout.GutterWidth*2 - 1) / 2

    var b strings.Builder

    // Sticky context header
    for _, cl := range state.ContextLines {
        ctxStyle := lipgloss.NewStyle().
            Background(lipgloss.Color(th.UI.Border)).
            Foreground(lipgloss.Color(th.UI.Text)).
            Width(width)
        b.WriteString(ctxStyle.Render(fmt.Sprintf("  %d: %s", cl.LineNumber, cl.Content)) + "\n")
    }

    // Render visible diff lines
    visibleLines := state.DiffLines
    start := state.ScrollY
    end := start + height - len(state.ContextLines) - 2 // -2 for header/footer
    if start < 0 {
        start = 0
    }
    if end > len(visibleLines) {
        end = len(visibleLines)
    }

    gutterWidth := state.Layout.GutterWidth
    if gutterWidth == 0 {
        gutterWidth = 5
    }

    for i := start; i < end; i++ {
        line := visibleLines[i]
        b.WriteString(renderDiffLine(line, th, panelWidth, gutterWidth, oldHL, newHL) + "\n")
    }

    return b.String()
}

func renderDiffLine(line git_entity.DiffLine, th theme.Theme, panelWidth, gutterWidth int, oldHL, newHL []hl.HighlightedLine) string {
    var oldBg, newBg lipgloss.Color
    switch line.ChangeType {
    case git_entity.Delete:
        oldBg = lipgloss.Color(th.Diff.DeletedBg)
    case git_entity.Insert:
        newBg = lipgloss.Color(th.Diff.AddedBg)
    case git_entity.Modified:
        oldBg = lipgloss.Color(th.Diff.DeletedBg)
        newBg = lipgloss.Color(th.Diff.AddedBg)
    }

    oldGutter := renderGutter(line.OldLine, line.ChangeType, true, th, gutterWidth)
    newGutter := renderGutter(line.NewLine, line.ChangeType, false, th, gutterWidth)

    oldPanel := renderPanel(line.OldLine, line.OldSegments, oldBg, panelWidth, oldHL)
    newPanel := renderPanel(line.NewLine, line.NewSegments, newBg, panelWidth, newHL)

    divider := lipgloss.NewStyle().Foreground(lipgloss.Color(th.UI.Border)).Render("│")

    return oldGutter + oldPanel + divider + newGutter + newPanel
}

func renderGutter(lc *git_entity.LineContent, ct git_entity.ChangeType, isOld bool, th theme.Theme, width int) string {
    var bg lipgloss.Color
    switch ct {
    case git_entity.Delete:
        if isOld {
            bg = lipgloss.Color(th.Diff.DeletedGutterBg)
        }
    case git_entity.Insert:
        if !isOld {
            bg = lipgloss.Color(th.Diff.AddedGutterBg)
        }
    case git_entity.Modified:
        if isOld {
            bg = lipgloss.Color(th.Diff.DeletedGutterBg)
        } else {
            bg = lipgloss.Color(th.Diff.AddedGutterBg)
        }
    }
    style := lipgloss.NewStyle().Width(width).Background(bg).
        Foreground(lipgloss.Color(th.UI.Text)).Faint(true)
    if lc != nil {
        return style.Render(fmt.Sprintf("%*d", width-1, lc.LineNo))
    }
    return style.Render(strings.Repeat(" ", width))
}

func renderPanel(lc *git_entity.LineContent, segs []git_entity.Segment, bg lipgloss.Color, width int, hlLines []hl.HighlightedLine) string {
    style := lipgloss.NewStyle().Width(width).Background(bg)
    if lc == nil {
        return style.Render("")
    }
    if len(segs) > 0 {
        // Word-level diff rendering
        var sb strings.Builder
        for _, seg := range segs {
            segStyle := lipgloss.NewStyle()
            if bg != "" {
                segStyle = segStyle.Background(bg)
            }
            if seg.Highlight {
                if bg == lipgloss.Color("") {
                    segStyle = segStyle.Background(lipgloss.Color("#333333"))
                }
            }
            sb.WriteString(segStyle.Render(seg.Text))
        }
        return style.Render(sb.String())
    }
    // Fallback to plain text
    return style.Render(lc.Text)
}
```

- [ ] **Step 4: Write `internal/diff/render/modal.go`**

```go
package render

import (
    "fmt"
    "strings"

    "github.com/charmbracelet/lipgloss"
    "gitlens/internal/diff"
    "gitlens/internal/diff/theme"
)

// HelpModal renders the keybindings help overlay.
func HelpModal(th theme.Theme, width, height int) string {
    help := [][]string{
        {"j/k", "scroll down/up"},
        {"ctrl+d/u", "half-page scroll"},
        {"g g / G", "top / bottom"},
        {"h/l", "scroll left/right"},
        {"{/}", "prev/next hunk"},
        {"ctrl+j/k", "prev/next file"},
        {"ctrl+p", "file picker"},
        {"[/]/=", "fullscreen old/new/reset"},
        {"tab", "toggle sidebar"},
        {"1/2", "focus sidebar/diff"},
        {"e", "open in $EDITOR"},
        {"y", "copy selection"},
        {"i", "annotate"},
        {"I", "view annotations"},
        {"space", "mark file viewed"},
        {"/", "search"},
        {"n/N", "next/prev match"},
        {"?", "toggle help"},
        {"q", "quit"},
        {"ctrl+l/h", "next/prev commit (stacked)"},
    }
    boxWidth := 50
    var b strings.Builder
    b.WriteString(lipgloss.NewStyle().Bold(true).Render("Keybindings") + "\n\n")
    for _, row := range help {
        key := lipgloss.NewStyle().Foreground(lipgloss.Color(th.Syntax.Keyword)).
            Width(14).Render(row[0])
        desc := lipgloss.NewStyle().Foreground(lipgloss.Color(th.UI.Text)).Render(row[1])
        b.WriteString(key + "  " + desc + "\n")
    }
    return lipgloss.NewStyle().
        Width(boxWidth).
        Border(lipgloss.RoundedBorder()).
        BorderForeground(lipgloss.Color(th.UI.Border)).
        Padding(1, 2).
        Render(b.String())
}

// AnnotationsModal renders the annotations list overlay.
func AnnotationsModal(state *diff.AppState, th theme.Theme) string {
    if len(state.Annotations) == 0 {
        return lipgloss.NewStyle().
            Border(lipgloss.RoundedBorder()).
            BorderForeground(lipgloss.Color(th.UI.Border)).
            Padding(1, 2).
            Render("No annotations yet.\n\nPress i to annotate a selection.")
    }

    var b strings.Builder
    b.WriteString(lipgloss.NewStyle().Bold(true).Render("Annotations") + "\n\n")
    for i, ann := range state.Annotations {
        loc := ann.Filename
        if ann.Target.Kind == diff.TargetLineRange {
            panel := "left"
            if ann.Target.Panel == diff.PanelNew {
                panel = "right"
            }
            loc = fmt.Sprintf("%s:%s L%d-%d", ann.Filename, panel, ann.Target.StartLine, ann.Target.EndLine)
        }
        header := lipgloss.NewStyle().
            Foreground(lipgloss.Color(th.Syntax.Function)).
            Render(fmt.Sprintf("[%d] %s", i+1, loc))
        preview := ann.Content
        if len(preview) > 60 {
            preview = preview[:60] + "..."
        }
        b.WriteString(header + "\n")
        b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(th.UI.Text)).Faint(true).
            Render("  "+preview) + "\n\n")
    }
    b.WriteString(lipgloss.NewStyle().Faint(true).Render("d delete  y copy  enter jump  esc close"))
    return lipgloss.NewStyle().
        Width(70).
        Border(lipgloss.RoundedBorder()).
        BorderForeground(lipgloss.Color(th.UI.Border)).
        Padding(1, 2).
        Render(b.String())
}
```

- [ ] **Step 5: Verify build**

```bash
CGO_ENABLED=1 go build ./...
```

- [ ] **Step 6: Commit**

```bash
git add internal/diff/render/
git commit -m "feat: add TUI renderers (sidebar, footer, diffview, modals)"
```

---

## Task 8: Bubble Tea App Model

**Files:**
- Create: `internal/diff/app.go`

This is the central event loop. It handles all keypresses, mouse events, and messages.

- [ ] **Step 1: Write `internal/diff/app.go`**

```go
package diff

import (
    "fmt"
    "os"
    "os/exec"
    "strings"

    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/lipgloss"
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
    state      *AppState
    theme      theme.Theme
    modal      modalKind
    branchName string
    editInput  string // annotation editor text
    cancelWatch func()
}

// NewModel creates the diff TUI model.
func NewModel(state *AppState, th theme.Theme, branchName string) Model {
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

    case ReloadMsg:
        // Reload handled by caller re-fetching diff
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

    // Modal-specific handling
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
        return m, nil

    case "I":
        m.modal = modalAnnotations
        return m, nil

    case "ctrl+p":
        m.modal = modalFilePicker
        return m, nil

    // Scrolling
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
        if s.PendingKey == PendingKeyG {
            s.ScrollY = 0
            s.PendingKey = PendingKeyNone
        } else {
            s.PendingKey = PendingKeyG
        }
        return m, nil
    case "h":
        if s.Focus == FocusDiff {
            s.ScrollX = clamp(s.ScrollX-4, 0, 1000)
        } else {
            // sidebar: navigate up
            s.SidebarSelected = clamp(s.SidebarSelected-1, 0, len(s.Files)-1)
        }
    case "l":
        if s.Focus == FocusDiff {
            s.ScrollX += 4
        } else {
            // sidebar: navigate down
            s.SidebarSelected = clamp(s.SidebarSelected+1, 0, len(s.Files)-1)
        }

    // Hunk navigation
    case "{":
        jumpToHunk(s, -1)
    case "}":
        jumpToHunk(s, 1)

    // File navigation
    case "ctrl+j":
        s.NavigateToFile(s.CurrentFileIdx + 1)
    case "ctrl+k":
        s.NavigateToFile(s.CurrentFileIdx - 1)

    // Panel fullscreen
    case "[":
        s.Fullscreen = FullscreenOld
    case "]":
        s.Fullscreen = FullscreenNew
    case "=":
        s.Fullscreen = FullscreenOff

    // Focus
    case "1":
        s.Focus = FocusSidebar
    case "2":
        s.Focus = FocusDiff
    case "tab":
        s.SidebarCollapsed = !s.SidebarCollapsed

    // Actions
    case "e":
        return m, openEditor(s)
    case "y":
        return m, copySelection(s)
    case "i":
        m.modal = modalAnnotationEditor
        m.editInput = ""
        return m, nil
    case "space":
        return m.handleSpace()
    case "/":
        s.SearchActive = true
        s.SearchQuery = ""
        return m, nil
    case "n":
        advanceSearch(s, 1)
    case "N":
        advanceSearch(s, -1)

    // Stacked mode
    case "ctrl+l":
        if s.StackedMode {
            navigateStack(s, 1)
        }
    case "ctrl+h":
        if s.StackedMode {
            navigateStack(s, -1)
        }

    // Enter in sidebar
    case "enter":
        if s.Focus == FocusSidebar {
            s.NavigateToFile(s.SidebarSelected)
            s.Focus = FocusDiff
        }
    }

    s.PendingKey = PendingKeyNone
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
            s.Anchor = &CursorPos{Row: msg.Y, Col: msg.X}
            s.Head = &CursorPos{Row: msg.Y, Col: msg.X}
            s.ShowSelectionTooltip = false
            // Determine if click is on line number (line-mode) or content (char-mode)
            if msg.X < s.Layout.GutterWidth || (msg.X >= s.Layout.NewPanelStart && msg.X < s.Layout.NewPanelStart+s.Layout.GutterWidth) {
                s.SelectionMode = SelectionLine
            } else {
                s.SelectionMode = SelectionChar
            }
        case tea.MouseActionMotion:
            if s.Anchor != nil {
                s.Head = &CursorPos{Row: msg.Y, Col: msg.X}
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
    if s.Focus == FocusSidebar {
        // Toggle viewed for selected sidebar item; bulk if directory
        if s.SidebarSelected < len(s.Files) {
            path := s.Files[s.SidebarSelected].Path
            if _, ok := s.ViewedFiles[path]; ok {
                delete(s.ViewedFiles, path)
            } else {
                s.ViewedFiles[path] = struct{}{}
            }
        }
    } else {
        // Mark current file viewed + advance to next unviewed
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

    // Update layout for mouse coordinate translation
    s.Layout = PanelLayout{
        SidebarWidth:  sidebarWidth,
        GutterWidth:   5,
        OldPanelStart: sidebarWidth + 5,
        OldPanelEnd:   sidebarWidth + diffWidth/2,
        NewPanelStart: sidebarWidth + diffWidth/2 + 6,
        NewPanelEnd:   s.Width,
    }

    diffView := render.DiffView(s, m.theme, diffWidth, s.Height-2)
    footer := render.Footer(s, m.theme, s.Width, m.branchName)

    // File header
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

    // Overlay modals
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

// --- Helpers ---

func clamp(v, min, max int) int {
    if v < min { return min }
    if v > max { return max }
    return v
}

func jumpToHunk(s *AppState, dir int) {
    idx := s.CurrentHunkIdx()
    next := idx + dir
    if next >= 0 && next < len(s.Hunks) {
        s.ScrollY = s.Hunks[next].StartIdx
    }
}

func navigateStack(s *AppState, dir int) {
    next := s.CurrentCommitIdx + dir
    if next >= 0 && next < len(s.StackedCommits) {
        s.CurrentCommitIdx = next
        // caller should reload diff for the new commit
    }
}

func advanceSearch(s *AppState, dir int) {
    if len(s.SearchMatches) == 0 {
        return
    }
    s.SearchIdx = (s.SearchIdx + dir + len(s.SearchMatches)) % len(s.SearchMatches)
    match := s.SearchMatches[s.SearchIdx]
    s.ScrollY = match.LineIdx
}

func addAnnotation(s *AppState, content string) {
    if s.CurrentFileIdx >= len(s.Files) {
        return
    }
    ann := Annotation{
        ID:        fmt.Sprintf("%d", len(s.Annotations)+1),
        Filename:  s.Files[s.CurrentFileIdx].Path,
        Content:   content,
        CreatedAt: timeNow(),
    }
    if s.Anchor != nil {
        ann.Target = AnnotationTarget{
            Kind:      TargetLineRange,
            Panel:     s.SelectionPanel,
            StartLine: s.Anchor.Row + s.ScrollY,
            EndLine:   s.Head.Row + s.ScrollY,
        }
    } else {
        ann.Target = AnnotationTarget{Kind: TargetFile}
    }
    s.Annotations = append(s.Annotations, ann)
}

func openEditor(s *AppState) tea.Cmd {
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

func copySelection(s *AppState) tea.Cmd {
    // clipboard write is a side effect — wrap in Cmd
    return func() tea.Msg {
        if s.Anchor == nil || s.Head == nil {
            return nil
        }
        // Collect selected text from DiffLines
        var sb strings.Builder
        start := clamp(s.Anchor.Row+s.ScrollY, 0, len(s.DiffLines)-1)
        end := clamp(s.Head.Row+s.ScrollY, 0, len(s.DiffLines)-1)
        if start > end { start, end = end, start }
        for i := start; i <= end; i++ {
            dl := s.DiffLines[i]
            if s.SelectionPanel == PanelOld && dl.OldLine != nil {
                sb.WriteString(dl.OldLine.Text + "\n")
            } else if s.SelectionPanel == PanelNew && dl.NewLine != nil {
                sb.WriteString(dl.NewLine.Text + "\n")
            }
        }
        _ = clipboardWrite(sb.String())
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
    if startY < 0 { startY = 0 }
    if startX < 0 { startX = 0 }
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
        overlay := []rune(line)
        for j, ch := range overlay {
            if startX+j < len(r) {
                r[startX+j] = ch
            }
        }
        baseLines[lineIdx] = string(r)
    }
    return strings.Join(baseLines, "\n")
}

// Shims for time and clipboard (testable)
var (
    timeNow       = func() time.Time { return time.Now() }
    clipboardWrite = func(s string) error {
        return clipboard.WriteAll(s)
    }
)
```

> **Note:** Add these imports at the top of `app.go`:
```go
import (
    "fmt"
    "os"
    "os/exec"
    "strings"
    "time"

    "github.com/atotto/clipboard"
    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/lipgloss"
    "gitlens/internal/diff/render"
    "gitlens/internal/diff/theme"
)
```

- [ ] **Step 2: Verify build**

```bash
CGO_ENABLED=1 go build ./...
```

Fix any compilation errors (missing imports, type mismatches).

- [ ] **Step 3: Commit**

```bash
git add internal/diff/app.go
git commit -m "feat: add Bubble Tea app model with full keybinding set"
```

---

## Task 9: `diff` Command

**Files:**
- Create: `cmd/diff.go`

- [ ] **Step 1: Write `cmd/diff.go`**

```go
package cmd

import (
    "fmt"
    "os"
    "path/filepath"

    tea "github.com/charmbracelet/bubbletea"
    "github.com/spf13/cobra"
    diffpkg "gitlens/internal/diff"
    "gitlens/internal/diff/theme"
    "gitlens/internal/git_entity"
    "gitlens/internal/vcs"
)

var (
    diffWatch   bool
    diffTheme   string
    diffStacked bool
    diffFocus   string
    diffFiles   []string
)

var diffCmd = &cobra.Command{
    Use:   "diff [ref]",
    Short: "Interactive side-by-side diff viewer",
    Args:  cobra.MaximumNArgs(1),
    RunE: func(cmd *cobra.Command, args []string) error {
        backend, err := vcs.NewGitBackend(".")
        if err != nil {
            return err
        }

        // Load diff
        var gitDiff *git_entity.Diff
        if len(args) > 0 {
            ref := vcs.ParseRef(args[0])
            if ref.Single != "" {
                gitDiff, err = backend.GetRangeDiff(ref.Single+"^", ref.Single, false)
                if err != nil {
                    gitDiff, err = backend.GetRangeDiff("", ref.Single, false)
                }
            } else {
                gitDiff, err = backend.GetRangeDiff(ref.From, ref.To, ref.ThreeDot)
            }
        } else {
            gitDiff, err = backend.GetWorkingTreeDiff(false)
        }
        if err != nil {
            return err
        }
        if len(gitDiff.Files) == 0 {
            return fmt.Errorf("no changes to display")
        }

        // Filter files if --file flags provided
        files := gitDiff.Files
        if len(diffFiles) > 0 {
            files = filterFiles(files, diffFiles)
        }

        // Load theme
        themeName := Cfg.Theme.Base
        if diffTheme != "" {
            themeName = diffTheme
            Cfg.Theme.Base = themeName
        }
        th := theme.Load(Cfg)

        // Get branch name for footer
        branchName := getBranchName(backend)

        // Build app state
        state := diffpkg.NewAppState(files)

        // Focus on specific file if --focus provided
        if diffFocus != "" {
            for i, f := range files {
                if f.Path == diffFocus || filepath.Base(f.Path) == diffFocus {
                    state.NavigateToFile(i)
                    break
                }
            }
        }

        // Stacked mode
        if diffStacked && len(args) > 0 {
            ref := vcs.ParseRef(args[0])
            if ref.From != "" {
                commits, err := backend.GetCommitsInRange(ref.From, ref.To)
                if err == nil && len(commits) > 0 {
                    state.StackedMode = true
                    state.StackedCommits = commits
                }
            }
        }

        model := diffpkg.NewModel(state, th, branchName)

        p := tea.NewProgram(model,
            tea.WithAltScreen(),
            tea.WithMouseCellMotion(),
        )

        // Watch mode
        if diffWatch {
            watchPaths := watchPathsFor(args, backend)
            cancel, err := diffpkg.WatchFiles(p, watchPaths)
            if err != nil {
                fmt.Fprintf(os.Stderr, "watch: %v\n", err)
            } else {
                defer cancel()
            }
        }

        _, err = p.Run()
        return err
    },
}

func init() {
    rootCmd.AddCommand(diffCmd)
    diffCmd.Flags().BoolVar(&diffWatch, "watch", false, "auto-reload on file changes")
    diffCmd.Flags().StringVar(&diffTheme, "theme", "", "color theme override")
    diffCmd.Flags().BoolVar(&diffStacked, "stacked", false, "commit-by-commit navigation")
    diffCmd.Flags().StringVar(&diffFocus, "focus", "", "start at this file")
    diffCmd.Flags().StringArrayVar(&diffFiles, "file", nil, "filter to specific files")
}

func filterFiles(files []git_entity.FileDiff, filter []string) []git_entity.FileDiff {
    filterSet := make(map[string]bool)
    for _, f := range filter {
        filterSet[f] = true
    }
    var result []git_entity.FileDiff
    for _, f := range files {
        if filterSet[f.Path] || filterSet[filepath.Base(f.Path)] {
            result = append(result, f)
        }
    }
    return result
}

func getBranchName(backend *vcs.GitBackend) string {
    ref, err := backend.ResolveRef("HEAD")
    _ = ref
    _ = err
    // go-git: get symbolic ref
    repo := backend.Repo()
    head, err := repo.Head()
    if err != nil {
        return "HEAD"
    }
    if head.Name().IsBranch() {
        return head.Name().Short()
    }
    return head.Hash().String()[:7]
}

func watchPathsFor(args []string, backend vcs.Backend) []string {
    if len(args) == 0 {
        // working tree diff: watch .git/index
        return []string{".git/index"}
    }
    // commit-based: watch .git/refs and .git/HEAD
    return []string{".git/refs", ".git/HEAD"}
}
```

> **Note:** `getBranchName` calls `backend.Repo()` — expose the underlying `*gogit.Repository` from `GitBackend` by adding:
```go
func (g *GitBackend) Repo() *gogit.Repository { return g.repo }
```
to `internal/vcs/git.go`.

- [ ] **Step 2: Add `Repo()` method to GitBackend**

In `internal/vcs/git.go`, add:
```go
func (g *GitBackend) Repo() *gogit.Repository { return g.repo }
```

Also update `cmd/diff.go`'s `getBranchName` signature to accept `*vcs.GitBackend` instead of `vcs.Backend`.

- [ ] **Step 3: Verify full build**

```bash
CGO_ENABLED=1 go build -o gitlens .
```

- [ ] **Step 4: Smoke-test the diff viewer**

```bash
# In any git repo with staged or unstaged changes:
./gitlens diff
./gitlens diff HEAD
./gitlens diff --watch
./gitlens diff --theme dracula
```

Expected: TUI launches, keybindings work, `q` quits cleanly.

- [ ] **Step 5: Commit**

```bash
git add cmd/diff.go internal/vcs/git.go
git commit -m "feat: wire up diff command to full TUI app"
```

---

## Task 10: Final Verification

- [ ] **Step 1: Run all tests**

```bash
CGO_ENABLED=1 go test ./... -v
```

Expected: all tests PASS.

- [ ] **Step 2: Build and run help**

```bash
CGO_ENABLED=1 go build -o gitlens .
./gitlens diff --help
```

- [ ] **Step 3: Test diff in this repo**

```bash
./gitlens diff HEAD
# Navigate: j/k scroll, tab sidebar, ? help, q quit
```

- [ ] **Step 4: Tag milestone commit**

```bash
git add .
git commit -m "feat: complete diff TUI implementation"
git tag v0.1.0
```
