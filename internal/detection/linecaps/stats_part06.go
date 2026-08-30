package linecaps

import (
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
)

func directoryDepth(path string) int {
	if path == "." {
		return 0
	}
	return strings.Count(path, "/") + 1
}

func collectMetricDirectory(absRoot, path string, entry fs.DirEntry, directories map[string]*directoryNode) (bool, error) {
	if !entry.IsDir() {
		return false, nil
	}
	name := entry.Name()
	if name == ".git" || name == "vendor" {
		return true, filepath.SkipDir
	}
	relative, err := filepath.Rel(absRoot, path)
	if err != nil {
		return true, err
	}
	relative = filepath.ToSlash(relative)
	if relative == "." {
		return true, nil
	}
	ensureDirectoryNode(directories, relative)
	parent := filepath.ToSlash(filepath.Dir(relative))
	ensureDirectoryNode(directories, parent)
	directories[parent].directFolders++
	return true, nil
}

func aggregateMetricDirectories(directories map[string]*directoryNode, entries []string) {
	for _, node := range directories {
		node.recursiveFiles = node.directFiles
	}
	for _, path := range entries {
		node := directories[path]
		if node == nil || path == "." {
			continue
		}
		parent := filepath.ToSlash(filepath.Dir(path))
		parentNode := directories[parent]
		if parentNode == nil {
			continue
		}
		parentNode.recursiveFolders += 1 + node.recursiveFolders
		parentNode.recursiveFiles += node.recursiveFiles
		parentNode.goFiles += node.goFiles
		parentNode.goooFiles += node.goooFiles
		parentNode.goLines += node.goLines
		parentNode.goooLines += node.goooLines
	}
}
