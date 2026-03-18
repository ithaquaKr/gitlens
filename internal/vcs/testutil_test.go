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
