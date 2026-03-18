# gitlens

> AI-powered git CLI — interactive diff viewer, commit drafting, and natural-language git operations.

[![CI](https://github.com/ithaquaKr/gitlens/actions/workflows/ci.yml/badge.svg)](https://github.com/ithaquaKr/gitlens/actions/workflows/ci.yml)
[![Go Version](https://img.shields.io/badge/go-1.24%2B-00ADD8?logo=go)](go.mod)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

---

```
┌─ src/api/handler.go [M] ────────────────────────────────────────────────────────────┐
│ main.go          [M] │  1    │ func HandleRequest(w http.ResponseWriter,            │
│ src/api/         [M] │- 2    │     r *http.Request) {                              │
│   handler.go     [M] │+ 2    │     r *http.Request, db *DB) {                      │
│   middleware.go  [A] │  3    │   log.Println("request:", r.URL.Path)               │
│ internal/db/     [M] │- 4    │   if err := validate(r); err != nil {               │
│   client.go      [M] │+ 4    │   if err := validate(r, db); err != nil {           │
│                      │  5    │     http.Error(w, err.Error(), 400)                 │
├──────────────────────┴───────┴─────────────────────────────────────────────────────┤
│  main  ●3 files  } hunks  tab sidebar  space viewed  / search  ? help  q quit      │
└─────────────────────────────────────────────────────────────────────────────────────┘
```

## Features

| Command | What it does |
|---------|-------------|
| `diff` | Interactive side-by-side diff viewer with syntax highlighting, search, and annotations |
| `draft` | AI-generated conventional commit messages from your staged changes |
| `explain` | Plain-English explanation of any diff, commit, or range |
| `operate` | Run git commands from natural language with confirmation |
| `configure` | Interactive setup wizard for API keys and preferences |

**Diff viewer highlights:**
- Side-by-side panels with word-level change highlights
- Syntax highlighting for 30+ languages (Go, Rust, TypeScript, Python, C#, and more)
- File sidebar with status badges and viewed tracking
- Hunk navigation, fullscreen panels, in-line annotations
- Search, mouse support, clipboard copy
- `--watch` mode — auto-reloads on file changes
- 11 built-in color themes

---

## Requirements

- **Go 1.24+** (for building from source)
- **A C compiler** (gcc or clang) for syntax highlighting — optional, the tool works without it
- An API key for [Anthropic Claude](https://console.anthropic.com/) or [Google Gemini](https://aistudio.google.com/)

> No system `git` required — uses pure-Go git via [go-git](https://github.com/go-git/go-git).

---

## Installation

### Pre-built binaries

Download the latest release for your platform from the [Releases page](https://github.com/ithaquaKr/gitlens/releases).

```bash
# macOS / Linux (one-liner)
curl -fsSL https://raw.githubusercontent.com/ithaquaKr/gitlens/main/scripts/install.sh | sh
```

### Homebrew (macOS / Linux)

```bash
brew install ithaquaKr/tap/gitlens
```

### From source

```bash
# With syntax highlighting (requires C compiler)
CGO_ENABLED=1 go install github.com/ithaquaKr/gitlens@latest

# Without syntax highlighting (pure Go, no C compiler needed)
CGO_ENABLED=0 go install github.com/ithaquaKr/gitlens@latest
```

### Build locally

```bash
git clone https://github.com/ithaquaKr/gitlens
cd gitlens
make build          # CGO_ENABLED=1 (recommended)
make build-nocgo    # CGO_ENABLED=0 (no C compiler required)
```

---

## Quick Start

```bash
# 1. Configure your AI provider (writes ~/.config/gitlens/config.toml)
gitlens configure

# 2. View current working tree diff
gitlens diff

# 3. Draft a commit message for staged changes
git add -p
gitlens draft

# 4. Explain the last commit
gitlens explain HEAD

# 5. Run git commands with natural language
gitlens operate "undo my last commit but keep the changes"
```

---

## Commands

### `gitlens diff [ref]`

Launch the interactive side-by-side diff viewer.

```bash
gitlens diff                     # Working tree vs HEAD
gitlens diff HEAD                # Last commit
gitlens diff main..feature       # Branch range
gitlens diff abc123              # Specific commit
gitlens diff --watch             # Auto-reload on file changes
gitlens diff --theme dracula     # Use a specific theme
gitlens diff --focus handler.go  # Open directly to a file
gitlens diff --file api.go --file db.go  # Filter to specific files
gitlens diff --stacked main..feature     # Step through commits one-by-one
```

**Keyboard shortcuts:**

| Key | Action |
|-----|--------|
| `j` / `↓` | Scroll down |
| `k` / `↑` | Scroll up |
| `Ctrl+D` / `Ctrl+U` | Half page down / up |
| `PgDn` / `PgUp` | Full page down / up |
| `gg` | Jump to top |
| `G` | Jump to bottom |
| `{` / `}` | Previous / next hunk |
| `Ctrl+J` / `Ctrl+K` | Next / previous file |
| `h` / `l` | Scroll left / right (diff focus) |
| `Tab` | Toggle sidebar |
| `1` | Focus sidebar |
| `2` | Focus diff panel |
| `Enter` | Open selected file (sidebar) |
| `Space` | Mark file viewed, jump to next unviewed |
| `[` | Fullscreen old (left) panel |
| `]` | Fullscreen new (right) panel |
| `=` | Exit fullscreen |
| `/` | Search |
| `n` / `N` | Next / previous search result |
| `e` | Open current file in `$EDITOR` |
| `y` | Copy selection to clipboard |
| `i` | Add annotation to current line |
| `I` | View all annotations |
| `Ctrl+P` | File picker |
| `Ctrl+L` / `Ctrl+H` | Next / prev commit (stacked mode) |
| `?` | Help overlay |
| `q` / `Ctrl+C` | Quit |

---

### `gitlens draft`

Generate a conventional commit message for staged changes.

```bash
gitlens draft
gitlens draft --context "fixes the race condition in auth handler"
```

Streams the result directly to stdout — pipe it or copy-paste:

```bash
git commit -m "$(gitlens draft)"
```

---

### `gitlens explain [ref|-]`

Ask AI to explain what changed and why.

```bash
gitlens explain HEAD             # Explain last commit
gitlens explain main..feature    # Explain a branch
gitlens explain --staged         # Explain staged changes
gitlens explain --list           # fzf picker to choose a commit
gitlens explain HEAD --query "why was the mutex added?"
git diff | gitlens explain -     # Pipe a diff from stdin
```

Renders markdown via `mdcat` if available, otherwise plain text.

---

### `gitlens operate <query>`

Turn natural language into a git command — shows what it will run before executing.

```bash
gitlens operate "rebase my branch onto main"
gitlens operate "show me all commits that touched auth.go"
gitlens operate "cherry-pick the last 3 commits from the feature branch"
```

Prompts `[y/N]` before running. Warns on destructive operations.

---

### `gitlens configure`

Interactive setup wizard. Walks through provider selection, API key, model, and theme.

```bash
gitlens configure
```

Writes to `~/.config/gitlens/config.toml`.

---

## Configuration

Configuration is resolved in this order (highest wins):

1. **CLI flags** — `--provider`, `--api-key`, `--model`, `--theme`
2. **Environment variables** — `GITLENS_PROVIDER`, `GITLENS_API_KEY`, `GITLENS_MODEL`, `GITLENS_THEME`
3. **Project config** — `gitlens.config.toml` in the repo root
4. **Global config** — `~/.config/gitlens/config.toml`
5. **Defaults** — Claude, claude-opus-4-6, dark theme

### Config file (`~/.config/gitlens/config.toml`)

```toml
provider = "claude"          # "claude" or "gemini"
api_key  = "sk-ant-..."
model    = "claude-opus-4-6"

[theme]
base = "catppuccin-mocha"    # See themes table below

# Optional per-color overrides (hex or ANSI 256)
[theme.override]
keyword  = "#cba6f7"
added_bg = "#1e3a2f"
```

### Supported models

| Provider | Models |
|----------|--------|
| Claude | `claude-opus-4-6`, `claude-sonnet-4-6`, `claude-haiku-4-5-20251001` |
| Gemini | `gemini-2.0-flash`, `gemini-1.5-pro`, `gemini-1.5-flash` |

---

## Themes

| Name | Style |
|------|-------|
| `dark` | Default dark |
| `light` | Default light |
| `catppuccin-mocha` | Catppuccin Mocha |
| `catppuccin-latte` | Catppuccin Latte |
| `dracula` | Dracula |
| `nord` | Nord |
| `gruvbox-dark` | Gruvbox Dark |
| `gruvbox-light` | Gruvbox Light |
| `one-dark` | One Dark |
| `solarized-dark` | Solarized Dark |
| `solarized-light` | Solarized Light |

Preview a theme without saving it:

```bash
gitlens diff --theme dracula
```

---

## Building from Source

```bash
git clone https://github.com/ithaquaKr/gitlens
cd gitlens

# Full build with syntax highlighting
make build

# Check all targets
make help
```

**Build without a C compiler:**

```bash
make build-nocgo
```

The diff viewer works without syntax highlighting — code is rendered as plain text with diff colors intact.

**Supported platforms:** macOS, Linux, Windows (CGo requires platform C toolchain).

---

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md). The short version:

```bash
make test         # Run all tests
make test-cover   # Tests with coverage report
make lint         # Run golangci-lint
make fmt          # Format code
```

---

## License

MIT — see [LICENSE](LICENSE).
