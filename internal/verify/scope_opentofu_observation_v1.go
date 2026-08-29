package verify

const opentofuObservationBranch = "agent/opentofu-observation-v1"

func init() {
	branchScopeAllowlist[opentofuObservationBranch] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/opentofu-observation.yml",
		"cmd/opentofu-observation-witness",
		"docs/external/opentofu-observation-v1.md",
		"examples/opentofu-observation",
		"internal/meta/opentofuobservation",
		"internal/verify/scope_opentofu_observation_v1.go",
		"internal/verify/scope_opentofu_observation_v1_test.go",
		"scripts/opentofu-observation",
	}
}
