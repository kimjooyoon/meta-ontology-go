package linecaps

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// FileLanguage indicates which extension is used to classify source files.
type FileLanguage string

const (
	FileLanguageGo    FileLanguage = "go"
	FileLanguageGooo  FileLanguage = "gooo"
	FileLanguageOther FileLanguage = "other"
)

// FileMetric is a per-file line-count metric for recognized source files.
type FileMetric struct {
	Path     string       `json:"path"`
	Language FileLanguage `json:"language"`
	Lines    int          `json:"lines"`
}

// DirectoryMetric is a directory-level aggregate.
type DirectoryMetric struct {
	Path             string `json:"path"`
	DirectFolders    int    `json:"direct_folders"`
	DirectFiles      int    `json:"direct_files"`
	RecursiveFolders int    `json:"recursive_folders"`
	RecursiveFiles   int    `json:"recursive_files"`
	GoFiles          int    `json:"go_files"`
	GoooFiles        int    `json:"gooo_files"`
	GoLines          int    `json:"go_lines"`
	GoooLines        int    `json:"gooo_lines"`
}

// LineMetricsReport is the repository's line and layout metric output.
type LineMetricsReport struct {
	Root        string            `json:"root"`
	Files       []FileMetric      `json:"files"`
	Directories []DirectoryMetric `json:"directories"`
}

type directoryNode struct {
	directFolders    int
	directFiles      int
	recursiveFolders int
	recursiveFiles   int
	goFiles          int
	goooFiles        int
	goLines          int
	goooLines        int
}

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
		source, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		sourceLines := lineCount(source)
		switch extension {
		case ".go":
			directories[parent].goFiles++
			directories[parent].goLines += sourceLines
			files = append(files, FileMetric{Path: relative, Language: FileLanguageGo, Lines: sourceLines})
		case ".gooo":
			directories[parent].goooFiles++
			directories[parent].goooLines += sourceLines
			files = append(files, FileMetric{Path: relative, Language: FileLanguageGooo, Lines: sourceLines})
		default:
			files = append(files, FileMetric{Path: relative, Language: FileLanguageOther, Lines: sourceLines})
		}
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
		iDepth := strings.Count(entries[i], "/")
		jDepth := strings.Count(entries[j], "/")
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
	for _, path := range entries {
		node := directories[path]
		if node == nil {
			continue
		}
		node.recursiveFiles = node.directFiles
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
	return LineMetricsReport{Root: filepath.ToSlash(root), Files: files, Directories: sorted}, nil
}

func directoriesToPaths(directories map[string]*directoryNode) []string {
	paths := make([]string, 0, len(directories))
	for path := range directories {
		paths = append(paths, path)
	}
	return paths
}

func orderedPaths(paths []string) []string {
	sort.Strings(paths)
	for i, path := range paths {
		if path == "." {
			if i != 0 {
				paths[0], paths[i] = paths[i], paths[0]
			}
			break
		}
	}
	return paths
}

func ensureDirectoryNode(nodes map[string]*directoryNode, path string) {
	if path == "" {
		path = "."
	}
	if _, ok := nodes[path]; ok {
		return
	}
	nodes[path] = &directoryNode{}
}

// JSON returns the deterministic machine-readable report.
func (r LineMetricsReport) JSON() ([]byte, error) {
	report := LineMetricsReport{Root: r.Root, Files: orderedFileMetrics(r.Files), Directories: orderedDirectoryMetrics(r.Directories)}
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(data, '\n'), nil
}

func orderedFileMetrics(files []FileMetric) []FileMetric {
	ordered := append([]FileMetric(nil), files...)
	sort.Slice(ordered, func(i, j int) bool {
		if ordered[i].Path != ordered[j].Path {
			return ordered[i].Path < ordered[j].Path
		}
		if ordered[i].Language != ordered[j].Language {
			return ordered[i].Language < ordered[j].Language
		}
		return ordered[i].Lines < ordered[j].Lines
	})
	return ordered
}

func orderedDirectoryMetrics(directories []DirectoryMetric) []DirectoryMetric {
	ordered := append([]DirectoryMetric(nil), directories...)
	sort.Slice(ordered, func(i, j int) bool {
		if ordered[i].Path == "." {
			return true
		}
		if ordered[j].Path == "." {
			return false
		}
		return ordered[i].Path < ordered[j].Path
	})
	return ordered
}

// Text returns a stable line-oriented report.
func (r LineMetricsReport) Text() string {
	var output strings.Builder
	sum := r.Total()
	fmt.Fprintf(&output, "line metrics: files=%d dirs=%d go_lines=%d gooo_lines=%d\n", sum.RecursiveFiles, sum.RecursiveFolders, sum.GoLines, sum.GoooLines)
	for _, directory := range orderedDirectoryMetrics(r.Directories) {
		fmt.Fprintf(&output, "%s: direct_folders=%d direct_files=%d folders=%d files=%d go_files=%d gooo_files=%d go_lines=%d gooo_lines=%d\n",
			directory.Path, directory.DirectFolders, directory.DirectFiles, directory.RecursiveFolders, directory.RecursiveFiles, directory.GoFiles, directory.GoooFiles, directory.GoLines, directory.GoooLines,
		)
	}
	return output.String()
}

// Total returns aggregate metrics at the repository root.
func (r LineMetricsReport) Total() DirectoryMetric {
	if len(r.Directories) == 0 {
		return DirectoryMetric{}
	}
	for _, directory := range r.Directories {
		if directory.Path == "." {
			return directory
		}
	}
	return r.Directories[0]
}
