package evidence

import (
	"os"
	"path/filepath"
	"strings"
)

func workflowDiscoveryRoot(physical string, children []os.DirEntry) bool {
	if physical != ".github/workflows" || len(children) == 0 {
		return false
	}
	for _, child := range children {
		extension := strings.ToLower(filepath.Ext(child.Name()))
		if child.IsDir() || child.Type()&os.ModeSymlink != 0 || (extension != ".yml" && extension != ".yaml") {
			return false
		}
	}
	return true
}
