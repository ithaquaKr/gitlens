# Changelog

All notable changes to this project will be documented in this file.

The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).
This project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [0.1.0] — 2026-03-18

Initial release.

### Added

- **`gitlens diff`** — interactive side-by-side diff viewer
  - Side-by-side panels with word-level change highlights (Myers diff algorithm)
  - Syntax highlighting for 30+ languages via tree-sitter (CGo-optional)
  - File sidebar with status badges (`A`, `M`, `D`, `R`) and viewed tracking
  - Hunk navigation (`{` / `}`) and jump-to-file (`Ctrl+J` / `Ctrl+K`)
  - Fullscreen old/new panel modes (`[`, `]`, `=`)
  - In-line annotations (`i` to add, `I` to list)
  - Clipboard copy (`y`) and editor integration (`e`)
  - Search (`/`, `n`, `N`)
  - Mouse support (scroll wheel, click-and-drag selection)
  - `--watch` mode — auto-reloads on file system changes
  - `--stacked` mode — step through commits in a range one-by-one
  - `--focus`, `--file`, `--theme` flags
  - 11 built-in color themes

- **`gitlens draft`** — AI-generated conventional commit messages
  - Streaming output with non-streaming fallback
  - `--context` flag for additional intent hint
  - Respects configurable commit type definitions

- **`gitlens explain`** — AI explanation of diffs and commits
  - Accepts ref, range, `--staged`, stdin (`-`), or fzf `--list` picker
  - Markdown rendering via `mdcat` when available
  - `--query` flag for specific questions

- **`gitlens operate`** — natural language → git command
  - Parses structured AI response (command, explanation, optional warning)
  - Interactive `[y/N]` confirmation before execution
  - Warns on destructive operations

- **`gitlens configure`** — interactive setup wizard
  - Provider, API key, model, and theme selection
  - Writes `~/.config/gitlens/config.toml`

- **Layered configuration** — CLI flags > env vars > project config > global config > defaults
- **Pluggable AI providers** — Claude (Anthropic) and Gemini (Google) via registry pattern
- **Pure-Go git** — no system `git` dependency (uses go-git)
- **Dual build paths** — `CGO_ENABLED=1` (full features) and `CGO_ENABLED=0` (no C compiler required)

[0.1.0]: https://github.com/ithaquaKr/gitlens/releases/tag/v0.1.0
