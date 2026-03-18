# AGENTS.md — AI Agent Instructions for gitlens

> This file is read by AI coding agents (OpenAI Codex, Gemini CLI, etc.). See also [CLAUDE.md](CLAUDE.md) for Claude-specific notes.

---

## Project Overview

`gitlens` is a Go CLI that combines AI language models with git operations:

- **`diff`** — interactive side-by-side diff viewer (Bubble Tea TUI, syntax highlighting)
- **`draft`** — AI-generated conventional commit messages from staged changes
- **`explain`** — plain-English explanation of diffs, commits, or ranges
- **`operate`** — natural-language → git command with confirmation
- **`configure`** — interactive setup wizard

**Module:** `gitlens` | **Go:** 1.24.2+ | **Binary:** `gitlens`

---

## Build

```bash
# Full build — recommended (requires C compiler for syntax highlighting)
CGO_ENABLED=1 go build -o gitlens .

# CGo-free fallback — must also succeed
CGO_ENABLED=0 go build -o gitlens .
```

Always verify BOTH build paths pass after making changes.

---

## Test

```bash
CGO_ENABLED=1 go test ./...    # Primary test run
CGO_ENABLED=0 go test ./...    # Must also pass
go vet ./...                   # No vet errors allowed
```

---

## Key Files

| File | Purpose |
|------|---------|
| `cmd/root.go` | Root Cobra command; loads config into global `Cfg` |
| `cmd/diff.go` | Wires TUI app: loads diff, state, theme, runs `tea.Program` |
| `internal/config/config.go` | Layered config (flags > env > file > defaults) |
| `internal/vcs/git.go` | GitBackend — all git operations via go-git |
| `internal/vcs/backend.go` | Backend interface |
| `internal/ai/provider.go` | Provider interface: `Complete`, `Stream`, `Name` |
| `internal/ai/registry.go` | Thread-safe provider registry (sync.RWMutex) |
| `internal/diff/state.go` | AppState — all mutable TUI state |
| `internal/diff/diff_algo.go` | Myers diff → word-level segments |
| `internal/diff/app/app.go` | Bubble Tea Model — Init/Update/View |
| `internal/diff/render/*.go` | Pure rendering functions (sidebar, diffview, footer, modal) |
| `internal/diff/theme/theme.go` | Theme struct + Load() |
| `internal/diff/highlight/highlight.go` | Tree-sitter syntax highlighting (CGo required) |

---

## Architecture Rules

### CGo build tags

Files using tree-sitter MUST have `//go:build cgo` at the top. Corresponding stubs MUST have `//go:build !cgo`. Both paths must compile and test cleanly.

### Import cycle constraint

`internal/diff/render` imports `internal/diff`. The Bubble Tea `Model` therefore lives in `internal/diff/app` (separate package) to avoid a cycle.

```
ALLOWED:  diff/app → diff, diff/app → diff/render, diff/render → diff
FORBIDDEN: diff → diff/render
```

### AI provider pattern

Providers self-register in `init()` and store the HTTP client on the struct (not per-call). Both `Complete` and `Stream` must honour `ctx.Done()` to avoid goroutine leaks.

### FileDiff model

`git_entity.FileDiff` stores raw strings (`OldContent`, `NewContent`) and `Status` (`"A"`, `"D"`, `"M"`, `"R"`). There are no pre-computed hunks — callers split on `\n`.

---

## Config Precedence (high → low)

1. CLI flags (`--provider`, `--api-key`, `--model`, `--theme`)
2. Env vars (`GITLENS_PROVIDER`, `GITLENS_API_KEY`, `GITLENS_MODEL`, `GITLENS_THEME`)
3. `--config` flag path
4. `gitlens.config.toml` (repo root)
5. `~/.config/gitlens/config.toml`
6. Hardcoded defaults

---

## Common Pitfalls

- **Import cycle:** never import `diff/render` from inside `internal/diff`
- **CGo stubs:** every `//go:build cgo` file needs a `//go:build !cgo` counterpart
- **Client lifetime:** AI provider clients are created once in the factory, not per request
- **`GetCommitsInRange`** returns `[]*git_entity.Commit` (pointer slice)
- **C# tree-sitter** path is `csharp`, not `c_sharp`

---

## Commit Conventions

```
feat:     new capability
fix:      bug correction
refactor: restructure without behavior change
test:     add or update tests
docs:     documentation changes
chore:    build/deps/config
```
