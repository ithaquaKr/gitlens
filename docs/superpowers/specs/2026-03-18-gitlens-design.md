# GitLens — Go Rewrite Design Spec

**Date:** 2026-03-18
**Source:** Rewrite of [lumen](https://github.com/jnsahaj/lumen) (Rust) in Go
**TUI Framework:** [Bubble Tea](https://github.com/charmbracelet/bubbletea)

---

## Overview

GitLens is an AI-powered Git CLI tool — a Go rewrite of lumen — with full feature parity for local Git workflows. It provides AI-assisted commit drafting, change explanation, natural language git command generation, and a rich terminal-based side-by-side diff viewer.

**Out of scope (v1):**
- Jujutsu (jj) VCS support
- GitHub/GitLab integration (PR mode, remote mark-viewed sync)
- More than 2 AI providers (Claude, Gemini)

---

## Project Structure

```
gitlens/
├── main.go
├── go.mod
├── cmd/
│   ├── root.go          # Cobra root command + global flags
│   ├── draft.go         # Generate AI commit messages
│   ├── explain.go       # Explain git changes
│   ├── operate.go       # Natural language → git commands
│   ├── diff.go          # Interactive TUI diff viewer
│   └── configure.go     # Interactive setup wizard
├── internal/
│   ├── config/          # TOML config loading + precedence chain
│   ├── vcs/             # VcsBackend interface + GitBackend impl
│   ├── ai/              # Provider interface + Claude, Gemini impls
│   ├── git_entity/      # Commit, Diff data models
│   └── diff/            # Full TUI app (Bubble Tea)
│       ├── app.go       # Bubble Tea Model: Init/Update/View
│       ├── state.go     # AppState struct
│       ├── diff_algo.go # Side-by-side diff computation
│       ├── context.go   # Sticky context lines (tree-sitter)
│       ├── highlight/   # Syntax highlighting (tree-sitter)
│       ├── theme/       # Preset themes + override merge
│       └── render/
│           ├── diffview.go
│           ├── sidebar.go
│           ├── footer.go
│           └── modal.go
└── docs/
    └── superpowers/specs/
```

---

## CLI

**Binary name:** `gitlens`

**Global flags:**
| Flag | Description |
|------|-------------|
| `--provider` | AI provider: `claude`, `gemini` |
| `--api-key` | API key override |
| `--model` | Model name override |
| `--config` | Path to config file |
| `--theme` | Color theme override |

**Commands:**

| Command | Description |
|---------|-------------|
| `gitlens draft [--context TEXT]` | Generate conventional commit message from staged diff |
| `gitlens explain [REF\|-] [--staged] [--query TEXT] [--list]` | Explain changes with AI; `-` reads SHA from stdin; `--list` opens interactive commit picker |
| `gitlens operate QUERY` | Generate git command from natural language |
| `gitlens diff [REF] [--file PATH]... [--watch] [--theme NAME] [--stacked] [--focus PATH]` | Interactive TUI diff viewer |
| `gitlens configure` | Interactive setup wizard |

**Commit reference formats (explain, diff):**
- `HEAD`, `abc123` — single commit
- `main..feature` — range
- `main...feature` — three-dot (merge-base)
- `-` (explain only) — read SHA from stdin

### `explain --list` (Interactive Commit Picker)

When `--list` is passed, gitlens presents an interactive fuzzy commit selector:
- Shell out to `fzf` if available, piping `git log --oneline` output. Capture selected SHA.
- **Fallback:** If `fzf` is not installed, print: `"explain --list requires fzf. Install it from https://github.com/junegunn/fzf"`

### `explain` AI Output

AI responses are in markdown. Pipe output through `mdcat` if available on `$PATH`; fall back to plain stdout. This matches lumen's behavior.

---

## Configuration System

**Precedence (highest → lowest):**
1. CLI flags (`--provider`, `--api-key`, `--model`, `--theme`)
2. `--config <path>` explicit file
3. Environment variables: `GITLENS_PROVIDER`, `GITLENS_API_KEY`, `GITLENS_MODEL`, `GITLENS_THEME`
4. `./gitlens.config.toml` (project root)
5. `~/.config/gitlens/config.toml` (global)
6. Hardcoded defaults

> **Note:** Env vars sit above project-level config so that a `GITLENS_API_KEY` in the shell always overrides any API key committed to a project config file.

**Full config schema (`config.toml`):**
```toml
provider = "claude"          # claude | gemini
api_key  = "sk-ant-..."
model    = "claude-opus-4-6"

[theme]
base = "catppuccin-mocha"    # preset base theme

  [theme.override]
  # Syntax colors (hex strings, 16 fields total)
  keyword          = "#cba6f7"
  string           = "#a6e3a1"
  comment          = "#585b70"
  function         = "#89b4fa"
  function_macro   = "#f38ba8"
  type             = "#f38ba8"
  number           = "#fab387"
  operator         = "#cdd6f4"
  variable         = "#cdd6f4"
  variable_builtin = "#f38ba8"
  variable_member  = "#cdd6f4"
  module           = "#89b4fa"
  tag              = "#f38ba8"
  attribute        = "#fab387"
  label            = "#f38ba8"
  punctuation      = "#cdd6f4"

  # Diff colors
  added_bg        = "#1e3a2a"
  deleted_bg      = "#3a1e1e"
  added_word_bg   = "#2d5a3d"
  deleted_word_bg = "#5a2d2d"

  # UI colors
  border    = "#313244"
  selection = "#45475a"

[draft]
  [draft.commit_types]
  feat     = "A new feature"
  fix      = "A bug fix"
  docs     = "Documentation changes"
  refactor = "Code refactoring"
  test     = "Adding tests"
  chore    = "Maintenance tasks"
```

---

## Theme System

**Preset bases (11 themes):**
`dark`, `light`, `catppuccin-mocha`, `catppuccin-latte`, `dracula`, `nord`, `gruvbox-dark`, `gruvbox-light`, `one-dark`, `solarized-dark`, `solarized-light`

**How it works:**
1. Load preset theme into `Theme` struct
2. Merge `[theme.override]` fields on top (only specified fields are overridden)
3. `--theme <name>` CLI flag overrides the base preset; overrides still apply

**`SyntaxColors` struct (16 fields, matching lumen's tree-sitter highlight names):**
`Keyword`, `String`, `Comment`, `Function`, `FunctionMacro`, `Type`, `Number`, `Operator`, `Variable`, `VariableBuiltin`, `VariableMember`, `Module`, `Tag`, `Attribute`, `Label`, `Punctuation`

**`DiffColors` struct:**
`AddedBg`, `DeletedBg`, `AddedWordBg`, `DeletedWordBg`, `AddedGutterBg`, `DeletedGutterBg`

**`UiColors` struct:**
`Border`, `Text`, `Selection`, `SearchHighlight`, `StatusAdded`, `StatusModified`, `StatusDeleted`

---

## AI Provider Layer

```go
// internal/ai/provider.go
type Provider interface {
    Complete(ctx context.Context, prompt string) (string, error)
    Stream(ctx context.Context, prompt string) (<-chan StreamChunk, error)
    Name() string
}

// StreamChunk carries either a token or a terminal error
type StreamChunk struct {
    Text string
    Err  error  // non-nil on stream error; signals channel close
}

type ProviderFactory func(apiKey, model string) (Provider, error)
```

> **Note:** Lumen uses non-streaming `Complete()` (with a spinner) for all commands. `Stream()` is a gitlens enhancement for real-time output on `explain` and `draft`. `Complete()` is the required interface; `Stream()` is optional and may be added per-provider incrementally.

**`operate` response format:** Plain multi-line text — one line for the command, one line for the explanation, and optionally one line prefixed `WARNING:` for destructive operations. gitlens parses this line-by-line (no XML).

**`operate` confirmation flow:**
1. Display the explanation line to the user
2. If a `WARNING:` line is present, print it prominently
3. Prompt `[y/N]:` — read one keystroke from stdin
4. If `y`: shell-exec the command via `os/exec`, streaming output to stdout
5. If anything else: abort with `"Aborted."`

**Registry pattern for extensibility:**
```go
// internal/ai/registry.go
var registry = map[string]ProviderFactory{}

func Register(name string, factory ProviderFactory)
func New(cfg *config.Config) (Provider, error)
```

**Implementations (v1):**
- `internal/ai/claude.go` — `github.com/anthropics/anthropic-sdk-go`
- `internal/ai/gemini.go` — `github.com/google/generative-ai-go/genai`

**Adding a new provider:** implement `Provider` interface, call `Register()` in `init()`.

**Prompts** (`internal/ai/prompts.go`):
- `ExplainPrompt(diff, query string) string`
- `DraftPrompt(diff, context string, commitTypes map[string]string) string`
- `OperatePrompt(query string) string`

---

## VCS Layer

```go
// internal/vcs/backend.go
type Backend interface {
    GetCommit(ref string) (*git_entity.Commit, error)
    GetWorkingTreeDiff(staged bool) (*git_entity.Diff, error)
    GetRangeDiff(from, to string, threeDot bool) (*git_entity.Diff, error)
    GetCommitsInRange(from, to string) ([]*git_entity.Commit, error)
    GetFileContentAtRef(path, ref string) (string, error)
    ResolveRef(ref string) (string, error)
}
```

**Implementation:** `internal/vcs/git.go` — `GitBackend` using `github.com/go-git/go-git/v5`

**Data models (`internal/git_entity/`):**

The data layer has two distinct model levels:

**Git layer models** (returned by VCS backend):
```go
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

// FileDiff holds full file content for both sides (needed by TUI renderer)
type FileDiff struct {
    Path       string   // current path
    Status     string   // "A" added, "M" modified, "D" deleted
    OldContent string   // full content at old ref
    NewContent string   // full content at new ref
    IsBinary   bool
}
```

**Rendering layer models** (computed by `diff_algo`, consumed by TUI):
```go
type ChangeType int
const (
    Equal ChangeType = iota
    Delete
    Insert
    Modified
)

type Segment struct {
    Text      string
    Highlight bool  // true = word-level diff highlight
}

type LineContent struct {
    LineNo int
    Text   string
}

type DiffLine struct {
    OldLine     *LineContent
    NewLine     *LineContent
    ChangeType  ChangeType
    OldSegments []Segment   // word-level diff (Modified lines only)
    NewSegments []Segment
}

// Hunk marks a contiguous block of non-Equal lines (used for {/} navigation)
type Hunk struct {
    StartIdx int   // index into []DiffLine
    EndIdx   int
}
```

---

## Diff TUI (Bubble Tea)

### Architecture Note

The TUI is built with Bubble Tea's Elm-architecture (`Init` / `Update` / `View`). All state mutations happen in `Update()` in response to `tea.Msg` values. Side effects (file watching, AI streaming, clipboard writes, editor launch) are dispatched as `tea.Cmd`. Modal overlays (annotation editor, help, file picker) are nested sub-models whose messages bubble up through the parent `Update()`.

### Mouse Event Handling

Mouse support enabled at startup with `tea.WithMouseCellMotion()`. Mouse messages arrive as `tea.MouseMsg` with `X`, `Y`, `Type`, and `Button` fields.

**Selection via mouse:**
- **Click-drag in content area:** character-level selection. On `MousePress`, record anchor `(x, y)`. On `MouseMotion`, update head. On `MouseRelease`, finalize and show selection tooltip (`i annotate / y copy / esc`).
- **Click on line number gutter:** line-level selection mode. Dragging selects whole lines.
- **Panel determination:** translate `(x, y)` to panel (old/new) using `AppState.Layout`.

```go
type PanelLayout struct {
    SidebarWidth  int
    OldPanelStart int
    OldPanelEnd   int
    NewPanelStart int
    NewPanelEnd   int
    GutterWidth   int
}
```

`PanelLayout` is recomputed each `View()` call and stored in `AppState` for use by `Update()` mouse handlers.

### Model Structure

```go
// internal/diff/app.go
type Model struct {
    state    *AppState
    sidebar  sidebar.Model
    diffview diffview.Model
    footer   footer.Model
    modal    modal.Model
}

func (m Model) Init() tea.Cmd
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd)
func (m Model) View() string
```

### AppState (`internal/diff/state.go`)

```go
type PendingKey int
const (
    PendingKeyNone PendingKey = iota
    PendingKeyG    // waiting for second 'g' (go-to-top)
)

type DiffFullscreen int
const (
    FullscreenOff DiffFullscreen = iota
    FullscreenOld
    FullscreenNew
)

type AppState struct {
    // File list
    Files            []git_entity.FileDiff
    CurrentFileIdx   int

    // Scroll
    ScrollY          int
    ScrollX          int
    SidebarScrollX   int

    // Navigation
    Hunks            []Hunk
    ContextLines     []ContextLine   // sticky context header (see below)
    PendingKey       PendingKey      // for double-key sequences (gg, G)
    DiffFullscreen   DiffFullscreen  // [/] panel fullscreen mode

    // Sidebar
    SidebarCollapsed bool
    SidebarSelected  int
    CollapsedDirs    map[string]bool

    // Selection
    Anchor              *CursorPos
    Head                *CursorPos
    SelectionMode       SelectionMode  // Char or Line
    Panel               PanelFocus    // Old or New
    ShowSelectionTooltip bool

    // Annotations
    Annotations []Annotation

    // Search
    SearchQuery   string
    SearchMatches []MatchPos
    SearchIdx     int

    // Viewed files
    ViewedFiles map[string]struct{}  // keyed by filename

    // Stacked mode
    StackedMode         bool
    StackedCommits      []*git_entity.Commit
    CurrentCommitIdx    int
    StackedViewedFiles  map[string]map[string]struct{}  // SHA → filename → viewed

    // Watch mode
    WatchReload chan struct{}

    // Layout (recomputed each View)
    Layout PanelLayout
}
```

### Sticky Context Lines (`internal/diff/context.go`)

When scrolling past a function/class definition, up to 5 enclosing scope headers float at the top of the diff view.

```go
type ContextLine struct {
    LineNumber int
    Content    string
}
```

- Uses `github.com/smacker/go-tree-sitter` with per-language CGo grammar packages
- **Build requirement:** CGo must be enabled (`CGO_ENABLED=1`) and a C compiler present
- Per-language packages required (each compiles a C grammar): `github.com/smacker/go-tree-sitter/golang`, `/rust`, `/typescript`, `/javascript`, `/python`, `/c_sharp`; TSX/JSX use the TypeScript/JavaScript grammars with dialect flags
- Supported languages: Go, Rust, TypeScript, TSX, JavaScript, JSX, Python, C#
- Fallback: no sticky header for unsupported languages
- `AppState.ContextLines []ContextLine` is updated on each scroll event
- Rendered above the diff area in `diffview.go`

### Syntax Highlighting (`internal/diff/highlight/`)

- Uses `github.com/smacker/go-tree-sitter` (same library and CGo grammars as context lines)
- Maps tree-sitter highlight names (e.g. `function.method`, `variable.member`) to `SyntaxColors` fields — same mapping as lumen
- Language detection from file extension
- Fallback to plain text for unsupported languages

### Diff Algorithm (`internal/diff/diff_algo.go`)

- Uses `github.com/sergi/go-diff` for sequence diffing
- Input: `OldContent`, `NewContent` strings from `FileDiff`
- Output: `[]DiffLine` for side-by-side rendering, `[]Hunk` for navigation
- Word-level highlights on `Modified` lines only when >20% of content is unchanged

### Rendering (`internal/diff/render/`)

| File | Responsibility |
|------|---------------|
| `diffview.go` | Sticky context lines header, side-by-side panels, line numbers, gutters (A/M/D), word-level highlights, selection highlight, search highlights |
| `sidebar.go` | Collapsible file tree, directory collapse toggle, status colors, viewed-file dimming, scroll |
| `footer.go` | Branch name badge, added/removed stats, keybindings hint, scroll position, stacked commit badge `[2/5]` |
| `modal.go` | Help keybindings, file picker, annotations list |

### Keybindings (full set, matching lumen)

**Scrolling:**
| Key | Action |
|-----|--------|
| `j` / `k` | Scroll down / up (1 line) |
| `ctrl+d` / `ctrl+u` | Half-page scroll down / up |
| `PageDown` / `PageUp` | Full-page scroll |
| `g` `g` | Scroll to top |
| `G` | Scroll to bottom |
| `h` / `l` | Scroll left / right (horizontal) |

**Navigation:**
| Key | Action |
|-----|--------|
| `{` / `}` | Jump to prev / next hunk |
| `ctrl+j` / `ctrl+k` | Jump to next / prev file |
| `ctrl+p` | Open file picker modal |
| `1` | Focus sidebar |
| `2` | Focus diff view |
| `enter` | Select file / toggle directory (when sidebar focused) |
| `tab` | Toggle sidebar open/closed |
| Arrow keys | Scroll / sidebar navigation |

**Panels:**
| Key | Action |
|-----|--------|
| `[` | Fullscreen old (left) panel |
| `]` | Fullscreen new (right) panel |
| `=` | Reset to side-by-side |

**Actions:**
| Key | Action |
|-----|--------|
| `e` | Open file in `$EDITOR` |
| `y` | Copy selection to clipboard |
| `i` | Annotate selection / hunk / file |
| `I` | View all annotations |
| `space` | Toggle current file as viewed (context-sensitive, see below) |
| `/` | Enter search mode |
| `n` / `N` | Next / prev search match |
| `?` | Toggle help modal |
| `q` | Quit |

**Stacked mode only:**
| Key | Action |
|-----|--------|
| `ctrl+l` | Next commit |
| `ctrl+h` | Previous commit |

### `space` Key — Context-Sensitive Behavior

- **Sidebar focused:** Toggle viewed on the selected sidebar item. If it is a directory, bulk-toggle all child files.
- **Diff view focused:** Mark current file as viewed AND auto-advance to the next unviewed file (wrapping around). This is the primary code-review workflow.

`ViewedFiles` is keyed by filename string. `StackedViewedFiles` is keyed by commit SHA (outer) → filename string (inner). Both use string keys (not indices) so state survives reloads and diff recalculations.

### Stacked Mode (`--stacked`)

- Requires a range ref (e.g. `gitlens diff main..feature --stacked`)
- `GetCommitsInRange()` loads all commits; each is displayed individually
- `ctrl+l` / `ctrl+h` navigate the stack
- `StackedViewedFiles` is keyed by commit SHA → filename set
- Footer shows `[2/5]` badge and short SHA

### Watch Mode

`--watch` spawns a goroutine using `github.com/fsnotify/fsnotify`, debounced ~200ms. On change, sends a `ReloadMsg` to the Bubble Tea program via `p.Send()`. The `Update()` handler reloads the diff.

**What to watch (depends on diff source):**
- **Working-tree diff** (`gitlens diff` with no ref, or `--staged`): watch individual changed files in the working directory plus `.git/index` for staged changes
- **Commit-based diff** (`gitlens diff HEAD`, `gitlens diff main..feature`): watch `.git/refs/` and `.git/HEAD` so the view refreshes when the ref moves (e.g. after a new commit or rebase)

### Annotations

```go
type AnnotationTarget struct {
    Kind      TargetKind  // TargetFile or TargetLineRange
    Panel     PanelFocus  // Old or New (line range only)
    StartLine int         // line range only
    EndLine   int         // line range only
}

type TargetKind int
const (
    TargetFile TargetKind = iota
    TargetLineRange
)

type Annotation struct {
    ID        string
    Filename  string
    Target    AnnotationTarget
    Content   string
    CreatedAt time.Time
}
```

- In-memory only (`[]Annotation` in `AppState`)
- Modal actions: view, edit, delete, copy to clipboard
- Jump-to-annotation scrolls to the annotated panel and line

### `configure` Command

Interactive wizard using `github.com/charmbracelet/huh`:
1. Provider select: `claude`, `gemini`
2. API key text input (hints if env var already set)
3. Model name input (provider default pre-filled)
4. Theme select (11 presets)
5. Writes `~/.config/gitlens/config.toml` (creates dirs as needed)

---

## Key Go Dependencies

| Package | Purpose |
|---------|---------|
| `github.com/spf13/cobra` | CLI framework |
| `github.com/charmbracelet/bubbletea` | TUI event loop |
| `github.com/charmbracelet/lipgloss` | TUI styling/layout |
| `github.com/charmbracelet/bubbles` | TUI components (viewport, textinput, list) |
| `github.com/charmbracelet/huh` | Interactive forms (`configure` wizard) |
| `github.com/go-git/go-git/v5` | Git operations |
| `github.com/smacker/go-tree-sitter` | Syntax highlighting + sticky context lines |
| `github.com/BurntSushi/toml` | TOML config parsing |
| `github.com/sergi/go-diff` | Diff algorithm |
| `github.com/anthropics/anthropic-sdk-go` | Claude AI |
| `github.com/google/generative-ai-go/genai` | Gemini AI |
| `github.com/fsnotify/fsnotify` | File watching (`--watch` mode) |
| `github.com/atotto/clipboard` | Clipboard (copy selection) |

---

## Error Handling

- All internal functions return `(T, error)` — no panics
- CLI layer wraps errors with `cobra.CheckErr()` for clean user messages
- AI streaming errors sent as `StreamChunk{Err: err}` then channel closed
- Git errors wrapped with context (e.g. `"resolving ref %q: %w"`)

---

## Testing Strategy

- `internal/vcs/` — integration tests against real temp git repos (`os.MkdirTemp` + `go-git` init)
- `internal/ai/` — unit tests with mock `Provider` implementations
- `internal/diff/diff_algo.go` — unit tests for diff computation correctness
- `internal/config/` — unit tests for precedence chain logic
- TUI rendering — Bubble Tea model logic tested via `Update()` message passing; no visual render tests
