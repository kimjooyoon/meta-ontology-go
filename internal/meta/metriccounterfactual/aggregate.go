package metriccounterfactual

import (
	"path"
	"strings"
)

func aggregateMetrics(directories []string, files []FileMetric) ([]DirectoryMetric, Totals) {
	metrics := make([]DirectoryMetric, len(directories))
	index := make(map[string]int, len(directories))
	for position, directory := range directories {
		metrics[position].Path = directory
		index[directory] = position
	}
	for _, directory := range directories {
		if directory == "." {
			continue
		}
		parent := path.Dir(directory)
		metrics[index[parent]].DirectFolders++
		for _, ancestor := range ancestorPaths(parent) {
			metrics[index[ancestor]].RecursiveFolders++
		}
	}
	for _, file := range files {
		parent := path.Dir(file.Path)
		metrics[index[parent]].DirectFiles++
		for _, ancestor := range ancestorPaths(parent) {
			target := &metrics[index[ancestor]]
			target.RecursiveFiles++
			addLanguage(target, file)
		}
	}
	root := metrics[index["."]]
	return metrics, Totals{
		DirectFolders: root.DirectFolders, DirectFiles: root.DirectFiles,
		RecursiveFolders: root.RecursiveFolders, RecursiveFiles: root.RecursiveFiles,
		GoFiles: root.GoFiles, GoooFiles: root.GoooFiles,
		GoLines: root.GoLines, GoooLines: root.GoooLines,
	}
}

func ancestorPaths(directory string) []string {
	result := []string{directory}
	for directory != "." {
		directory = path.Dir(directory)
		result = append(result, directory)
	}
	return result
}

func languageForPath(file string) string {
	if strings.HasSuffix(file, ".gooo") {
		return "gooo"
	}
	if strings.HasSuffix(file, ".go") {
		return "go"
	}
	return "other"
}

func addLanguage(directory *DirectoryMetric, file FileMetric) {
	if file.Language == "go" {
		directory.GoFiles++
		directory.GoLines += file.Lines
	}
	if file.Language == "gooo" {
		directory.GoooFiles++
		directory.GoooLines += file.Lines
	}
}
