package linecaps

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
)

// AnalyzeLineMetrics traverses a workspace and returns folder/file counts and
// extension-specific line totals. It excludes .git and vendor from traversal.
func AnalyzeLineMetrics(root string) (LineMetricsReport, error) {
	if root == "" {
		return LineMetricsReport{}, fmt.Errorf("line metrics root must not be empty")
	}
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return LineMetricsReport{}, err
	}
	directories := map[string]*directoryNode{}
	ensureDirectoryNode(directories, ".")
	files := make([]FileMetric, 0)
	err = filepath.WalkDir(absRoot, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			name := entry.Name()
			if name == ".git" || name == "vendor" {
				return filepath.SkipDir
			}
			relative, relErr := filepath.Rel(absRoot, path)
			if relErr != nil {
				return relErr
			}
			relative = filepath.ToSlash(relative)
			if relative == "." {
				return nil
			}
			ensureDirectoryNode(directories, relative)
			parent := filepath.ToSlash(filepath.Dir(relative))
			ensureDirectoryNode(directories, parent)
			directories[parent].directFolders++
			return nil
		}
		relative, relErr := filepath.Rel(absRoot, path)
		if relErr != nil {
			return relErr
		}
		relative = filepath.ToSlash(relative)
		parent := filepath.ToSlash(filepath.Dir(relative))
		ensureDirectoryNode(directories, parent)
		directories[parent].directFiles++
		extension := strings.ToLower(filepath.Ext(relative))
		metric, metricErr := measureFileMetric(path, relative, extension)
		if metricErr != nil {
			return metricErr
		}
		files = append(files, metric)
		accumulateFileMetric(directories[parent], metric)
		return nil
	})
	if err != nil {
		return LineMetricsReport{}, err
	}

	entries := make([]string, 0, len(directories))
	for path := range directories {
		entries = append(entries, path)
	}
	sort.Slice(entries, func(i, j int) bool {
		iDepth := directoryDepth(entries[i])
		jDepth := directoryDepth(entries[j])
		if iDepth != jDepth {
			return iDepth > jDepth
		}
		if entries[i] == "." {
			return false
		}
		if entries[j] == "." {
			return true
		}
		return entries[i] < entries[j]
	})
	for _, node := range directories {
		node.recursiveFiles = node.directFiles
	}
	for _, path := range entries {
		node := directories[path]
		if node == nil {
			continue
		}
		if path == "." {
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

	sorted := make([]DirectoryMetric, 0, len(directories))
	for _, path := range orderedPaths(directoriesToPaths(directories)) {
		node := directories[path]
		sorted = append(sorted, DirectoryMetric{
			Path:             path,
			DirectFolders:    node.directFolders,
			DirectFiles:      node.directFiles,
			RecursiveFolders: node.recursiveFolders,
			RecursiveFiles:   node.recursiveFiles,
			GoFiles:          node.goFiles,
			GoooFiles:        node.goooFiles,
			GoLines:          node.goLines,
			GoooLines:        node.goooLines,
		})
	}
	report := LineMetricsReport{Root: filepath.ToSlash(root), Files: files, Directories: sorted}
	meta, err := EvaluateLineMetricIndicators(report, sourcepolicy.Default())
	if err != nil {
		return LineMetricsReport{}, err
	}
	report.Meta = meta
	return report, nil
}
