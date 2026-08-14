package verify

import (
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
)

func TestCISCOPE003TraversalPathsFailClosed(t *testing.T) {
	allowed := []string{".github", "scripts", "internal/verify"}
	for _, path := range []string{"../scripts/verify.go", "/tmp/verify.go", "scripts/../scripts/verify.go"} {
		if err := CheckPathScope([]string{path}, allowed); err == nil {
			t.Errorf("traversal path %q was accepted", path)
		}
	}
}

func TestCISCOPE004NonCanonicalPathsFailClosed(t *testing.T) {
	allowed := []string{".github", "scripts", "internal/verify"}
	for _, path := range []string{"./scripts/verify.go", "scripts//verify.go", "internal/verify/./policy.go", "scripts\\verify.go"} {
		if err := CheckPathScope([]string{path}, allowed); err == nil {
			t.Errorf("non-canonical path %q was accepted", path)
		}
	}
}

func TestCISCOPE005ImplementationLanesRejectCrossScopePaths(t *testing.T) {
	lanes := map[string][]string{
		"agent/syntax":   {"internal/syntax/parser.go", "internal/semantic/graph.go", "internal/bidir/lens.go", "internal/lsp/server.go"},
		"agent/semantic": {"internal/semantic/graph.go", "internal/syntax/parser.go", "internal/bidir/lens.go", "internal/lsp/server.go"},
		"agent/bidir":    {"internal/bidir/lens.go", "internal/syntax/parser.go", "internal/semantic/graph.go", "internal/lsp/server.go"},
		"agent/lsp":      {"internal/lsp/server.go", "internal/syntax/parser.go", "internal/semantic/graph.go", "internal/bidir/lens.go"},
	}
	for branch, paths := range lanes {
		allowed, ok := BranchScope(branch)
		if !ok {
			t.Fatalf("missing canonical lane %s", branch)
		}
		if err := CheckPathScopeForBranch([]string{paths[0]}, branch); err != nil {
			t.Fatalf("canonical path for %s rejected: %v", branch, err)
		}
		for _, path := range paths[1:] {
			if err := CheckPathScope([]string{path}, allowed); err == nil {
				t.Errorf("%s incorrectly allowed %s", branch, path)
			}
		}
	}
}

func TestCISCOPE006ScopeTableValuesMatchExecutableMap(t *testing.T) {
	table := readScopeTable(t)
	for branch, expected := range branchScopeAllowlist {
		actual, ok := table[branch]
		if !ok {
			t.Errorf("scope table is missing %s", branch)
			continue
		}
		if !reflect.DeepEqual(sortedUnique(expected), sortedUnique(actual)) {
			t.Errorf("scope mismatch for %s: executable=%v table=%v", branch, expected, actual)
		}
	}
}

func TestCISCOPE007ScopeKeysAndRowsAreUnique(t *testing.T) {
	branches := ConfiguredBranches()
	if len(branches) != len(sortedUnique(branches)) {
		t.Fatalf("duplicate executable scope key detected: %v", branches)
	}
	for _, branch := range branches {
		if strings.ContainsAny(branch, "*?") {
			t.Fatalf("wildcard executable scope key detected: %s", branch)
		}
	}
	readScopeTable(t)
}

func readScopeTable(t *testing.T) map[string][]string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "..", ".github", "agent-scope-table.md"))
	if err != nil {
		t.Fatal(err)
	}
	result := make(map[string][]string)
	for _, line := range strings.Split(string(data), "\n") {
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
	for _, value := range strings.Split(cell, ",") {
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
