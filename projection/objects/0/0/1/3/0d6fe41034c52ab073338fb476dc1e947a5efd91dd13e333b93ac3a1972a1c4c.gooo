package linecaps

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func padSourceToLines(source string, lines int) string {
	current := strings.Count(source, "\n")
	if !strings.HasSuffix(source, "\n") {
		current++
	}
	return source + strings.Repeat("\n", lines-current)
}
func writeGoFile(t *testing.T, root, path, source string) {
	t.Helper()
	fullPath := filepath.Join(root, path)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(fullPath, []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
}
func hasRule(findings []Finding, rule Rule) bool {
	return countRule(findings, rule) > 0
}
func countRule(findings []Finding, rule Rule) int {
	count := 0
	for _, finding := range findings {
		if finding.Rule == rule {
			count++
		}
	}
	return count
}
func hasFinding(findings []Finding, rule Rule) bool {
	for _, finding := range findings {
		if finding.Rule == rule {
			return true
		}
	}
	return false
}
func filterRules(findings []Finding, rules []Rule) []Finding {
	result := make([]Finding, 0, len(findings))
	for _, finding := range findings {
		if slices.Contains(rules, finding.Rule) {
			result = append(result, finding)
		}
	}
	return result
}
