# GitLens — Go Rewrite Design Spec

**Date:** 2026-03-18
**Source:** Rewrite of [lumen](https://github.com/jnsahaj/lumen) (Rust) in Go
**TUI Framework:** [Bubble Tea](https://github.com/charmbracelet/bubbletea)

---

## Overview

GitLens is an AI-powered Git CLI tool — a Go rewrite of lumen — with full feature parity for local Git workflows. It provides AI-assisted commit drafting, change explanation, natural language git command generation, and a rich terminal-based side-by-side diff viewer.

**Out of scope (v1):**
- Jujutsu (jj) VCS support
- GitHub/GitLab integration (PR mode, mark-viewed sync)
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
│       ├── app.go
│       ├── state.go
│       ├── diff_algo.go
│       ├── highlight/
│       ├── theme/
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
| `gitlens explain [REF] [--staged] [--query TEXT]` | Explain changes with AI |
| `gitlens operate QUERY` | Generate git command from natural language |
| `gitlens diff [REF] [--file PATH]... [--watch] [--theme NAME] [--focus PATH]` | Interactive TUI diff viewer |
| `gitlens configure` | Interactive setup wizard |

**Commit reference formats (explain, diff):**
- `HEAD`, `abc123` — single commit
- `main..feature` — range
- `main...feature` — three-dot (merge-base)

---

## Configuration System

**Precedence (highest → lowest):**
1. CLI flags (`--provider`, `--api-key`, `--model`, `--theme`)
2. `--config <path>` explicit file
3. `./gitlens.config.toml` (project root)
4. `~/.config/gitlens/config.toml` (global)
5. Environment variables: `GITLENS_PROVIDER`, `GITLENS_API_KEY`, `GITLENS_MODEL`, `GITLENS_THEME`

**Full config schema (`config.toml`):**
```toml
provider = "claude"          # claude | gemini
api_key  = "sk-ant-..."
model    = "claude-opus-4-6"

[theme]
base = "catppuccin-mocha"    # preset base theme

  [theme.override]
  # Syntax colors (hex strings)
  keyword  = "#cba6f7"
  string   = "#a6e3a1"
  comment  = "#585b70"
  function = "#89b4fa"
  type     = "#f38ba8"
  number   = "#fab387"

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

**Theme struct covers:**
- Syntax colors: `Keyword`, `String`, `Comment`, `Function`, `Type`, `Number`, `Operator`, `Variable`
- Diff colors: `AddedBg`, `DeletedBg`, `AddedWordBg`, `DeletedWordBg`, `AddedGutterBg`, `DeletedGutterBg`
- UI colors: `Border`, `Text`, `Selection`, `SearchHighlight`, `StatusAdded`, `StatusModified`, `StatusDeleted`

This allows full colorscheme sync with terminal/neovim palettes via manual hex configuration.

---

## AI Provider Layer

```go
// internal/ai/provider.go
type Provider interface {
    Complete(ctx context.Context, prompt string) (string, error)
    Stream(ctx context.Context, prompt string) (<-chan string, error)
    Name() string
}

type ProviderFactory func(apiKey, model string) (Provider, error)
```

**Registry pattern for extensibility:**
```go
// internal/ai/registry.go
var registry = map[string]ProviderFactory{}

func Register(name string, factory ProviderFactory)
func New(cfg *config.Config) (Provider, error)
```

**Implementations (v1):**
- `internal/ai/claude.go` — Anthropic SDK (`github.com/anthropics/anthropic-sdk-go`)
- `internal/ai/gemini.go` — Google GenAI SDK (`google.golang.org/genai`)

**Adding a new provider:** implement `Provider` interface, call `Register()` in `init()`.

**Prompts** (`internal/ai/prompts.go`):
- `ExplainPrompt(diff, query string) string`
- `DraftPrompt(diff, context string, commitTypes map[string]string) string`
- `OperatePrompt(query string) string`

**Streaming:** used for `explain` and `draft` (real-time output). `operate` uses non-streaming (needs full XML response to parse).

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
```go
type Commit struct {
    Hash    string
    Message string
    Diff    string
    Author  string
    Email   string
    Date    time.Time
}

type FileDiff struct {
    Path      string
    OldPath   string
    Status    string  // A, M, D
    Hunks     []Hunk
}

type Diff struct {
    Raw   string
    Files []FileDiff
}
```

---

## Diff TUI (Bubble Tea)

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

- Current file index, file list
- Scroll position (vertical offset)
- Hunk positions for jump navigation
- Annotations (in-memory, `[]Annotation`)
- Selection: anchor, head, mode (char/line), panel (old/new)
- Sidebar: collapsed bool, scroll offset
- Search: query string, match positions, current match index
- Watch mode: reload trigger channel

### Diff Algorithm (`internal/diff/diff_algo.go`)

- Uses `github.com/sergi/go-diff` for sequence diffing
- Produces `[]DiffLine` for side-by-side rendering:
  ```go
  type DiffLine struct {
      OldLine     *LineContent
      NewLine     *LineContent
      ChangeType  ChangeType  // Equal, Delete, Insert, Modified
      OldSegments []Segment   // word-level diff (Modified only)
      NewSegments []Segment
  }
  ```
- Word-level highlights on `Modified` lines only when >20% of content unchanged (same threshold as lumen)

### Syntax Highlighting (`internal/diff/highlight/`)

- Uses `github.com/alecthomas/chroma/v2`
- Language detection from file extension
- Applies theme syntax colors to highlighted tokens
- Fallback to plain text for unsupported languages

### Rendering (`internal/diff/render/`)

| File | Responsibility |
|------|---------------|
| `diffview.go` | Side-by-side panels, line numbers, gutters (A/M/D badge), word-level highlights, selection highlight |
| `sidebar.go` | Collapsible file tree, single-child chain collapse, status colors, scroll |
| `footer.go` | Branch name badge, added/removed line stats, keybindings hint, scroll position |
| `modal.go` | Help keybindings modal, file picker modal, annotations list modal |

### Keybindings (identical to lumen)

| Key | Action |
|-----|--------|
| `j` / `k` | Scroll down / up |
| `{` / `}` | Jump to prev / next hunk |
| `tab` | Toggle sidebar |
| `e` | Open file in `$EDITOR` |
| `y` | Copy selection to clipboard |
| `i` | Annotate selection/hunk/file |
| `I` | View all annotations |
| `?` | Toggle help modal |
| `q` | Quit |
| `/` | Search |
| `n` / `N` | Next / prev search match |
| `space` | (reserved for future viewed tracking) |

### Watch Mode

`--watch` flag spawns a goroutine using `github.com/fsnotify/fsnotify` that sends a `ReloadMsg` to the Bubble Tea program on file changes. The model reloads the diff on receipt.

### Annotations

- In-memory only (`[]Annotation` in `AppState`)
- Fields: `ID`, `Filename`, `Target` (file/line range), `Content`, `CreatedAt`
- Actions: create, view, edit, delete, copy to clipboard
- Displayed in annotations modal

---

## Key Go Dependencies

| Package | Purpose |
|---------|---------|
| `github.com/spf13/cobra` | CLI framework |
| `github.com/charmbracelet/bubbletea` | TUI event loop |
| `github.com/charmbracelet/lipgloss` | TUI styling/layout |
| `github.com/charmbracelet/bubbles` | TUI components (viewport, textinput, list) |
| `github.com/go-git/go-git/v5` | Git operations |
| `github.com/alecthomas/chroma/v2` | Syntax highlighting |
| `github.com/BurntSushi/toml` | TOML config parsing |
| `github.com/sergi/go-diff` | Diff algorithm |
| `github.com/anthropics/anthropic-sdk-go` | Claude AI |
| `google.golang.org/genai` | Gemini AI |
| `github.com/fsnotify/fsnotify` | File watching (--watch mode) |
| `github.com/atotto/clipboard` | Clipboard (copy selection) |

---

## Error Handling

- All internal functions return `(T, error)` — no panics
- CLI layer wraps errors with `cobra.CheckErr()` for clean user messages
- AI streaming errors sent through error channel
- Git errors wrapped with context (e.g. `"resolving ref %q: %w"`)

---

## Testing Strategy

- `internal/vcs/` — integration tests against real temp git repos (using `os.MkdirTemp`)
- `internal/ai/` — unit tests with mock `Provider` implementations
- `internal/diff/diff_algo.go` — unit tests for diff computation correctness
- `internal/config/` — unit tests for precedence chain logic
- TUI rendering — not unit tested (Bubble Tea model logic tested via `Update()` message passing)
