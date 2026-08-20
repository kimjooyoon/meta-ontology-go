package linecaps

import "os"

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
