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
		input          string
		single         string
		from, to       string
		three          bool
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
