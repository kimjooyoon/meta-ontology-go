package linecaps

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
)

func workflowDiscoveryDetail(root string, directory DirectoryMetric) string {
	if directory.Path != ".github/workflows" || directory.DirectFolders != 0 || directory.DirectFiles == 0 {
		return ""
	}
	entries, err := os.ReadDir(filepath.Join(root, filepath.FromSlash(directory.Path)))
	if err != nil || len(entries) != directory.DirectFiles {
		return ""
	}
	for _, entry := range entries {
		info, infoErr := entry.Info()
		extension := strings.ToLower(filepath.Ext(entry.Name()))
		if infoErr != nil || !info.Mode().IsRegular() || (extension != ".yml" && extension != ".yaml") {
			return ""
		}
	}
	return sourcepolicy.WorkflowDiscoveryObservationDetail
}
