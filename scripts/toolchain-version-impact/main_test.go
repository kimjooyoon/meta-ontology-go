package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestClassifyOnlyExactSelectors(t *testing.T) {
	for _, test := range []struct {
		line       string
		classified bool
	}{
		{`go-version: '1.27.3'`, true},
		{`GOTOOLCHAIN: go1.27.3 # selected`, true},
		{`go-version: '1.27.2' # previous 1.27.3`, false},
		{`run: test "$(go env GOVERSION)" = "go1.27.3"`, true},
		{`test "$(go env GOVERSION)" = "go1.27.3"`, true},
	} {
		_, got := classify(".github/workflows/ci.yml", test.line, "1.27.3", 1)
		if got != test.classified {
			t.Fatalf("line %q: classified=%t, want %t", test.line, got, test.classified)
		}
	}
}

func TestScanKeepsHistoricalAndUnknownSeparate(t *testing.T) {
	root := t.TempDir()
	write := func(path, data string) {
		t.Helper()
		full := filepath.Join(root, path)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(data), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("go.mod", "module example.org/impact\n\ngo 1.27.3\n")
	write(".github/workflows/ci.yml", "go-version: '1.27.3'\nrun: test go1.27.3\n")
	write(".github/workflows/stale.yml", "go-version: '1.27.2'\n")
	write("docs/releases/old.md", "Built with Go 1.27.2; SDK 0.2.20.\n")
	write("untracked.md", "Go 1.27.3\n")
	for _, args := range [][]string{{"init", "-q"}, {"add", "go.mod", ".github/workflows/ci.yml", ".github/workflows/stale.yml", "docs/releases/old.md"}} {
		command := exec.Command("git", args...)
		command.Dir = root
		if output, err := command.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v: %s", args, err, output)
		}
	}
	result, err := scan(root)
	if err != nil {
		t.Fatal(err)
	}
	if result.ActiveOccurrences != 4 || result.ClassifiedOccurrences != 2 ||
		result.UnresolvedOccurrences != 2 || result.StaleSelectors != 1 ||
		result.HistoricalOccurrences != 1 || result.OutsideScope != 0 {
		t.Fatalf("unexpected impact: %+v", result)
	}
	if err := os.Remove(filepath.Join(root, ".github/workflows/ci.yml")); err != nil {
		t.Fatal(err)
	}
	if _, err := scan(root); err == nil {
		t.Fatal("missing tracked workflow was silently omitted")
	}
}
