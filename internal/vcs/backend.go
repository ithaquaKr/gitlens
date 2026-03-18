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
