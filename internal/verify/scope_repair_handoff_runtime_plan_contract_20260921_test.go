package verify

import "testing"

func TestRepairHandoffRuntimePlanContractScope(t *testing.T) {
	branch := "agent/repair-handoff-runtime-plan-contract-20260921"
	paths, ok := BranchScope(branch)
	if !ok || len(paths) != 6 {
		t.Fatalf("runtime-plan contract scope was not registered exactly: known=%t paths=%d", ok, len(paths))
	}
	allowed := []string{
		".github/workflows/domain-observation.yml",
		".github/workflows/repair-handoff-dogfood.yml",
	}
	if err := CheckPathScopeForBranch(allowed, branch); err != nil {
		t.Fatalf("runtime-plan workflow paths were rejected: %v", err)
	}
	if err := CheckPathScopeForBranch([]string{"cmd/gooo/runtime_plan.go"}, branch); err == nil {
		t.Fatal("runtime-plan implementation was accepted outside this contract-only scope")
	}
}
