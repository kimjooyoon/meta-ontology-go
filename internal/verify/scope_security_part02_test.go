package verify

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

func readScopeTable(t *testing.T) map[string][]string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "..", ".github", "agent-scope-table.md"))
	if err != nil {
		t.Fatal(err)
	}
	result := make(map[string][]string)
	for line := range strings.SplitSeq(string(data), "\n") {
		cells := strings.Split(line, "|")
		if len(cells) < 3 {
			continue
		}
		branch := strings.Trim(strings.TrimSpace(cells[1]), "`")
		if !strings.HasPrefix(branch, "agent/") {
			continue
		}
		if _, duplicate := result[branch]; duplicate {
			t.Fatalf("duplicate scope table row: %s", branch)
		}
		result[branch] = tablePaths(cells[2])
	}
	return result
}
func tablePaths(cell string) []string {
	paths := make([]string, 0)
	for value := range strings.SplitSeq(cell, ",") {
		value = strings.TrimSpace(value)
		if marker := strings.Index(value, " ("); marker >= 0 {
			value = value[:marker]
		}
		value = strings.Trim(strings.TrimSpace(value), "`")
		value = strings.TrimSuffix(value, "/**")
		if value != "" {
			paths = append(paths, value)
		}
	}
	sort.Strings(paths)
	return paths
}
