package diff

import (
	"path/filepath"
	"sort"
	"strings"

	"gitlens/internal/git_entity"
)

// TreeItemKind distinguishes directory entries from file entries in the sidebar tree.
type TreeItemKind int

const (
	TreeItemDir  TreeItemKind = iota
	TreeItemFile TreeItemKind = iota
)

// TreeItem represents one visible row in the sidebar tree.
type TreeItem struct {
	Kind    TreeItemKind
	Name    string // directory or file base name
	DirPath string // full path for dirs; parent dir for files
	FileIdx int    // index into AppState.Files; -1 for dirs
	Depth   int    // nesting depth (0 = top-level)
}

type dirNode struct {
	name     string
	fullPath string
	children []*dirNode
	fileIdxs []int
}

// VisibleTreeItems returns the ordered list of sidebar rows, skipping collapsed subtrees.
func VisibleTreeItems(files []git_entity.FileDiff, collapsedDirs map[string]bool) []TreeItem {
	root := &dirNode{}
	for i, f := range files {
		insertPath(root, f.Path, i)
	}
	sortNode(root)
	var items []TreeItem
	walkNode(root, files, collapsedDirs, -1, &items)
	return items
}

func insertPath(root *dirNode, path string, idx int) {
	parts := strings.Split(path, "/")
	node := root
	for d, part := range parts[:len(parts)-1] {
		fullPath := strings.Join(parts[:d+1], "/")
		var child *dirNode
		for _, c := range node.children {
			if c.name == part {
				child = c
				break
			}
		}
		if child == nil {
			child = &dirNode{name: part, fullPath: fullPath}
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

// walkNode emits tree items depth-first. The root node (depth == -1) is not emitted itself.
func walkNode(node *dirNode, files []git_entity.FileDiff, collapsedDirs map[string]bool, depth int, out *[]TreeItem) {
	if depth >= 0 {
		*out = append(*out, TreeItem{
			Kind:    TreeItemDir,
			Name:    node.name,
			DirPath: node.fullPath,
			FileIdx: -1,
			Depth:   depth,
		})
		if collapsedDirs[node.fullPath] {
			return
		}
	}

	childDepth := depth + 1
	for _, child := range node.children {
		walkNode(child, files, collapsedDirs, childDepth, out)
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
			Kind:    TreeItemFile,
			Name:    filepath.Base(f.Path),
			DirPath: filepath.Dir(f.Path),
			FileIdx: idx,
			Depth:   childDepth,
		})
	}
}
