package main

import (
	"os"
	"path/filepath"
	"strings"
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
	paths := []string{"go.mod", ".github/workflows/ci.yml", ".github/workflows/stale.yml", "docs/releases/old.md"}
	result, err := scanPaths(root, paths)
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
	if _, err := scanPaths(root, paths); err == nil {
		t.Fatal("missing tracked workflow was silently omitted")
	}
}

func TestProjectionPlanBindsExactSpansAndLeavesUnknownOpen(t *testing.T) {
	root := t.TempDir()
	for name, data := range map[string]string{
		"go.mod":                   "module example.org/impact\n\ngo 1.27.3\n",
		".github/workflows/ci.yml": "go-version: '1.27.3'\nrun: echo go1.27.3\n",
	} {
		full := filepath.Join(root, name)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(data), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	observed, err := scanPaths(root, []string{"go.mod", ".github/workflows/ci.yml"})
	if err != nil {
		t.Fatal(err)
	}
	plan, err := buildProjectionPlan(observed, "1.27.4")
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.CandidateEdits) != 2 || len(plan.Unresolved) != 1 || plan.HistoricalWrites != 0 {
		t.Fatalf("unexpected plan coverage: %+v", plan)
	}
	if plan.ObservationSHA256 == "" || plan.PlanCoreSHA256 == "" || plan.Provenance["@graph"] == nil {
		t.Fatalf("plan lacks bound observation and provenance: %+v", plan)
	}
	for _, edit := range plan.CandidateEdits {
		data, err := os.ReadFile(filepath.Join(root, edit.Path))
		if err != nil {
			t.Fatal(err)
		}
		if string(data[edit.StartByte:edit.EndByte]) != edit.Before || digest(data) != edit.PreimageSHA256 {
			t.Fatalf("candidate lacks exact preimage: %+v", edit)
		}
		if !strings.HasPrefix(edit.RuleID, "urn:gooo:toolchain-projection:rule:") {
			t.Fatalf("candidate lacks source rule: %+v", edit)
		}
	}
	if _, err := buildProjectionPlan(observed, "next"); err == nil {
		t.Fatal("unsupported target version accepted")
	}
}
