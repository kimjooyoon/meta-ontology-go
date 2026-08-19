package linecaps

import (
	"encoding/json"
	"sort"
)

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
