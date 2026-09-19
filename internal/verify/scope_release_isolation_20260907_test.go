package verify

import (
	"reflect"
	"testing"
)

func TestReleaseArtifactIsolationScopeIsExact(t *testing.T) {
	want := []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/gooo-release-publish.yml",
		"internal/verify/scope_release_isolation_20260907.go",
		"internal/verify/scope_release_isolation_20260907_test.go",
	}
	got, known := BranchScope("agent/release-artifact-execution-isolation-20260907")
	if !known || !reflect.DeepEqual(got, want) {
		t.Fatalf("release isolation scope changed: known=%t paths=%v", known, got)
	}
	if err := CheckPathScopeForBranch(want, "agent/release-artifact-execution-isolation-20260907"); err != nil {
		t.Fatal(err)
	}
	for _, denied := range []string{".github/workflows/ci.yml", ".github/foundation-authorization.json",
		".github/workflows/gooo-release-readiness.yml", ".github/workflows/metric-counterfactual.yml",
		"go.mod", "internal/meta/semantic/graph.go"} {
		if err := CheckPathScopeForBranch([]string{denied}, "agent/release-artifact-execution-isolation-20260907"); err == nil {
			t.Errorf("release isolation allowed unrelated path %q", denied)
		}
	}
}
