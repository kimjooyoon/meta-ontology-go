package verify

import (
	"reflect"
	"testing"
)

func TestCIScopePreflightOwnershipIsExact(t *testing.T) {
	want := []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/ci.yml",
		"internal/verify/ci_scope_preflight_order_test.go",
		"internal/verify/scope_ci_preflight_20260907.go",
		"internal/verify/scope_ci_preflight_20260907_test.go",
	}
	got, known := BranchScope("agent/ci-scope-preflight-20260907")
	if !known || !reflect.DeepEqual(got, want) {
		t.Fatalf("preflight ownership is not exact: known=%t paths=%v", known, got)
	}
	if err := CheckPathScopeForBranch(want, "agent/ci-scope-preflight-20260907"); err != nil {
		t.Fatal(err)
	}
}

func TestCIScopePreflightOwnershipDoesNotWidenAuthority(t *testing.T) {
	for _, path := range []string{
		".github/foundation-authorization.json",
		".github/workflows/ci-guardian.yml",
		".github/workflows/gooo-release-publish.yml",
		"internal/verify/scope_part01.go",
		"scripts/verify/main_part02.go",
		"internal/semantic/graph.go",
		"go.mod",
		".github/ci-governance.json.backup",
	} {
		if err := CheckPathScopeForBranch([]string{path}, "agent/ci-scope-preflight-20260907"); err == nil {
			t.Errorf("preflight scope allowed unrelated path %q", path)
		}
	}
	for _, branch := range []string{"agent/ci-scope-preflight-20260907-other", "agent/unknown"} {
		if err := CheckPathScopeForBranch([]string{".github/workflows/ci.yml"}, branch); err == nil {
			t.Errorf("unregistered branch %q inherited preflight scope", branch)
		}
	}
}
