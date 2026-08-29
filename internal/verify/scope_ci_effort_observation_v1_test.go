package verify

import "testing"

func TestCIEffortObservationScope(t *testing.T) {
	paths, ok := BranchScope(ciEffortObservationBranch)
	if !ok || len(paths) != 7 {
		t.Fatalf("CI effort branch registration: known=%t paths=%d", ok, len(paths))
	}
	allowed := []string{
		".github/workflows/ci-effort-observation.yml",
		"examples/ci-effort-observation/contract.json",
		"internal/verify/scope_ci_effort_observation_v1.go",
		"scripts/ci-effort-observation/main.go",
	}
	if err := CheckPathScopeForBranch(allowed, ciEffortObservationBranch); err != nil {
		t.Fatalf("representative CI effort paths rejected: %v", err)
	}
	if err := CheckPathScopeForBranch([]string{".github/workflows/ci.yml"}, ciEffortObservationBranch); err == nil {
		t.Fatal("unrelated protected workflow accepted")
	}
}
