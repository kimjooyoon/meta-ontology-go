package verify

import "testing"

func TestOpenTofuObservationScope(t *testing.T) {
	paths, ok := BranchScope(opentofuObservationBranch)
	if !ok || len(paths) != 10 {
		t.Fatalf("OpenTofu branch was not registered exactly: known=%t paths=%d", ok, len(paths))
	}
	allowed := []string{
		".github/workflows/opentofu-observation.yml",
		"cmd/opentofu-observation-witness/run.go",
		"docs/external/opentofu-observation-v1.md",
		"examples/opentofu-observation/main.gooo",
		"internal/meta/opentofuobservation/evaluate.go",
		"scripts/opentofu-observation/main.sh",
		"internal/verify/scope_opentofu_observation_v1.go",
		"internal/verify/scope_opentofu_observation_v1_test.go",
		".github/agent-scope-table.md",
		".github/ci-governance.json",
	}
	if err := CheckPathScopeForBranch(allowed, opentofuObservationBranch); err != nil {
		t.Fatalf("representative OpenTofu paths were rejected: %v", err)
	}
	if err := CheckPathScopeForBranch([]string{"docs/unrelated.md"}, opentofuObservationBranch); err == nil {
		t.Fatal("unrelated path was accepted")
	}
}
