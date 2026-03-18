# CLAUDE.md — Agent Instructions for gitlens

This file is read by Claude Code and other AI agents. It documents conventions, build commands, architecture, and constraints for this codebase.

---

## Project

`gitlens` is a Go CLI tool that wraps AI providers (Claude, Gemini) and git operations into five commands: `diff`, `draft`, `explain`, `operate`, `configure`.

**Module:** `gitlens` (Go 1.24.2+)
**Binary:** `gitlens`
**Entry point:** `main.go` → `cmd.Execute()`

---

## Build Commands

Always prefer `CGO_ENABLED=1` when building or testing. Syntax highlighting requires CGo.

```bash
# Recommended build (full features, requires C compiler)
CGO_ENABLED=1 go build -o gitlens .

# CGo-free fallback (no syntax highlighting)
CGO_ENABLED=0 go build -o gitlens .

# Both must succeed — never break either build path
```

**macOS:** C compiler is available after `xcode-select --install`
**Linux:** Install `build-essential` (Debian/Ubuntu) or `gcc` (Alpine: `apk add gcc musl-dev`)

---

## Test Commands

```bash
CGO_ENABLED=1 go test ./...           # All tests
CGO_ENABLED=1 go test ./... -v        # Verbose
CGO_ENABLED=1 go test ./internal/...  # Internal packages only
CGO_ENABLED=0 go test ./...           # Must also pass (no CGo)
go vet ./...                          # Static analysis
```

Tests use the standard `testing` package only. No test framework dependencies.

---

## Architecture

```
gitlens/
├── main.go                    # Entry point — calls cmd.Execute()
├── cmd/                       # Cobra subcommands
│   ├── root.go                # Root cmd, global flags, config loading, Cfg var
│   ├── diff.go                # gitlens diff — wires TUI app
│   ├── draft.go               # gitlens draft — AI commit message
│   ├── explain.go             # gitlens explain — AI diff explanation
│   ├── operate.go             # gitlens operate — natural language → git cmd
│   └── configure.go           # gitlens configure — interactive setup wizard
└── internal/
    ├── config/                # Layered config (flags > env > files > defaults)
    ├── git_entity/            # Domain types: Commit, Diff, FileDiff, DiffLine, Hunk
    ├── vcs/                   # VCS abstraction; GitBackend uses go-git (no system git)
    ├── ai/                    # Provider interface + registry; Claude + Gemini impls
    └── diff/                  # TUI diff viewer
        ├── app/               # Bubble Tea Model (package "app", NOT "diff")
        ├── highlight/         # Syntax highlighting via tree-sitter (CGo required)
        ├── render/            # Pure rendering functions (sidebar, diffview, footer, modal)
        ├── theme/             # Theme struct + 11 presets
        ├── state.go           # AppState — all mutable TUI state
        ├── diff_algo.go       # Myers diff → word-level segments
        ├── watcher.go         # fsnotify file watcher for --watch mode
        ├── context.go         # Sticky context lines via tree-sitter (CGo)
        └── context_stub.go    # Fallback stub when CGo=0
```

---

## Key Patterns

### CGo build tags

Files that use tree-sitter must have `//go:build cgo` at the top. Their stubs must have `//go:build !cgo`. Both must compile cleanly.

```go
// highlight.go — requires CGo
//go:build cgo

// highlight_stub.go — fallback
//go:build !cgo
```

Never add tree-sitter imports to a file without the `cgo` build tag.

### AI provider registry

Providers self-register via `init()`:

```go
// In claude.go or gemini.go
func init() {
    registry.Register("claude", func(apiKey, model string) ai.Provider {
        return newClaudeProvider(apiKey, model)
    })
}
```

The registry uses `sync.RWMutex`. `Register` locks for write; `New` and `registeredNames` lock for read.

### Import cycle constraint

`internal/diff/render/` imports `internal/diff` (for types). Therefore the Bubble Tea `Model` cannot live in `package diff` — it would create a cycle. The model lives in `internal/diff/app/` (package `app`).

```
diff/app → diff       OK
diff/app → diff/render  OK
diff/render → diff    OK
diff → diff/render    FORBIDDEN (cycle)
```

### Config precedence

`cmd/root.go` resolves config in `PersistentPreRunE` and stores the result in the global `Cfg` variable. All subcommands read from `Cfg`. The order is:

1. CLI flags (highest)
2. `GITLENS_*` environment variables
3. `--config` flag path
4. `gitlens.config.toml` (repo root)
5. `~/.config/gitlens/config.toml`
6. Hardcoded defaults (lowest)

### FileDiff data model

`git_entity.FileDiff` has `OldContent` and `NewContent` as raw strings, plus `Status` (`"A"`, `"D"`, `"M"`, `"R"`). There are no pre-computed hunks or lines — callers split on `\n` and use `Status` to determine diff direction.

### Streaming pattern

AI providers implement `Stream(ctx, prompt) (<-chan string, error)`. Callers must respect `ctx.Done()` to avoid goroutine leaks. Both `Complete` and `Stream` methods must not create a new HTTP client per call — the client is created once in the factory and stored on the struct.

---

## Common Pitfalls

- **Don't add `internal/diff/render` imports inside `internal/diff`** — import cycle.
- **Don't forget `_ = themeName` is dead code** — set `Cfg.Theme.Base` directly.
- **`GetCommitsInRange` returns `[]*git_entity.Commit`** — pointer slice, not value slice.
- **LSP may report false positives** for files with build tags and test files in worktrees. Trust `go build ./...` and `go test ./...` over LSP warnings.
- **`go-tree-sitter` C# package path** is `github.com/smacker/go-tree-sitter/csharp` (not `c_sharp`).
- **Always verify `CGO_ENABLED=0` builds cleanly** after any change to `highlight/` or `context.go`.

---

## Commit Conventions

Use conventional commits:

```
feat: add new capability
fix: correct wrong behavior
refactor: restructure without behavior change
test: add or fix tests
docs: update documentation
chore: build/deps/config changes
```

---

## Makefile Targets

```bash
make build        # CGO_ENABLED=1 build
make build-nocgo  # CGO_ENABLED=0 build
make test         # CGO_ENABLED=1 go test ./...
make test-cover   # Tests with HTML coverage report
make lint         # golangci-lint run
make fmt          # gofmt -w
make install      # Install to $GOPATH/bin
make clean        # Remove build artifacts
make help         # List all targets
```
