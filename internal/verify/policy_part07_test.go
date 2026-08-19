package verify

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestScopeTableMatchesAllowlist(t *testing.T) {
	table, err := os.ReadFile(filepath.Join("..", "..", ".github", "agent-scope-table.md"))
	if err != nil {
		t.Fatal(err)
	}
	registered := make(map[string]bool)
	for _, line := range strings.Split(string(table), "\n") {
		cells := strings.Split(line, "|")
		if len(cells) < 3 {
			continue
		}
		branch := strings.Trim(strings.TrimSpace(cells[1]), "`")
		if !strings.HasPrefix(branch, "agent/") {
			continue
		}
		if registered[branch] {
			t.Fatalf("duplicate branch row in scope table: %q", branch)
		}
		registered[branch] = true
		if _, ok := BranchScope(branch); !ok {
			t.Fatalf("scope table contains unconfigured branch: %q", branch)
		}
	}
	for _, branch := range ConfiguredBranches() {
		if !registered[branch] {
			t.Errorf("scope table is missing configured branch: %q", branch)
		}
	}
}
