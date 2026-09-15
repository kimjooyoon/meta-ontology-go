package linecaps

import (
	"path/filepath"
	"strings"
)

func isWorkflowDiscoveryRoot(report LineMetricsReport, directory DirectoryMetric) bool {
	if directory.Path != ".github/workflows" || directory.DirectFiles == 0 ||
		directory.DirectFolders != 0 {
		return false
	}
	files := 0
	for _, file := range report.Files {
		if filepath.ToSlash(filepath.Dir(file.Path)) != directory.Path {
			continue
		}
		extension := strings.ToLower(filepath.Ext(file.Path))
		if extension != ".yml" && extension != ".yaml" {
			return false
		}
		files++
	}
	return files == directory.DirectFiles
}
