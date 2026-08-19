package verify

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/detection/linecaps"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func discoverSourceFiles(root string) ([]string, error) {
	files := make([]string, 0)
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() && (entry.Name() == ".git" || entry.Name() == "vendor") {
			return filepath.SkipDir
		}
		if entry.IsDir() {
			return nil
		}
		extension := strings.ToLower(filepath.Ext(entry.Name()))
		if extension != ".go" && extension != ".gooo" {
			return nil
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		files = append(files, filepath.ToSlash(relative))
		return nil
	})
	sort.Strings(files)
	return files, err
}
func checkDirectoryLayout(root string, maxDirectEntries int) []Violation {
	report, err := linecaps.AnalyzeLineMetrics(root)
	if err != nil {
		return []Violation{{Path: ".", Rule: "directory layout", Detail: err.Error()}}
	}
	violations := make([]Violation, 0)
	for _, directory := range report.Directories {
		directEntries := directory.DirectFiles + directory.DirectFolders
		if directEntries > maxDirectEntries {
			violations = append(violations, Violation{
				Path:   directory.Path,
				Rule:   "directory direct entries",
				Actual: directEntries,
				Limit:  maxDirectEntries,
				Detail: "too many direct children",
			})
		}
		if directory.DirectFolders > 0 && directory.DirectFiles > 0 {
			violations = append(violations, Violation{
				Path:   directory.Path,
				Rule:   "directory mixed entries",
				Detail: "must contain either files or folders, not both",
			})
		}
	}
	return violations
}
func discoverGoFiles(root string) ([]string, error) {
	return discoverSourceFiles(root)
}
