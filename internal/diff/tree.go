package diff

import (
	"path/filepath"
	"sort"
	"strings"

	"gitlens/internal/git_entity"
)

// TreeItemKind distinguishes the three kinds of rows in the sidebar tree.
type TreeItemKind int

const (
	TreeItemRepo TreeItemKind = iota // repository group header
	TreeItemDir                      // directory entry
	TreeItemFile                     // file entry
)

// TreeItem represents one visible row in the sidebar tree.
type TreeItem struct {
	Kind     TreeItemKind
	Name     string // repo name, directory base name, or file base name
	DirPath  string // unique key for collapse state: "<repoName>:<relPath>" for dirs
	RepoName string // which repo this item belongs to
	FileIdx  int    // index into AppState.Files; -1 for repos and dirs
	Depth    int    // visual nesting depth (0 = repo header, 1 = top-level in repo, …)
}

type dirNode struct {
	name     string
	relPath  string // path relative to repo root
	children []*dirNode
	fileIdxs []int
}

// VisibleTreeItems returns the ordered list of sidebar rows, grouped by repository,
// with collapsed subtrees omitted.
func VisibleTreeItems(files []git_entity.FileDiff, collapsedDirs map[string]bool) []TreeItem {
	// Group files by repo, preserving first-seen order.
	var repoOrder []string
	repoFiles := map[string][]int{}
	for i, f := range files {
		name := f.RepoName
		if name == "" {
			name = "repo"
		}
		if _, seen := repoFiles[name]; !seen {
			repoOrder = append(repoOrder, name)
		}
		repoFiles[name] = append(repoFiles[name], i)
	}

	var items []TreeItem
	for _, repoName := range repoOrder {
		// Repo header row.
		items = append(items, TreeItem{
			Kind:     TreeItemRepo,
			Name:     repoName,
			RepoName: repoName,
			FileIdx:  -1,
			Depth:    0,
		})

		// Build the directory tree for this repo's files.
		root := &dirNode{}
		for _, idx := range repoFiles[repoName] {
			insertPath(root, files[idx].Path, idx)
		}
		sortNode(root)
		// Items under a repo header start at depth 1.
		walkNode(root, files, collapsedDirs, repoName, -1, 1, &items)
	}
	return items
}

// AllFilesInTreeOrder returns file indices ordered by the tree's display order
// (alphabetical within each directory, dirs before files), ignoring collapse state.
// Use this for J/K file navigation so that it matches what the user sees in the sidebar.
func AllFilesInTreeOrder(files []git_entity.FileDiff) []int {
	items := VisibleTreeItems(files, map[string]bool{})
	var order []int
	for _, item := range items {
		if item.Kind == TreeItemFile {
			order = append(order, item.FileIdx)
		}
	}
	return order
}

// AdjacentFileIdx finds currentIdx in the tree order and returns the index one
// step in direction dir (+1 or -1). Returns currentIdx if already at a boundary.
func AdjacentFileIdx(files []git_entity.FileDiff, currentIdx, dir int) int {
	order := AllFilesInTreeOrder(files)
	for i, idx := range order {
		if idx == currentIdx {
			next := i + dir
			if next >= 0 && next < len(order) {
				return order[next]
			}
			return currentIdx // at boundary
		}
	}
	// currentIdx not found in tree (shouldn't happen); fall back to array order.
	next := currentIdx + dir
	if next >= 0 && next < len(files) {
		return next
	}
	return currentIdx
}

func insertPath(root *dirNode, path string, idx int) {
	parts := strings.Split(path, "/")
	node := root
	for d, part := range parts[:len(parts)-1] {
		relPath := strings.Join(parts[:d+1], "/")
		var child *dirNode
		for _, c := range node.children {
			if c.name == part {
				child = c
				break
			}
		}
		if child == nil {
			child = &dirNode{name: part, relPath: relPath}
			node.children = append(node.children, child)
		}
		node = child
	}
	node.fileIdxs = append(node.fileIdxs, idx)
}

func sortNode(node *dirNode) {
	sort.Slice(node.children, func(i, j int) bool {
		return node.children[i].name < node.children[j].name
	})
	for _, c := range node.children {
		sortNode(c)
	}
}

// walkNode emits tree items depth-first.
// repoName is used to build unique DirPath keys (prevents collisions across repos).
// depth == -1 means the root node (not emitted); depthShift is added to all emitted depths.
func walkNode(node *dirNode, files []git_entity.FileDiff, collapsedDirs map[string]bool,
	repoName string, depth, depthShift int, out *[]TreeItem) {

	if depth >= 0 {
		dirPath := repoName + ":" + node.relPath
		*out = append(*out, TreeItem{
			Kind:     TreeItemDir,
			Name:     node.name,
			DirPath:  dirPath,
			RepoName: repoName,
			FileIdx:  -1,
			Depth:    depth + depthShift,
		})
		if collapsedDirs[dirPath] {
			return
		}
	}

	childDepth := depth + 1
	for _, child := range node.children {
		walkNode(child, files, collapsedDirs, repoName, childDepth, depthShift, out)
	}

	// Sort files in this directory by base name.
	sorted := make([]int, len(node.fileIdxs))
	copy(sorted, node.fileIdxs)
	sort.Slice(sorted, func(i, j int) bool {
		return filepath.Base(files[sorted[i]].Path) < filepath.Base(files[sorted[j]].Path)
	})
	for _, idx := range sorted {
		f := files[idx]
		*out = append(*out, TreeItem{
			Kind:     TreeItemFile,
			Name:     filepath.Base(f.Path),
			DirPath:  repoName + ":" + filepath.Dir(f.Path),
			RepoName: repoName,
			FileIdx:  idx,
			Depth:    childDepth + depthShift,
		})
	}
}
