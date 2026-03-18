# Contributing

Thanks for your interest in contributing to gitlens.

## Getting Started

```bash
git clone https://github.com/your-username/gitlens
cd gitlens
make build    # CGO_ENABLED=1 — verifies your C toolchain is available
make test     # All tests must pass before submitting
```

### C compiler (for syntax highlighting)

Syntax highlighting uses [tree-sitter](https://tree-sitter.github.io/) via CGo. A C compiler is required to build the full feature set:

- **macOS:** `xcode-select --install`
- **Debian/Ubuntu:** `sudo apt install build-essential`
- **Arch:** `sudo pacman -S base-devel`
- **Alpine:** `apk add gcc musl-dev`
- **Windows:** [MSYS2 with MinGW](https://www.msys2.org/) or Visual Studio Build Tools

The tool compiles without CGo (`make build-nocgo`), but syntax highlighting is disabled.

## Development Workflow

```bash
make build        # Build binary
make test         # Run all tests (CGO_ENABLED=1)
make test-cover   # Tests with coverage report
make lint         # golangci-lint (install: https://golangci-lint.run/usage/install/)
make fmt          # Format all Go files
```

### Running a specific test

```bash
go test ./internal/diff/... -run TestComputeDiffLines -v
```

### Testing both build paths

All changes must compile and all tests must pass with both:

```bash
CGO_ENABLED=1 go build ./... && CGO_ENABLED=1 go test ./...
CGO_ENABLED=0 go build ./... && CGO_ENABLED=0 go test ./...
```

## Project Structure

See [CLAUDE.md](CLAUDE.md) for a full architecture overview and important constraints (import cycle rules, CGo build tag requirements, etc.).

The short version:

- `cmd/` — Cobra subcommands, one file per command
- `internal/ai/` — Provider interface and Claude/Gemini implementations
- `internal/vcs/` — Git abstraction via go-git (no system git dependency)
- `internal/config/` — Layered configuration system
- `internal/diff/` — TUI diff viewer (Bubble Tea + Lipgloss)
- `internal/git_entity/` — Shared domain types

## Adding a New AI Provider

1. Create `internal/ai/<name>.go` implementing the `ai.Provider` interface
2. Register it in `init()`:
   ```go
   func init() {
       registry.Register("<name>", func(apiKey, model string) ai.Provider {
           return &myProvider{apiKey: apiKey, model: model}
       })
   }
   ```
3. Add a test in `internal/ai/registry_test.go`
4. Update `cmd/configure.go` to include it in the provider selection list

## Adding a New Theme

Edit `internal/diff/theme/presets.go` — copy an existing preset and adjust the color values.

## Pull Requests

- One logical change per PR
- Include tests for new behavior
- Run `make lint` and `make fmt` before pushing
- Title follows [conventional commits](https://www.conventionalcommits.org/) format: `feat: ...`, `fix: ...`, etc.
- Reference any related issues: `Closes #123`

## Filing Issues

Include:
- OS and Go version (`go version`)
- Build mode (`CGO_ENABLED=0` or `1`)
- The command you ran and the error output
- Git repository context if relevant (number of files, size of diff)
