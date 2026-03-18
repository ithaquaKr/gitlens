package diff_test

import (
	"testing"

	"gitlens/internal/diff"
	"gitlens/internal/git_entity"
)

func TestComputeAdded(t *testing.T) {
	lines := diff.ComputeDiffLines("", "hello\nworld\n")
	if len(lines) == 0 {
		t.Fatal("expected diff lines")
	}
	for _, l := range lines {
		if l.ChangeType != git_entity.Insert {
			t.Errorf("expected Insert, got %v", l.ChangeType)
		}
	}
}

func TestComputeDeleted(t *testing.T) {
	lines := diff.ComputeDiffLines("hello\nworld\n", "")
	for _, l := range lines {
		if l.ChangeType != git_entity.Delete {
			t.Errorf("expected Delete, got %v", l.ChangeType)
		}
	}
}

func TestComputeEqual(t *testing.T) {
	lines := diff.ComputeDiffLines("hello\n", "hello\n")
	for _, l := range lines {
		if l.ChangeType != git_entity.Equal {
			t.Errorf("expected Equal, got %v", l.ChangeType)
		}
	}
}

func TestComputeModified(t *testing.T) {
	lines := diff.ComputeDiffLines("hello world\n", "hello earth\n")
	found := false
	for _, l := range lines {
		if l.ChangeType == git_entity.Modified {
			found = true
			if len(l.OldSegments) == 0 && len(l.NewSegments) == 0 {
				t.Error("expected word-level segments on Modified line")
			}
		}
	}
	if !found {
		t.Error("expected at least one Modified line")
	}
}

func TestComputeHunks(t *testing.T) {
	old := "a\nb\nc\nd\ne\nf\ng\n"
	newStr := "a\nb\nX\nd\ne\nf\ng\n"
	lines := diff.ComputeDiffLines(old, newStr)
	hunks := diff.ComputeHunks(lines)
	if len(hunks) == 0 {
		t.Error("expected at least one hunk")
	}
}
