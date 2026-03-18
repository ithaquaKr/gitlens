package diff

import (
	"strings"

	"github.com/sergi/go-diff/diffmatchpatch"
	"gitlens/internal/git_entity"
)

// ComputeDiffLines produces a side-by-side []DiffLine from two file content strings.
func ComputeDiffLines(oldContent, newContent string) []git_entity.DiffLine {
	dmp := diffmatchpatch.New()
	oldLines := splitLines(oldContent)
	newLines := splitLines(newContent)

	chars1, chars2, lineArray := dmp.DiffLinesToChars(
		strings.Join(oldLines, ""),
		strings.Join(newLines, ""),
	)
	diffs := dmp.DiffMain(chars1, chars2, false)
	diffs = dmp.DiffCharsToLines(diffs, lineArray)

	var result []git_entity.DiffLine
	oldIdx, newIdx := 1, 1

	for _, d := range diffs {
		lines := splitLines(d.Text)
		switch d.Type {
		case diffmatchpatch.DiffEqual:
			for _, l := range lines {
				if l == "" {
					continue
				}
				result = append(result, git_entity.DiffLine{
					OldLine:    &git_entity.LineContent{LineNo: oldIdx, Text: l},
					NewLine:    &git_entity.LineContent{LineNo: newIdx, Text: l},
					ChangeType: git_entity.Equal,
				})
				oldIdx++
				newIdx++
			}
		case diffmatchpatch.DiffDelete:
			for _, l := range lines {
				if l == "" {
					continue
				}
				result = append(result, git_entity.DiffLine{
					OldLine:    &git_entity.LineContent{LineNo: oldIdx, Text: l},
					ChangeType: git_entity.Delete,
				})
				oldIdx++
			}
		case diffmatchpatch.DiffInsert:
			for _, l := range lines {
				if l == "" {
					continue
				}
				result = append(result, git_entity.DiffLine{
					NewLine:    &git_entity.LineContent{LineNo: newIdx, Text: l},
					ChangeType: git_entity.Insert,
				})
				newIdx++
			}
		}
	}

	result = pairModified(result)
	return result
}

func pairModified(lines []git_entity.DiffLine) []git_entity.DiffLine {
	dmp := diffmatchpatch.New()
	result := make([]git_entity.DiffLine, 0, len(lines))
	i := 0
	for i < len(lines) {
		if i+1 < len(lines) &&
			lines[i].ChangeType == git_entity.Delete &&
			lines[i+1].ChangeType == git_entity.Insert {
			oldStr := lines[i].OldLine.Text
			newStr := lines[i+1].NewLine.Text
			oldSegs, newSegs := wordDiff(dmp, oldStr, newStr)
			result = append(result, git_entity.DiffLine{
				OldLine:     lines[i].OldLine,
				NewLine:     lines[i+1].NewLine,
				ChangeType:  git_entity.Modified,
				OldSegments: oldSegs,
				NewSegments: newSegs,
			})
			i += 2
			continue
		}
		result = append(result, lines[i])
		i++
	}
	return result
}

func wordDiff(dmp *diffmatchpatch.DiffMatchPatch, oldStr, newStr string) ([]git_entity.Segment, []git_entity.Segment) {
	diffs := dmp.DiffMain(oldStr, newStr, false)
	diffs = dmp.DiffCleanupSemantic(diffs)

	unchanged := 0
	total := len(oldStr) + len(newStr)
	for _, d := range diffs {
		if d.Type == diffmatchpatch.DiffEqual {
			unchanged += len(d.Text) * 2
		}
	}
	if total > 0 && float64(unchanged)/float64(total) < 0.2 {
		return []git_entity.Segment{{Text: oldStr}}, []git_entity.Segment{{Text: newStr}}
	}

	var oldSegs, newSegs []git_entity.Segment
	for _, d := range diffs {
		switch d.Type {
		case diffmatchpatch.DiffEqual:
			oldSegs = append(oldSegs, git_entity.Segment{Text: d.Text})
			newSegs = append(newSegs, git_entity.Segment{Text: d.Text})
		case diffmatchpatch.DiffDelete:
			oldSegs = append(oldSegs, git_entity.Segment{Text: d.Text, Highlight: true})
		case diffmatchpatch.DiffInsert:
			newSegs = append(newSegs, git_entity.Segment{Text: d.Text, Highlight: true})
		}
	}
	return oldSegs, newSegs
}

// ComputeHunks identifies contiguous blocks of non-Equal lines.
func ComputeHunks(lines []git_entity.DiffLine) []git_entity.Hunk {
	var hunks []git_entity.Hunk
	inHunk := false
	start := 0
	for i, l := range lines {
		if l.ChangeType != git_entity.Equal {
			if !inHunk {
				start = i
				inHunk = true
			}
		} else {
			if inHunk {
				hunks = append(hunks, git_entity.Hunk{StartIdx: start, EndIdx: i - 1})
				inHunk = false
			}
		}
	}
	if inHunk {
		hunks = append(hunks, git_entity.Hunk{StartIdx: start, EndIdx: len(lines) - 1})
	}
	return hunks
}

func splitLines(s string) []string {
	if s == "" {
		return nil
	}
	lines := strings.Split(s, "\n")
	result := make([]string, 0, len(lines))
	for _, l := range lines {
		if l != "" {
			result = append(result, l+"\n")
		}
	}
	return result
}
