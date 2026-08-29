package verify

const ciEffortObservationBranch = "agent/ci-effort-observation-v1"

func init() {
	branchScopeAllowlist[ciEffortObservationBranch] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/ci-effort-observation.yml",
		"examples/ci-effort-observation",
		"internal/verify/scope_ci_effort_observation_v1.go",
		"internal/verify/scope_ci_effort_observation_v1_test.go",
		"scripts/ci-effort-observation",
	}
}
