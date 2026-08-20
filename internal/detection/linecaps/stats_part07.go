package linecaps

import (
	"os"
	"path/filepath"
	"strings"
)

func measureFileMetric(fullPath, relative, extension string) (FileMetric, error) {
	metric := FileMetric{Path: relative, Language: FileLanguageOther}
	switch extension {
	case ".go":
		metric.Language = FileLanguageGo
	case ".gooo":
		metric.Language = FileLanguageGooo
	default:
		return metric, nil
	}
	source, err := os.ReadFile(fullPath)
	if err != nil {
		return FileMetric{}, err
	}
	metric.Lines = lineCount(source)
	return metric, nil
}

func accumulateFileMetric(directory *directoryNode, metric FileMetric) {
	switch metric.Language {
	case FileLanguageGo:
		directory.goFiles++
		directory.goLines += metric.Lines
	case FileLanguageGooo:
		directory.goooFiles++
		directory.goooLines += metric.Lines
	}
}

func collectMetricFile(absRoot, path string, directories map[string]*directoryNode, files *[]FileMetric) error {
	relative, err := filepath.Rel(absRoot, path)
	if err != nil {
		return err
	}
	relative = filepath.ToSlash(relative)
	parent := filepath.ToSlash(filepath.Dir(relative))
	ensureDirectoryNode(directories, parent)
	directories[parent].directFiles++
	extension := strings.ToLower(filepath.Ext(relative))
	metric, err := measureFileMetric(path, relative, extension)
	if err != nil {
		return err
	}
	*files = append(*files, metric)
	accumulateFileMetric(directories[parent], metric)
	return nil
}
