package verify

import "testing"

func TestValueProgramRuntimeScopeTableFixScope(t *testing.T) {
	branch := "agent/value-program-runtime-scope-table-fix-20260920"
	paths, ok := BranchScope(branch)
	if !ok || len(paths) != 4 {
		t.Fatalf("scope table fix was not registered exactly: known=%t paths=%d", ok, len(paths))
	}
	if err := CheckPathScopeForBranch([]string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
	}, branch); err != nil {
		t.Fatalf("scope metadata paths were rejected: %v", err)
	}
	if err := CheckPathScopeForBranch([]string{"internal/verify/scope_value_program_runtime_dogfood_min_20260920.go"}, branch); err == nil {
		t.Fatal("unrelated operation scope registration was accepted")
	}
}
