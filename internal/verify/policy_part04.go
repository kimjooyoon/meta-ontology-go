package verify

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/detection/linecaps"
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
func checkDirectoryLayout(root string, policy LinePolicy) []Violation {
	report, err := linecaps.AnalyzeLineMetrics(root)
	if err != nil {
		return []Violation{{Path: ".", Rule: "directory layout", Detail: err.Error()}}
	}
	meta, err := linecaps.EvaluateLineMetricIndicators(report, policy)
	if err != nil {
		return []Violation{{Path: ".", Rule: "directory layout", Detail: err.Error()}}
	}
	violations := make([]Violation, 0)
	for _, indicator := range meta.Failed() {
		if violation, ok := directoryIndicatorViolation(indicator); ok {
			violations = append(violations, violation)
		}
	}
	return violations
}
func discoverGoFiles(root string) ([]string, error) {
	return discoverSourceFiles(root)
}
