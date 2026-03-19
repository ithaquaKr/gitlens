package vcs

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/utils/merkletrie"
	"gitlens/internal/git_entity"
)

// GitBackend implements Backend using go-git.
type GitBackend struct {
	repo           *gogit.Repository
	path           string
	rootName       string            // name of the main repo
	submodulePaths map[string]string // submodule dir path → submodule name
}

// NewGitBackend opens the git repo at the given path (or searches parent dirs).
func NewGitBackend(path string) (*GitBackend, error) {
	repo, err := gogit.PlainOpenWithOptions(path, &gogit.PlainOpenOptions{DetectDotGit: true})
	if err != nil {
		return nil, fmt.Errorf("opening git repo at %q: %w", path, err)
	}
	g := &GitBackend{repo: repo, path: path, submodulePaths: map[string]string{}}
	g.detectNames()
	return g, nil
}

// detectNames resolves the main repo name and any submodule paths/names.
func (g *GitBackend) detectNames() {
	// Try to derive name from remote origin URL.
	if cfg, err := g.repo.Config(); err == nil {
		if origin, ok := cfg.Remotes["origin"]; ok && len(origin.URLs) > 0 {
			if n := repoNameFromURL(origin.URLs[0]); n != "" {
				g.rootName = n
			}
		}
	}
	// Fall back to the worktree directory basename.
	if g.rootName == "" {
		if wt, err := g.repo.Worktree(); err == nil {
			g.rootName = filepath.Base(wt.Filesystem.Root())
		}
	}
	if g.rootName == "" {
		g.rootName = "repo"
	}

	// Discover submodules from .gitmodules so we can group their files separately.
	if wt, err := g.repo.Worktree(); err == nil {
		if subs, err := wt.Submodules(); err == nil {
			for _, sub := range subs {
				cfg := sub.Config()
				if cfg.Path != "" {
					g.submodulePaths[cfg.Path] = cfg.Name
				}
			}
		}
	}
}

// repoNameFromURL extracts a human-readable repo name from a remote URL.
func repoNameFromURL(url string) string {
	name := filepath.Base(url)
	name = strings.TrimSuffix(name, ".git")
	if name == "" || name == "." || name == "/" {
		return ""
	}
	return name
}

// fileRepoName returns the repo name for a given file path, checking submodule prefixes first.
func (g *GitBackend) fileRepoName(path string) string {
	for prefix, name := range g.submodulePaths {
		if strings.HasPrefix(path, prefix+"/") || path == prefix {
			return name
		}
	}
	return g.rootName
}

// Repo exposes the underlying go-git repository (used by cmd/diff.go for branch name).
func (g *GitBackend) Repo() *gogit.Repository { return g.repo }

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
	return &git_entity.Commit{
		Hash:    hash,
		Message: strings.TrimSpace(commit.Message),
		Author:  commit.Author.Name,
		Email:   commit.Author.Email,
		Date:    commit.Author.When,
	}, nil
}

func (g *GitBackend) commitTree(commit *object.Commit) (*object.Tree, error) {
	return commit.Tree()
}

func (g *GitBackend) parentTree(commit *object.Commit) (*object.Tree, error) {
	if commit.NumParents() == 0 {
		return nil, nil
	}
	parent, err := commit.Parent(0)
	if err != nil {
		return nil, err
	}
	return parent.Tree()
}

func (g *GitBackend) treeDiff(from, to *object.Tree) (*git_entity.Diff, error) {
	changes, err := object.DiffTree(from, to)
	if err != nil {
		return nil, fmt.Errorf("diff trees: %w", err)
	}
	var files []git_entity.FileDiff
	for _, change := range changes {
		fd, err := g.changeToFileDiff(change)
		if err != nil {
			continue
		}
		files = append(files, fd)
	}
	return &git_entity.Diff{Files: files}, nil
}

func (g *GitBackend) changeToFileDiff(change *object.Change) (git_entity.FileDiff, error) {
	action, err := change.Action()
	if err != nil {
		return git_entity.FileDiff{}, err
	}
	path := change.To.Name
	if path == "" {
		path = change.From.Name
	}
	oldContent, _ := fileContent(change.From)
	newContent, _ := fileContent(change.To)
	return git_entity.FileDiff{
		Path:       path,
		Status:     actionToStatus(action),
		OldContent: oldContent,
		NewContent: newContent,
		RepoName:   g.fileRepoName(path),
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
	return blob.Contents()
}

func actionToStatus(a merkletrie.Action) string {
	switch a {
	case merkletrie.Insert:
		return "A"
	case merkletrie.Delete:
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
		code := fs.Worktree
		if staged {
			code = fs.Staging
		}
		var s string
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
		// For modified/deleted files, load the committed version as OldContent
		// so the diff shows real removals, not just additions against empty.
		var oldContent string
		if s == "M" || s == "D" {
			oldContent, _ = g.GetFileContentAtRef(path, "HEAD")
		}
		// For added/modified files, read the on-disk version as NewContent.
		var newContent string
		if s == "A" || s == "M" {
			b, _ := os.ReadFile(filepath.Join(g.path, path))
			newContent = string(b)
		}
		files = append(files, git_entity.FileDiff{
			Path:       path,
			Status:     s,
			OldContent: oldContent,
			NewContent: newContent,
			RepoName:   g.fileRepoName(path),
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
		return a, nil
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
	stopErr := fmt.Errorf("stop")
	var commits []*git_entity.Commit
	err = iter.ForEach(func(c *object.Commit) error {
		if c.Hash.String() == fromHash {
			return stopErr
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
	if err != nil && err != stopErr {
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
