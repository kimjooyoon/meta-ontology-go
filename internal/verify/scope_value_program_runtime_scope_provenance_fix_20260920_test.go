package verify

import "testing"

func TestValueProgramRuntimeScopeProvenanceFixScope(t *testing.T) {
	branch := "agent/value-program-runtime-scope-provenance-fix-20260920"
	paths, ok := BranchScope(branch)
	if !ok || len(paths) != 6 {
		t.Fatalf("scope provenance fix was not registered exactly: known=%t paths=%d", ok, len(paths))
	}
	if err := CheckPathScopeForBranch([]string{
		"internal/verify/scope_value_program_runtime_dogfood_min_20260920.go",
		"internal/verify/scope_value_program_runtime_dogfood_iszero_20260920.go",
	}, branch); err != nil {
		t.Fatalf("runtime scope registrations were rejected: %v", err)
	}
	if err := CheckPathScopeForBranch([]string{"internal/verify/scope_value_program_runtime_dogfood_max_20260920.go"}, branch); err == nil {
		t.Fatal("unrelated max scope registration was accepted")
	}
}
