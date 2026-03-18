# GitLens Foundation + CLI Commands — Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build the working `gitlens` CLI binary with all non-TUI commands: `draft`, `explain`, `operate`, and `configure`, backed by a layered config system, Git VCS backend, and extensible AI provider interface (Claude + Gemini).

**Architecture:** Layered internal packages (`config` → `vcs` + `ai` → `cmd`). Each layer depends only on layers below it. The AI provider uses a registry pattern so new providers are added by implementing one interface and calling `Register()` in `init()`. The VCS layer exposes a `Backend` interface backed by `go-git`.

**Tech Stack:** Go 1.22+, Cobra (CLI), go-git (Git), BurntSushi/toml (config), anthropics/anthropic-sdk-go (Claude), google/generative-ai-go (Gemini), charmbracelet/huh (configure wizard), fsnotify (unused here, installed for Plan 2).

---

## File Map

```
gitlens/
├── main.go                          # Cobra Execute()
├── go.mod                           # module gitlens
├── go.sum
├── cmd/
│   ├── root.go                      # Root command, global flags, config bootstrap
│   ├── draft.go                     # gitlens draft
│   ├── explain.go                   # gitlens explain
│   ├── operate.go                   # gitlens operate
│   └── configure.go                 # gitlens configure
├── internal/
│   ├── config/
│   │   ├── config.go                # Config struct + Load() + precedence merge
│   │   └── config_test.go
│   ├── git_entity/
│   │   └── types.go                 # Commit, Diff, FileDiff, DiffLine, Hunk, Segment, etc.
│   ├── vcs/
│   │   ├── backend.go               # Backend interface
│   │   ├── git.go                   # GitBackend (go-git)
│   │   ├── ref.go                   # CommitReference parsing
│   │   ├── git_test.go
│   │   └── testutil_test.go         # helpers: initRepo, addCommit
│   └── ai/
│       ├── provider.go              # Provider interface, StreamChunk, ProviderFactory
│       ├── registry.go              # Register(), New()
│       ├── prompts.go               # ExplainPrompt, DraftPrompt, OperatePrompt
│       ├── claude.go                # ClaudeProvider
│       ├── gemini.go                # GeminiProvider
│       └── mock_test.go             # MockProvider for tests
```

---

## Task 1: Project Scaffold

**Files:**
- Create: `go.mod`
- Create: `main.go`
- Create: `cmd/root.go`

- [ ] **Step 1: Initialize the Go module**

```bash
cd /Users/ithaqua/work/project/gitlens
go mod init gitlens
```

Expected: `go.mod` created with `module gitlens` and current Go version.

- [ ] **Step 2: Install core dependencies**

```bash
go get github.com/spf13/cobra@latest
go get github.com/BurntSushi/toml@latest
go get github.com/go-git/go-git/v5@latest
go get github.com/anthropics/anthropic-sdk-go@latest
go get github.com/google/generative-ai-go/genai@latest
go get github.com/charmbracelet/huh@latest
go get github.com/charmbracelet/bubbletea@latest
go get github.com/charmbracelet/lipgloss@latest
go get github.com/charmbracelet/bubbles@latest
go get github.com/sergi/go-diff@latest
go get github.com/fsnotify/fsnotify@latest
go get github.com/atotto/clipboard@latest
go mod tidy
```

- [ ] **Step 3: Write `main.go`**

```go
package main

import "gitlens/cmd"

func main() {
    cmd.Execute()
}
```

- [ ] **Step 4: Write `cmd/root.go`**

```go
package cmd

import (
    "fmt"
    "os"

    "github.com/spf13/cobra"
    "gitlens/internal/config"
)

var (
    cfgFile    string
    providerFlag string
    apiKeyFlag   string
    modelFlag    string
    themeFlag    string

    // Cfg is loaded once at PersistentPreRunE and shared across subcommands.
    Cfg *config.Config
)

var rootCmd = &cobra.Command{
    Use:   "gitlens",
    Short: "AI-powered git CLI tool",
    PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
        var err error
        Cfg, err = config.Load(config.LoadOptions{
            ExplicitPath: cfgFile,
            Overrides: config.Overrides{
                Provider: providerFlag,
                APIKey:   apiKeyFlag,
                Model:    modelFlag,
                Theme:    themeFlag,
            },
        })
        return err
    },
}

func Execute() {
    if err := rootCmd.Execute(); err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }
}

func init() {
    rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file path")
    rootCmd.PersistentFlags().StringVarP(&providerFlag, "provider", "p", "", "AI provider (claude, gemini)")
    rootCmd.PersistentFlags().StringVarP(&apiKeyFlag, "api-key", "k", "", "API key override")
    rootCmd.PersistentFlags().StringVarP(&modelFlag, "model", "m", "", "model name override")
    rootCmd.PersistentFlags().StringVar(&themeFlag, "theme", "", "color theme override")
}
```

- [ ] **Step 5: Verify build**

```bash
go build ./...
```

Expected: no errors.

- [ ] **Step 6: Commit**

```bash
git add go.mod go.sum main.go cmd/root.go
git commit -m "chore: scaffold project with Cobra root command"
```

---

## Task 2: Config System

**Files:**
- Create: `internal/config/config.go`
- Create: `internal/config/config_test.go`

- [ ] **Step 1: Write the failing tests**

```go
// internal/config/config_test.go
package config_test

import (
    "os"
    "path/filepath"
    "testing"

    "gitlens/internal/config"
)

func TestLoadDefaults(t *testing.T) {
    cfg, err := config.Load(config.LoadOptions{})
    if err != nil {
        t.Fatal(err)
    }
    if cfg.Provider != "claude" {
        t.Errorf("default provider: got %q, want %q", cfg.Provider, "claude")
    }
}

func TestLoadFromFile(t *testing.T) {
    dir := t.TempDir()
    content := `
provider = "gemini"
api_key  = "test-key"
model    = "gemini-pro"
`
    path := filepath.Join(dir, "config.toml")
    if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
        t.Fatal(err)
    }
    cfg, err := config.Load(config.LoadOptions{ExplicitPath: path})
    if err != nil {
        t.Fatal(err)
    }
    if cfg.Provider != "gemini" {
        t.Errorf("provider: got %q, want %q", cfg.Provider, "gemini")
    }
    if cfg.APIKey != "test-key" {
        t.Errorf("api_key: got %q, want %q", cfg.APIKey, "test-key")
    }
}

func TestEnvVarOverridesFileConfig(t *testing.T) {
    dir := t.TempDir()
    content := `provider = "gemini"\napi_key = "file-key"`
    path := filepath.Join(dir, "config.toml")
    os.WriteFile(path, []byte(content), 0o600)

    t.Setenv("GITLENS_API_KEY", "env-key")

    cfg, err := config.Load(config.LoadOptions{ExplicitPath: path})
    if err != nil {
        t.Fatal(err)
    }
    if cfg.APIKey != "env-key" {
        t.Errorf("env var should override file: got %q, want %q", cfg.APIKey, "env-key")
    }
}

func TestCLIOverrideOverridesAll(t *testing.T) {
    t.Setenv("GITLENS_PROVIDER", "gemini")
    cfg, err := config.Load(config.LoadOptions{
        Overrides: config.Overrides{Provider: "claude"},
    })
    if err != nil {
        t.Fatal(err)
    }
    if cfg.Provider != "claude" {
        t.Errorf("CLI override: got %q, want %q", cfg.Provider, "claude")
    }
}
```

- [ ] **Step 2: Run tests to confirm they fail**

```bash
go test ./internal/config/... -v
```

Expected: compilation error — `config` package does not exist yet.

- [ ] **Step 3: Implement `internal/config/config.go`**

```go
package config

import (
    "os"
    "path/filepath"

    "github.com/BurntSushi/toml"
)

// Config is the fully-resolved configuration after applying precedence rules.
type Config struct {
    Provider string
    APIKey   string
    Model    string
    Theme    ThemeConfig
    Draft    DraftConfig
}

type ThemeConfig struct {
    Base     string
    Override ThemeOverride
}

type ThemeOverride struct {
    // Syntax colors
    Keyword         string `toml:"keyword"`
    String          string `toml:"string"`
    Comment         string `toml:"comment"`
    Function        string `toml:"function"`
    FunctionMacro   string `toml:"function_macro"`
    Type            string `toml:"type"`
    Number          string `toml:"number"`
    Operator        string `toml:"operator"`
    Variable        string `toml:"variable"`
    VariableBuiltin string `toml:"variable_builtin"`
    VariableMember  string `toml:"variable_member"`
    Module          string `toml:"module"`
    Tag             string `toml:"tag"`
    Attribute       string `toml:"attribute"`
    Label           string `toml:"label"`
    Punctuation     string `toml:"punctuation"`
    // Diff colors
    AddedBg        string `toml:"added_bg"`
    DeletedBg      string `toml:"deleted_bg"`
    AddedWordBg    string `toml:"added_word_bg"`
    DeletedWordBg  string `toml:"deleted_word_bg"`
    // UI colors
    Border    string `toml:"border"`
    Selection string `toml:"selection"`
}

type DraftConfig struct {
    CommitTypes map[string]string `toml:"commit_types"`
}

// fileConfig mirrors the TOML file structure for unmarshalling.
type fileConfig struct {
    Provider string      `toml:"provider"`
    APIKey   string      `toml:"api_key"`
    Model    string      `toml:"model"`
    Theme    ThemeConfig `toml:"theme"`
    Draft    DraftConfig `toml:"draft"`
}

// Overrides holds values passed via CLI flags.
type Overrides struct {
    Provider string
    APIKey   string
    Model    string
    Theme    string
}

// LoadOptions controls how config is loaded.
type LoadOptions struct {
    ExplicitPath string   // --config flag
    Overrides    Overrides
}

// Load resolves config using the precedence chain:
// CLI flags > env vars > explicit file > project file > global file > defaults
func Load(opts LoadOptions) (*Config, error) {
    cfg := defaults()

    // Layer: global file
    if global, err := globalConfigPath(); err == nil {
        _ = mergeFile(cfg, global) // ignore missing file
    }

    // Layer: project file
    _ = mergeFile(cfg, "gitlens.config.toml")

    // Layer: explicit file
    if opts.ExplicitPath != "" {
        if err := mergeFile(cfg, opts.ExplicitPath); err != nil {
            return nil, err
        }
    }

    // Layer: env vars
    applyEnv(cfg)

    // Layer: CLI overrides
    if opts.Overrides.Provider != "" {
        cfg.Provider = opts.Overrides.Provider
    }
    if opts.Overrides.APIKey != "" {
        cfg.APIKey = opts.Overrides.APIKey
    }
    if opts.Overrides.Model != "" {
        cfg.Model = opts.Overrides.Model
    }
    if opts.Overrides.Theme != "" {
        cfg.Theme.Base = opts.Overrides.Theme
    }

    return cfg, nil
}

func defaults() *Config {
    return &Config{
        Provider: "claude",
        Model:    "claude-opus-4-6",
        Theme:    ThemeConfig{Base: "dark"},
        Draft: DraftConfig{
            CommitTypes: map[string]string{
                "feat":     "A new feature",
                "fix":      "A bug fix",
                "docs":     "Documentation changes",
                "refactor": "Code refactoring",
                "test":     "Adding tests",
                "chore":    "Maintenance tasks",
            },
        },
    }
}

func mergeFile(cfg *Config, path string) error {
    var fc fileConfig
    if _, err := toml.DecodeFile(path, &fc); err != nil {
        return err
    }
    if fc.Provider != "" {
        cfg.Provider = fc.Provider
    }
    if fc.APIKey != "" {
        cfg.APIKey = fc.APIKey
    }
    if fc.Model != "" {
        cfg.Model = fc.Model
    }
    if fc.Theme.Base != "" {
        cfg.Theme.Base = fc.Theme.Base
    }
    cfg.Theme.Override = fc.Theme.Override
    if len(fc.Draft.CommitTypes) > 0 {
        cfg.Draft.CommitTypes = fc.Draft.CommitTypes
    }
    return nil
}

func applyEnv(cfg *Config) {
    if v := os.Getenv("GITLENS_PROVIDER"); v != "" {
        cfg.Provider = v
    }
    if v := os.Getenv("GITLENS_API_KEY"); v != "" {
        cfg.APIKey = v
    }
    if v := os.Getenv("GITLENS_MODEL"); v != "" {
        cfg.Model = v
    }
    if v := os.Getenv("GITLENS_THEME"); v != "" {
        cfg.Theme.Base = v
    }
}

func globalConfigPath() (string, error) {
    home, err := os.UserHomeDir()
    if err != nil {
        return "", err
    }
    return filepath.Join(home, ".config", "gitlens", "config.toml"), nil
}
```

- [ ] **Step 4: Run tests to confirm they pass**

```bash
go test ./internal/config/... -v
```

Expected: all 4 tests PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/config/
git commit -m "feat: add config system with TOML loading and precedence chain"
```

---

## Task 3: Git Entity Types

**Files:**
- Create: `internal/git_entity/types.go`

No tests — these are pure data types.

- [ ] **Step 1: Write `internal/git_entity/types.go`**

```go
package git_entity

import "time"

// --- Git layer models (returned by VCS backend) ---

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

// FileDiff holds both sides of a file change for the TUI renderer.
type FileDiff struct {
    Path       string // current path
    Status     string // "A", "M", "D"
    OldContent string
    NewContent string
    IsBinary   bool
}

// --- Rendering layer models (output of diff_algo) ---

type ChangeType int

const (
    Equal    ChangeType = iota
    Delete
    Insert
    Modified
)

type Segment struct {
    Text      string
    Highlight bool // word-level diff highlight
}

type LineContent struct {
    LineNo int
    Text   string
}

type DiffLine struct {
    OldLine     *LineContent
    NewLine     *LineContent
    ChangeType  ChangeType
    OldSegments []Segment
    NewSegments []Segment
}

// Hunk is a contiguous block of non-Equal lines used for {/} navigation.
type Hunk struct {
    StartIdx int
    EndIdx   int
}
```

- [ ] **Step 2: Verify build**

```bash
go build ./...
```

- [ ] **Step 3: Commit**

```bash
git add internal/git_entity/
git commit -m "feat: add git entity data models"
```

---

## Task 4: VCS Backend — Git

**Files:**
- Create: `internal/vcs/backend.go`
- Create: `internal/vcs/ref.go`
- Create: `internal/vcs/git.go`
- Create: `internal/vcs/git_test.go`
- Create: `internal/vcs/testutil_test.go`

- [ ] **Step 1: Write `internal/vcs/backend.go`**

```go
package vcs

import "gitlens/internal/git_entity"

// Backend abstracts VCS operations. Currently only Git is implemented.
type Backend interface {
    GetCommit(ref string) (*git_entity.Commit, error)
    GetWorkingTreeDiff(staged bool) (*git_entity.Diff, error)
    GetRangeDiff(from, to string, threeDot bool) (*git_entity.Diff, error)
    GetCommitsInRange(from, to string) ([]*git_entity.Commit, error)
    GetFileContentAtRef(path, ref string) (string, error)
    ResolveRef(ref string) (string, error)
}
```

- [ ] **Step 2: Write `internal/vcs/ref.go`**

```go
package vcs

import "strings"

// CommitReference represents a parsed git ref argument.
type CommitReference struct {
    Single    string // HEAD, sha
    From, To  string // range or three-dot
    ThreeDot  bool
}

// ParseRef parses strings like "HEAD", "abc123", "main..feature", "main...feature".
func ParseRef(s string) CommitReference {
    if strings.Contains(s, "...") {
        parts := strings.SplitN(s, "...", 2)
        return CommitReference{From: parts[0], To: parts[1], ThreeDot: true}
    }
    if strings.Contains(s, "..") {
        parts := strings.SplitN(s, "..", 2)
        return CommitReference{From: parts[0], To: parts[1]}
    }
    return CommitReference{Single: s}
}
```

- [ ] **Step 3: Write the failing tests**

```go
// internal/vcs/git_test.go
package vcs_test

import (
    "testing"
    "gitlens/internal/vcs"
)

func TestGetCommit(t *testing.T) {
    repo, cleanup := newTestRepo(t)
    defer cleanup()

    hash := addCommit(t, repo, "test.txt", "hello", "initial commit")

    backend, err := vcs.NewGitBackend(repo)
    if err != nil {
        t.Fatal(err)
    }

    commit, err := backend.GetCommit(hash)
    if err != nil {
        t.Fatal(err)
    }
    if commit.Message != "initial commit" {
        t.Errorf("message: got %q, want %q", commit.Message, "initial commit")
    }
    if commit.Hash != hash {
        t.Errorf("hash: got %q, want %q", commit.Hash, hash)
    }
}

func TestGetWorkingTreeDiff(t *testing.T) {
    repo, cleanup := newTestRepo(t)
    defer cleanup()

    addCommit(t, repo, "file.go", "package main\n", "initial")
    modifyFile(t, repo, "file.go", "package main\n\nfunc main() {}\n")

    backend, _ := vcs.NewGitBackend(repo)
    diff, err := backend.GetWorkingTreeDiff(false)
    if err != nil {
        t.Fatal(err)
    }
    if len(diff.Files) == 0 {
        t.Error("expected at least one changed file")
    }
}

func TestParseRef(t *testing.T) {
    cases := []struct {
        input    string
        single   string
        from, to string
        three    bool
    }{
        {"HEAD", "HEAD", "", "", false},
        {"abc123", "abc123", "", "", false},
        {"main..feature", "", "main", "feature", false},
        {"main...feature", "", "main", "feature", true},
    }
    for _, tc := range cases {
        ref := vcs.ParseRef(tc.input)
        if ref.Single != tc.single || ref.From != tc.from || ref.To != tc.to || ref.ThreeDot != tc.three {
            t.Errorf("ParseRef(%q) = %+v, unexpected", tc.input, ref)
        }
    }
}
```

```go
// internal/vcs/testutil_test.go
package vcs_test

import (
    "os"
    "path/filepath"
    "testing"
    "time"

    gogit "github.com/go-git/go-git/v5"
    "github.com/go-git/go-git/v5/plumbing/object"
)

func newTestRepo(t *testing.T) (string, func()) {
    t.Helper()
    dir := t.TempDir()
    _, err := gogit.PlainInit(dir, false)
    if err != nil {
        t.Fatalf("init repo: %v", err)
    }
    return dir, func() { os.RemoveAll(dir) }
}

func addCommit(t *testing.T, repoPath, filename, content, message string) string {
    t.Helper()
    repo, err := gogit.PlainOpen(repoPath)
    if err != nil {
        t.Fatal(err)
    }
    wt, _ := repo.Worktree()
    path := filepath.Join(repoPath, filename)
    os.WriteFile(path, []byte(content), 0o644)
    wt.Add(filename)
    hash, err := wt.Commit(message, &gogit.CommitOptions{
        Author: &object.Signature{Name: "Test", Email: "test@test.com", When: time.Now()},
    })
    if err != nil {
        t.Fatalf("commit: %v", err)
    }
    return hash.String()
}

func modifyFile(t *testing.T, repoPath, filename, content string) {
    t.Helper()
    os.WriteFile(filepath.Join(repoPath, filename), []byte(content), 0o644)
}
```

- [ ] **Step 4: Run tests to confirm they fail**

```bash
go test ./internal/vcs/... -v
```

Expected: compilation error — `NewGitBackend` not defined.

- [ ] **Step 5: Implement `internal/vcs/git.go`**

```go
package vcs

import (
    "bytes"
    "fmt"
    "strings"

    gogit "github.com/go-git/go-git/v5"
    "github.com/go-git/go-git/v5/plumbing"
    "github.com/go-git/go-git/v5/plumbing/object"
    "gitlens/internal/git_entity"
)

// GitBackend implements Backend using go-git.
type GitBackend struct {
    repo *gogit.Repository
    path string
}

// NewGitBackend opens the git repo at the given path (or searches parent dirs).
func NewGitBackend(path string) (*GitBackend, error) {
    repo, err := gogit.PlainOpenWithOptions(path, &gogit.PlainOpenOptions{DetectDotGit: true})
    if err != nil {
        return nil, fmt.Errorf("opening git repo at %q: %w", path, err)
    }
    return &GitBackend{repo: repo, path: path}, nil
}

func (g *GitBackend) ResolveRef(ref string) (string, error) {
    h, err := g.repo.ResolveRevision(plumbing.Revision(ref))
    if err != nil {
        return "", fmt.Errorf("resolving ref %q: %w", ref, err)
    }
    return h.String(), nil
}

func (g *GitBackend) GetCommit(ref string) (*git_entity.Commit, error) {
    hash, err := g.ResolveRef(ref)
    if err != nil {
        return nil, err
    }
    h := plumbing.NewHash(hash)
    commit, err := g.repo.CommitObject(h)
    if err != nil {
        return nil, fmt.Errorf("getting commit %q: %w", hash, err)
    }
    diff, err := g.commitDiff(commit)
    if err != nil {
        return nil, err
    }
    return &git_entity.Commit{
        Hash:    hash,
        Message: strings.TrimSpace(commit.Message),
        Author:  commit.Author.Name,
        Email:   commit.Author.Email,
        Date:    commit.Author.When,
    }, diff
}

func (g *GitBackend) commitDiff(commit *object.Commit) (*git_entity.Diff, error) {
    var parentTree *object.Tree
    if commit.NumParents() > 0 {
        parent, err := commit.Parent(0)
        if err != nil {
            return nil, err
        }
        parentTree, err = parent.Tree()
        if err != nil {
            return nil, err
        }
    }
    currentTree, err := commit.Tree()
    if err != nil {
        return nil, err
    }
    return g.treeDiff(parentTree, currentTree)
}

func (g *GitBackend) treeDiff(from, to *object.Tree) (*git_entity.Diff, error) {
    changes, err := object.DiffTree(from, to)
    if err != nil {
        return nil, fmt.Errorf("diff trees: %w", err)
    }
    var files []git_entity.FileDiff
    for _, change := range changes {
        fd, err := changeToFileDiff(change)
        if err != nil {
            continue // skip unreadable files
        }
        files = append(files, fd)
    }
    return &git_entity.Diff{Files: files}, nil
}

func changeToFileDiff(change *object.Change) (git_entity.FileDiff, error) {
    action, err := change.Action()
    if err != nil {
        return git_entity.FileDiff{}, err
    }
    status := actionToStatus(action)
    path := change.To.Name
    if path == "" {
        path = change.From.Name
    }
    oldContent, _ := fileContent(change.From)
    newContent, _ := fileContent(change.To)
    return git_entity.FileDiff{
        Path:       path,
        Status:     status,
        OldContent: oldContent,
        NewContent: newContent,
    }, nil
}

func fileContent(entry object.ChangeEntry) (string, error) {
    if entry.TreeEntry.Mode == 0 {
        return "", nil
    }
    blob, err := entry.Tree.TreeEntryFile(&entry.TreeEntry)
    if err != nil {
        return "", err
    }
    content, err := blob.Contents()
    return content, err
}

func actionToStatus(a object.Action) string {
    switch a {
    case object.Insert:
        return "A"
    case object.Delete:
        return "D"
    default:
        return "M"
    }
}

func (g *GitBackend) GetWorkingTreeDiff(staged bool) (*git_entity.Diff, error) {
    wt, err := g.repo.Worktree()
    if err != nil {
        return nil, err
    }
    status, err := wt.Status()
    if err != nil {
        return nil, err
    }
    var files []git_entity.FileDiff
    for path, fs := range status {
        var s string
        code := fs.Staging
        if !staged {
            code = fs.Worktree
        }
        switch code {
        case gogit.Added, gogit.Untracked:
            s = "A"
        case gogit.Deleted:
            s = "D"
        case gogit.Modified:
            s = "M"
        default:
            continue
        }
        newBytes, _ := os.ReadFile(filepath.Join(g.path, path))
        files = append(files, git_entity.FileDiff{
            Path:       path,
            Status:     s,
            NewContent: string(newBytes),
        })
    }
    return &git_entity.Diff{Files: files}, nil
}

func (g *GitBackend) GetRangeDiff(from, to string, threeDot bool) (*git_entity.Diff, error) {
    fromHash, err := g.ResolveRef(from)
    if err != nil {
        return nil, err
    }
    toHash, err := g.ResolveRef(to)
    if err != nil {
        return nil, err
    }
    if threeDot {
        // Use merge-base as from
        fromHash, err = g.mergeBase(fromHash, toHash)
        if err != nil {
            return nil, err
        }
    }
    fromCommit, err := g.repo.CommitObject(plumbing.NewHash(fromHash))
    if err != nil {
        return nil, err
    }
    toCommit, err := g.repo.CommitObject(plumbing.NewHash(toHash))
    if err != nil {
        return nil, err
    }
    fromTree, _ := fromCommit.Tree()
    toTree, _ := toCommit.Tree()
    return g.treeDiff(fromTree, toTree)
}

func (g *GitBackend) mergeBase(a, b string) (string, error) {
    commitA, err := g.repo.CommitObject(plumbing.NewHash(a))
    if err != nil {
        return "", err
    }
    commitB, err := g.repo.CommitObject(plumbing.NewHash(b))
    if err != nil {
        return "", err
    }
    bases, err := commitA.MergeBase(commitB)
    if err != nil || len(bases) == 0 {
        return a, nil // fallback
    }
    return bases[0].Hash.String(), nil
}

func (g *GitBackend) GetCommitsInRange(from, to string) ([]*git_entity.Commit, error) {
    toHash, err := g.ResolveRef(to)
    if err != nil {
        return nil, err
    }
    fromHash, err := g.ResolveRef(from)
    if err != nil {
        return nil, err
    }
    iter, err := g.repo.Log(&gogit.LogOptions{From: plumbing.NewHash(toHash)})
    if err != nil {
        return nil, err
    }
    var commits []*git_entity.Commit
    err = iter.ForEach(func(c *object.Commit) error {
        if c.Hash.String() == fromHash {
            return fmt.Errorf("stop") // stop iteration
        }
        commits = append(commits, &git_entity.Commit{
            Hash:    c.Hash.String(),
            Message: strings.TrimSpace(c.Message),
            Author:  c.Author.Name,
            Email:   c.Author.Email,
            Date:    c.Author.When,
        })
        return nil
    })
    if err != nil && err.Error() != "stop" {
        return nil, err
    }
    return commits, nil
}

func (g *GitBackend) GetFileContentAtRef(path, ref string) (string, error) {
    hash, err := g.ResolveRef(ref)
    if err != nil {
        return "", err
    }
    commit, err := g.repo.CommitObject(plumbing.NewHash(hash))
    if err != nil {
        return "", err
    }
    tree, err := commit.Tree()
    if err != nil {
        return "", err
    }
    file, err := tree.File(path)
    if err != nil {
        return "", fmt.Errorf("file %q at ref %q: %w", path, ref, err)
    }
    return file.Contents()
}

// needed for GetWorkingTreeDiff
var (
    _ = bytes.NewBuffer
)
import "os"
import "path/filepath"
```

> **Note:** The `GetWorkingTreeDiff` implementation above needs `os` and `path/filepath` imports. Consolidate all imports at the top of the file.

- [ ] **Step 6: Fix imports in git.go** — ensure all imports are at the top:

```go
import (
    "fmt"
    "os"
    "path/filepath"
    "strings"

    gogit "github.com/go-git/go-git/v5"
    "github.com/go-git/go-git/v5/plumbing"
    "github.com/go-git/go-git/v5/plumbing/object"
    "gitlens/internal/git_entity"
)
```

Also fix `GetCommit` — it incorrectly returns two values. Replace with:

```go
func (g *GitBackend) GetCommit(ref string) (*git_entity.Commit, error) {
    hash, err := g.ResolveRef(ref)
    if err != nil {
        return nil, err
    }
    h := plumbing.NewHash(hash)
    commit, err := g.repo.CommitObject(h)
    if err != nil {
        return nil, fmt.Errorf("getting commit %q: %w", hash, err)
    }
    return &git_entity.Commit{
        Hash:    hash,
        Message: strings.TrimSpace(commit.Message),
        Author:  commit.Author.Name,
        Email:   commit.Author.Email,
        Date:    commit.Author.When,
    }, nil
}
```

- [ ] **Step 7: Run tests**

```bash
go test ./internal/vcs/... -v
```

Expected: all tests PASS.

- [ ] **Step 8: Commit**

```bash
git add internal/vcs/
git commit -m "feat: add Git VCS backend with go-git"
```

---

## Task 5: AI Provider Layer

**Files:**
- Create: `internal/ai/provider.go`
- Create: `internal/ai/registry.go`
- Create: `internal/ai/prompts.go`
- Create: `internal/ai/claude.go`
- Create: `internal/ai/gemini.go`
- Create: `internal/ai/mock_test.go`
- Create: `internal/ai/registry_test.go`

- [ ] **Step 1: Write `internal/ai/provider.go`**

```go
package ai

import "context"

// Provider is the interface all AI backends must implement.
type Provider interface {
    Complete(ctx context.Context, prompt string) (string, error)
    // Stream is optional — falls back to Complete if not needed.
    // Channel sends text chunks; final chunk with Err != nil signals failure; close = done.
    Stream(ctx context.Context, prompt string) (<-chan StreamChunk, error)
    Name() string
}

// StreamChunk is one token from a streaming response.
type StreamChunk struct {
    Text string
    Err  error
}

// ProviderFactory constructs a Provider from credentials.
type ProviderFactory func(apiKey, model string) (Provider, error)
```

- [ ] **Step 2: Write `internal/ai/registry.go`**

```go
package ai

import (
    "fmt"
    "gitlens/internal/config"
)

var registry = map[string]ProviderFactory{}

// Register adds a provider factory. Call from init() in each provider file.
func Register(name string, factory ProviderFactory) {
    registry[name] = factory
}

// New creates a Provider from the resolved config.
func New(cfg *config.Config) (Provider, error) {
    factory, ok := registry[cfg.Provider]
    if !ok {
        return nil, fmt.Errorf("unknown AI provider %q (registered: %v)", cfg.Provider, registeredNames())
    }
    return factory(cfg.APIKey, cfg.Model)
}

func registeredNames() []string {
    names := make([]string, 0, len(registry))
    for k := range registry {
        names = append(names, k)
    }
    return names
}
```

- [ ] **Step 3: Write `internal/ai/prompts.go`**

```go
package ai

import (
    "fmt"
    "strings"
)

// ExplainPrompt builds the prompt for the explain command.
func ExplainPrompt(diff, query string) string {
    if query != "" {
        return fmt.Sprintf("Given the following git diff, answer this question: %s\n\n```diff\n%s\n```", query, diff)
    }
    return fmt.Sprintf(`Summarize the following git diff in clear, concise prose. Focus on what changed and why it matters.

\`\`\`diff
%s
\`\`\``, diff)
}

// DraftPrompt builds the prompt for the draft command.
func DraftPrompt(diff, context string, commitTypes map[string]string) string {
    types := make([]string, 0, len(commitTypes))
    for k, v := range commitTypes {
        types = append(types, fmt.Sprintf("  %s: %s", k, v))
    }
    contextLine := ""
    if context != "" {
        contextLine = fmt.Sprintf("\nAdditional context from the author: %s\n", context)
    }
    return fmt.Sprintf(`Generate a conventional commit message for the following diff.
%s
Commit types:
%s

Rules:
- Format: <type>(<optional scope>): <description>
- Description must be imperative mood, lowercase, no period
- Keep it under 72 characters
- Output ONLY the commit message, nothing else

\`\`\`diff
%s
\`\`\``, contextLine, strings.Join(types, "\n"), diff)
}

// OperatePrompt builds the prompt for the operate command.
func OperatePrompt(query string) string {
    return fmt.Sprintf(`Generate a git command for the following request: %s

Respond with exactly 3 lines:
1. The git command (e.g. "git rebase -i HEAD~3")
2. A one-sentence explanation
3. Either "WARNING: <reason>" if the command is destructive, or leave this line empty

Output only these lines, no markdown, no extra text.`, query)
}
```

- [ ] **Step 4: Write the failing tests for the registry**

```go
// internal/ai/registry_test.go
package ai_test

import (
    "context"
    "testing"

    "gitlens/internal/ai"
    "gitlens/internal/config"
)

func TestRegistryUnknownProvider(t *testing.T) {
    cfg := &config.Config{Provider: "nonexistent"}
    _, err := ai.New(cfg)
    if err == nil {
        t.Error("expected error for unknown provider")
    }
}

func TestRegistryKnownProvider(t *testing.T) {
    ai.Register("testprovider", func(apiKey, model string) (ai.Provider, error) {
        return &mockProvider{name: "testprovider"}, nil
    })
    cfg := &config.Config{Provider: "testprovider", APIKey: "key", Model: "m"}
    p, err := ai.New(cfg)
    if err != nil {
        t.Fatal(err)
    }
    if p.Name() != "testprovider" {
        t.Errorf("name: got %q, want testprovider", p.Name())
    }
}
```

```go
// internal/ai/mock_test.go
package ai_test

import (
    "context"
    "gitlens/internal/ai"
)

type mockProvider struct{ name string }

func (m *mockProvider) Name() string { return m.name }

func (m *mockProvider) Complete(_ context.Context, prompt string) (string, error) {
    return "mock response for: " + prompt[:min(len(prompt), 20)], nil
}

func (m *mockProvider) Stream(_ context.Context, prompt string) (<-chan ai.StreamChunk, error) {
    ch := make(chan ai.StreamChunk, 1)
    ch <- ai.StreamChunk{Text: "mock"}
    close(ch)
    return ch, nil
}

func min(a, b int) int {
    if a < b { return a }
    return b
}
```

- [ ] **Step 5: Run tests to confirm they fail**

```bash
go test ./internal/ai/... -v
```

Expected: compilation error.

- [ ] **Step 6: Write `internal/ai/claude.go`**

```go
package ai

import (
    "context"
    "fmt"

    "github.com/anthropics/anthropic-sdk-go"
    "github.com/anthropics/anthropic-sdk-go/option"
)

func init() {
    Register("claude", func(apiKey, model string) (Provider, error) {
        if apiKey == "" {
            return nil, fmt.Errorf("claude: api_key is required")
        }
        if model == "" {
            model = "claude-opus-4-6"
        }
        return &claudeProvider{apiKey: apiKey, model: model}, nil
    })
}

type claudeProvider struct {
    apiKey string
    model  string
}

func (c *claudeProvider) Name() string { return "claude" }

func (c *claudeProvider) Complete(ctx context.Context, prompt string) (string, error) {
    client := anthropic.NewClient(option.WithAPIKey(c.apiKey))
    msg, err := client.Messages.New(ctx, anthropic.MessageNewParams{
        Model:     anthropic.Model(c.model),
        MaxTokens: 1024,
        Messages: []anthropic.MessageParam{
            anthropic.NewUserMessage(anthropic.NewTextBlock(prompt)),
        },
    })
    if err != nil {
        return "", fmt.Errorf("claude complete: %w", err)
    }
    if len(msg.Content) == 0 {
        return "", fmt.Errorf("claude: empty response")
    }
    return msg.Content[0].Text, nil
}

func (c *claudeProvider) Stream(ctx context.Context, prompt string) (<-chan StreamChunk, error) {
    ch := make(chan StreamChunk)
    go func() {
        defer close(ch)
        client := anthropic.NewClient(option.WithAPIKey(c.apiKey))
        stream := client.Messages.NewStreaming(ctx, anthropic.MessageNewParams{
            Model:     anthropic.Model(c.model),
            MaxTokens: 1024,
            Messages: []anthropic.MessageParam{
                anthropic.NewUserMessage(anthropic.NewTextBlock(prompt)),
            },
        })
        for stream.Next() {
            event := stream.Current()
            switch e := event.AsUnion().(type) {
            case anthropic.ContentBlockDeltaEvent:
                if e.Delta.Text != "" {
                    ch <- StreamChunk{Text: e.Delta.Text}
                }
            }
        }
        if err := stream.Err(); err != nil {
            ch <- StreamChunk{Err: fmt.Errorf("claude stream: %w", err)}
        }
    }()
    return ch, nil
}
```

- [ ] **Step 7: Write `internal/ai/gemini.go`**

```go
package ai

import (
    "context"
    "fmt"

    "github.com/google/generative-ai-go/genai"
    "google.golang.org/api/option"
)

func init() {
    Register("gemini", func(apiKey, model string) (Provider, error) {
        if apiKey == "" {
            return nil, fmt.Errorf("gemini: api_key is required")
        }
        if model == "" {
            model = "gemini-2.0-flash"
        }
        return &geminiProvider{apiKey: apiKey, model: model}, nil
    })
}

type geminiProvider struct {
    apiKey string
    model  string
}

func (g *geminiProvider) Name() string { return "gemini" }

func (g *geminiProvider) Complete(ctx context.Context, prompt string) (string, error) {
    client, err := genai.NewClient(ctx, option.WithAPIKey(g.apiKey))
    if err != nil {
        return "", fmt.Errorf("gemini client: %w", err)
    }
    defer client.Close()
    model := client.GenerativeModel(g.model)
    resp, err := model.GenerateContent(ctx, genai.Text(prompt))
    if err != nil {
        return "", fmt.Errorf("gemini complete: %w", err)
    }
    if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
        return "", fmt.Errorf("gemini: empty response")
    }
    return fmt.Sprintf("%s", resp.Candidates[0].Content.Parts[0]), nil
}

func (g *geminiProvider) Stream(ctx context.Context, prompt string) (<-chan StreamChunk, error) {
    ch := make(chan StreamChunk)
    go func() {
        defer close(ch)
        client, err := genai.NewClient(ctx, option.WithAPIKey(g.apiKey))
        if err != nil {
            ch <- StreamChunk{Err: fmt.Errorf("gemini client: %w", err)}
            return
        }
        defer client.Close()
        model := client.GenerativeModel(g.model)
        iter := model.GenerateContentStream(ctx, genai.Text(prompt))
        for {
            resp, err := iter.Next()
            if err != nil {
                if err.Error() != "iterator done" {
                    ch <- StreamChunk{Err: fmt.Errorf("gemini stream: %w", err)}
                }
                return
            }
            for _, cand := range resp.Candidates {
                for _, part := range cand.Content.Parts {
                    ch <- StreamChunk{Text: fmt.Sprintf("%s", part)}
                }
            }
        }
    }()
    return ch, nil
}
```

- [ ] **Step 8: Run tests**

```bash
go test ./internal/ai/... -v
```

Expected: registry tests PASS (mock provider). Claude/Gemini won't be integration tested here (require real API keys).

- [ ] **Step 9: Verify build**

```bash
go build ./...
```

- [ ] **Step 10: Commit**

```bash
git add internal/ai/
git commit -m "feat: add AI provider interface with Claude and Gemini implementations"
```

---

## Task 6: `draft` Command

**Files:**
- Create: `cmd/draft.go`

- [ ] **Step 1: Write `cmd/draft.go`**

```go
package cmd

import (
    "context"
    "fmt"
    "os"

    "github.com/spf13/cobra"
    "gitlens/internal/ai"
    _ "gitlens/internal/ai" // ensure providers register
    "gitlens/internal/vcs"
)

var draftContext string

var draftCmd = &cobra.Command{
    Use:   "draft",
    Short: "Generate a conventional commit message for staged changes",
    RunE: func(cmd *cobra.Command, args []string) error {
        backend, err := vcs.NewGitBackend(".")
        if err != nil {
            return err
        }
        diff, err := backend.GetWorkingTreeDiff(true) // staged only
        if err != nil {
            return err
        }
        if len(diff.Files) == 0 {
            return fmt.Errorf("no staged changes found")
        }

        // Build diff text for the prompt
        diffText := buildDiffText(diff)

        provider, err := ai.New(Cfg)
        if err != nil {
            return err
        }

        prompt := ai.DraftPrompt(diffText, draftContext, Cfg.Draft.CommitTypes)
        ctx := context.Background()

        // Use streaming if available, otherwise Complete
        ch, err := provider.Stream(ctx, prompt)
        if err != nil {
            // Fallback to Complete
            result, err := provider.Complete(ctx, prompt)
            if err != nil {
                return err
            }
            fmt.Print(result)
            return nil
        }
        for chunk := range ch {
            if chunk.Err != nil {
                return chunk.Err
            }
            fmt.Print(chunk.Text)
        }
        if !isTerminal(os.Stdout) {
            // No trailing newline when piped (so output can be used directly)
        } else {
            fmt.Println()
        }
        return nil
    },
}

func init() {
    rootCmd.AddCommand(draftCmd)
    draftCmd.Flags().StringVar(&draftContext, "context", "", "additional context about the intent of these changes")
}

func buildDiffText(diff *git_entity.Diff) string {
    var sb strings.Builder
    for _, f := range diff.Files {
        sb.WriteString(fmt.Sprintf("--- %s\n+++ %s\n", f.Path, f.Path))
        // Simple unified diff representation
        oldLines := strings.Split(f.OldContent, "\n")
        newLines := strings.Split(f.NewContent, "\n")
        _ = oldLines
        _ = newLines
        // For the prompt, just include the raw content difference
        sb.WriteString(f.NewContent)
    }
    return sb.String()
}

func isTerminal(f *os.File) bool {
    fi, err := f.Stat()
    if err != nil {
        return false
    }
    return fi.Mode()&os.ModeCharDevice != 0
}
```

> **Note:** `buildDiffText` needs the `git_entity` and `strings` imports. Add to import block:

```go
import (
    "context"
    "fmt"
    "os"
    "strings"

    "github.com/spf13/cobra"
    "gitlens/internal/ai"
    "gitlens/internal/git_entity"
    "gitlens/internal/vcs"
)
```

> **Note on provider registration:** `claude.go` and `gemini.go` live in `package ai` (not sub-packages). Their `init()` functions run automatically when `gitlens/internal/ai` is imported. The blank imports shown above (`_ "gitlens/internal/ai/claude"`) are incorrect — remove them. The single `"gitlens/internal/ai"` import is sufficient to trigger both `Register()` calls.

- [ ] **Step 2: Verify build**

```bash
go build ./...
```

- [ ] **Step 3: Commit**

```bash
git add cmd/draft.go
git commit -m "feat: add draft command for AI commit message generation"
```

---

## Task 7: `explain` Command

**Files:**
- Create: `cmd/explain.go`

- [ ] **Step 1: Write `cmd/explain.go`**

```go
package cmd

import (
    "bufio"
    "context"
    "fmt"
    "os"
    "os/exec"
    "strings"

    "github.com/spf13/cobra"
    "gitlens/internal/ai"
    "gitlens/internal/vcs"
)

var (
    explainStaged bool
    explainQuery  string
    explainList   bool
)

var explainCmd = &cobra.Command{
    Use:   "explain [ref|-]",
    Short: "Explain git changes using AI",
    Args:  cobra.MaximumNArgs(1),
    RunE: func(cmd *cobra.Command, args []string) error {
        backend, err := vcs.NewGitBackend(".")
        if err != nil {
            return err
        }

        var diffText string

        ref := ""
        if len(args) > 0 {
            ref = args[0]
        }

        switch {
        case explainList:
            ref, err = fzfSelectCommit()
            if err != nil {
                return err
            }
            fallthrough
        case ref == "-":
            scanner := bufio.NewScanner(os.Stdin)
            scanner.Scan()
            ref = strings.TrimSpace(scanner.Text())
            if ref == "" {
                return fmt.Errorf("no ref received from stdin")
            }
            fallthrough
        case ref != "":
            parsed := vcs.ParseRef(ref)
            var diff *git_entity.Diff
            if parsed.Single != "" {
                commit, err := backend.GetCommit(parsed.Single)
                if err != nil {
                    return err
                }
                diff = &git_entity.Diff{} // commit diff retrieved separately
                _ = commit
                diff, err = backend.GetRangeDiff(parsed.Single+"^", parsed.Single, false)
                if err != nil {
                    // Initial commit has no parent — use empty tree
                    diff = &git_entity.Diff{}
                }
            } else {
                diff, err = backend.GetRangeDiff(parsed.From, parsed.To, parsed.ThreeDot)
                if err != nil {
                    return err
                }
            }
            diffText = diffToText(diff)
        default:
            diff, err := backend.GetWorkingTreeDiff(explainStaged)
            if err != nil {
                return err
            }
            diffText = diffToText(diff)
        }

        if diffText == "" {
            return fmt.Errorf("no changes to explain")
        }

        provider, err := ai.New(Cfg)
        if err != nil {
            return err
        }

        prompt := ai.ExplainPrompt(diffText, explainQuery)
        ctx := context.Background()

        ch, err := provider.Stream(ctx, prompt)
        if err != nil {
            result, err := provider.Complete(ctx, prompt)
            if err != nil {
                return err
            }
            return printWithMdcat(result)
        }

        var sb strings.Builder
        for chunk := range ch {
            if chunk.Err != nil {
                return chunk.Err
            }
            fmt.Print(chunk.Text)
            sb.WriteString(chunk.Text)
        }
        fmt.Println()
        return nil
    },
}

func init() {
    rootCmd.AddCommand(explainCmd)
    explainCmd.Flags().BoolVar(&explainStaged, "staged", false, "explain staged changes only")
    explainCmd.Flags().StringVar(&explainQuery, "query", "", "ask a specific question about the changes")
    explainCmd.Flags().BoolVar(&explainList, "list", false, "interactively select a commit via fzf")
}

// fzfSelectCommit shells out to fzf with git log output.
func fzfSelectCommit() (string, error) {
    if _, err := exec.LookPath("fzf"); err != nil {
        return "", fmt.Errorf("explain --list requires fzf. Install it from https://github.com/junegunn/fzf")
    }
    logCmd := exec.Command("git", "log", "--oneline", "--color=always")
    fzfCmd := exec.Command("fzf", "--ansi", "--reverse")
    fzfCmd.Stdin, _ = logCmd.StdoutPipe()
    fzfCmd.Stderr = os.Stderr
    if err := logCmd.Start(); err != nil {
        return "", fmt.Errorf("git log: %w", err)
    }
    out, err := fzfCmd.Output()
    logCmd.Wait()
    if err != nil {
        return "", fmt.Errorf("fzf: %w", err)
    }
    fields := strings.Fields(strings.TrimSpace(string(out)))
    if len(fields) == 0 {
        return "", fmt.Errorf("no commit selected")
    }
    return fields[0], nil // first field is the short SHA
}

// printWithMdcat renders markdown via mdcat if available, else plain stdout.
func printWithMdcat(text string) error {
    if _, err := exec.LookPath("mdcat"); err != nil {
        fmt.Println(text)
        return nil
    }
    cmd := exec.Command("mdcat")
    cmd.Stdin = strings.NewReader(text)
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    return cmd.Run()
}

func diffToText(diff *git_entity.Diff) string {
    var sb strings.Builder
    for _, f := range diff.Files {
        sb.WriteString(fmt.Sprintf("=== %s (%s) ===\n", f.Path, f.Status))
        if f.OldContent != "" {
            sb.WriteString("--- old\n")
            sb.WriteString(f.OldContent)
            sb.WriteString("\n")
        }
        sb.WriteString("+++ new\n")
        sb.WriteString(f.NewContent)
        sb.WriteString("\n")
    }
    return sb.String()
}
```

- [ ] **Step 2: Add missing import `git_entity` to explain.go**

Add to import block: `"gitlens/internal/git_entity"`

- [ ] **Step 3: Verify build**

```bash
go build ./...
```

- [ ] **Step 4: Commit**

```bash
git add cmd/explain.go
git commit -m "feat: add explain command with --list fzf, --staged, --query, stdin ref"
```

---

## Task 8: `operate` Command

**Files:**
- Create: `cmd/operate.go`

- [ ] **Step 1: Write `cmd/operate.go`**

```go
package cmd

import (
    "bufio"
    "context"
    "fmt"
    "os"
    "os/exec"
    "strings"

    "github.com/spf13/cobra"
    "gitlens/internal/ai"
)

var operateCmd = &cobra.Command{
    Use:   "operate <query>",
    Short: "Generate and execute a git command from natural language",
    Args:  cobra.MinimumNArgs(1),
    RunE: func(cmd *cobra.Command, args []string) error {
        query := strings.Join(args, " ")

        provider, err := ai.New(Cfg)
        if err != nil {
            return err
        }

        prompt := ai.OperatePrompt(query)
        ctx := context.Background()

        result, err := provider.Complete(ctx, prompt)
        if err != nil {
            return err
        }

        command, explanation, warning := parseOperateResponse(result)
        if command == "" {
            return fmt.Errorf("could not parse command from AI response:\n%s", result)
        }

        fmt.Printf("Command: %s\n", command)
        fmt.Printf("Explanation: %s\n", explanation)
        if warning != "" {
            fmt.Printf("\n⚠  WARNING: %s\n", warning)
        }
        fmt.Printf("\nRun this command? [y/N]: ")

        reader := bufio.NewReader(os.Stdin)
        answer, _ := reader.ReadString('\n')
        answer = strings.TrimSpace(strings.ToLower(answer))

        if answer != "y" {
            fmt.Println("Aborted.")
            return nil
        }

        parts := strings.Fields(command)
        if len(parts) == 0 {
            return fmt.Errorf("empty command")
        }
        execCmd := exec.Command(parts[0], parts[1:]...)
        execCmd.Stdout = os.Stdout
        execCmd.Stderr = os.Stderr
        execCmd.Stdin = os.Stdin
        return execCmd.Run()
    },
}

func init() {
    rootCmd.AddCommand(operateCmd)
}

// parseOperateResponse parses the 3-line AI response for operate.
// Line 1: command, Line 2: explanation, Line 3 (optional): WARNING: ...
func parseOperateResponse(response string) (command, explanation, warning string) {
    lines := strings.Split(strings.TrimSpace(response), "\n")
    if len(lines) >= 1 {
        command = strings.TrimSpace(lines[0])
    }
    if len(lines) >= 2 {
        explanation = strings.TrimSpace(lines[1])
    }
    if len(lines) >= 3 {
        line := strings.TrimSpace(lines[2])
        if strings.HasPrefix(line, "WARNING:") {
            warning = strings.TrimPrefix(line, "WARNING:")
            warning = strings.TrimSpace(warning)
        }
    }
    return
}
```

- [ ] **Step 2: Write test for `parseOperateResponse`**

```go
// cmd/operate_test.go
package cmd

import "testing"

func TestParseOperateResponse(t *testing.T) {
    cases := []struct {
        input               string
        command, explanation, warning string
    }{
        {
            "git rebase -i HEAD~3\nReorder the last 3 commits interactively\nWARNING: rewrites commit history",
            "git rebase -i HEAD~3", "Reorder the last 3 commits interactively", "rewrites commit history",
        },
        {
            "git log --oneline\nShow commit history\n",
            "git log --oneline", "Show commit history", "",
        },
    }
    for _, tc := range cases {
        cmd, exp, warn := parseOperateResponse(tc.input)
        if cmd != tc.command || exp != tc.explanation || warn != tc.warning {
            t.Errorf("parseOperateResponse(%q) = (%q, %q, %q), want (%q, %q, %q)",
                tc.input, cmd, exp, warn, tc.command, tc.explanation, tc.warning)
        }
    }
}
```

- [ ] **Step 3: Run test**

```bash
go test ./cmd/... -run TestParseOperateResponse -v
```

Expected: PASS.

- [ ] **Step 4: Verify build**

```bash
go build ./...
```

- [ ] **Step 5: Commit**

```bash
git add cmd/operate.go cmd/operate_test.go
git commit -m "feat: add operate command with AI-generated git commands and confirmation flow"
```

---

## Task 9: `configure` Command

**Files:**
- Create: `cmd/configure.go`

- [ ] **Step 1: Write `cmd/configure.go`**

```go
package cmd

import (
    "fmt"
    "os"
    "path/filepath"

    "github.com/BurntSushi/toml"
    "github.com/charmbracelet/huh"
    "github.com/spf13/cobra"
)

var configureCmd = &cobra.Command{
    Use:   "configure",
    Short: "Interactive setup wizard",
    RunE: func(cmd *cobra.Command, args []string) error {
        var (
            provider string
            apiKey   string
            model    string
            theme    string
        )

        providerDefaults := map[string]string{
            "claude": "claude-opus-4-6",
            "gemini": "gemini-2.0-flash",
        }

        form := huh.NewForm(
            huh.NewGroup(
                huh.NewSelect[string]().
                    Title("AI Provider").
                    Options(
                        huh.NewOption("Claude (Anthropic)", "claude"),
                        huh.NewOption("Gemini (Google)", "gemini"),
                    ).
                    Value(&provider),
            ),
            huh.NewGroup(
                huh.NewInput().
                    Title("API Key").
                    Description("Leave empty to use environment variable (GITLENS_API_KEY)").
                    EchoMode(huh.EchoModePassword).
                    Value(&apiKey),
                huh.NewInput().
                    Title("Model name").
                    Description("Leave empty for default").
                    Value(&model),
            ),
            huh.NewGroup(
                huh.NewSelect[string]().
                    Title("Color theme").
                    Options(
                        huh.NewOption("Dark", "dark"),
                        huh.NewOption("Light", "light"),
                        huh.NewOption("Catppuccin Mocha", "catppuccin-mocha"),
                        huh.NewOption("Catppuccin Latte", "catppuccin-latte"),
                        huh.NewOption("Dracula", "dracula"),
                        huh.NewOption("Nord", "nord"),
                        huh.NewOption("Gruvbox Dark", "gruvbox-dark"),
                        huh.NewOption("Gruvbox Light", "gruvbox-light"),
                        huh.NewOption("One Dark", "one-dark"),
                        huh.NewOption("Solarized Dark", "solarized-dark"),
                        huh.NewOption("Solarized Light", "solarized-light"),
                    ).
                    Value(&theme),
            ),
        )

        if err := form.Run(); err != nil {
            return err
        }

        if model == "" {
            model = providerDefaults[provider]
        }

        // Build TOML config
        type themeSection struct {
            Base string `toml:"base"`
        }
        type cfg struct {
            Provider string       `toml:"provider"`
            APIKey   string       `toml:"api_key,omitempty"`
            Model    string       `toml:"model"`
            Theme    themeSection `toml:"theme"`
        }

        out := cfg{
            Provider: provider,
            Model:    model,
            Theme:    themeSection{Base: theme},
        }
        if apiKey != "" {
            out.APIKey = apiKey
        }

        configDir := filepath.Join(mustHomeDir(), ".config", "gitlens")
        if err := os.MkdirAll(configDir, 0o755); err != nil {
            return fmt.Errorf("creating config dir: %w", err)
        }
        configPath := filepath.Join(configDir, "config.toml")
        f, err := os.Create(configPath)
        if err != nil {
            return fmt.Errorf("creating config file: %w", err)
        }
        defer f.Close()

        if err := toml.NewEncoder(f).Encode(out); err != nil {
            return fmt.Errorf("writing config: %w", err)
        }

        fmt.Printf("\nConfig saved to %s\n", configPath)
        return nil
    },
}

func init() {
    rootCmd.AddCommand(configureCmd)
}

func mustHomeDir() string {
    h, err := os.UserHomeDir()
    if err != nil {
        panic(err)
    }
    return h
}
```

- [ ] **Step 2: Verify build**

```bash
go build ./...
```

- [ ] **Step 3: Commit**

```bash
git add cmd/configure.go
git commit -m "feat: add configure command with interactive huh wizard"
```

---

## Task 10: Final Integration Verification

- [ ] **Step 1: Run all tests**

```bash
go test ./... -v
```

Expected: all tests PASS.

- [ ] **Step 2: Build binary and smoke-test help output**

```bash
go build -o gitlens .
./gitlens --help
./gitlens draft --help
./gitlens explain --help
./gitlens operate --help
./gitlens configure --help
```

Expected: all subcommands listed with correct flags.

- [ ] **Step 3: Verify `gitlens explain` works in the gitlens repo itself**

```bash
GITLENS_API_KEY=<your-key> ./gitlens explain HEAD
```

Expected: AI-generated explanation of the last commit.

- [ ] **Step 4: Commit binary to .gitignore, tag the milestone**

```bash
echo "gitlens" >> .gitignore
git add .gitignore
git commit -m "chore: ignore built binary"
```
