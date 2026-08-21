package linecaps

import (
	"fmt"
	"io/fs"
	"path/filepath"

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
		if handled, visitErr := collectMetricDirectory(absRoot, path, entry, directories); handled || visitErr != nil {
			return visitErr
		}
		return collectMetricFile(absRoot, path, directories, &files)
	})
	if err != nil {
		return LineMetricsReport{}, err
	}

	entries := orderedMetricDirectories(directories)
	aggregateMetricDirectories(directories, entries)

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
