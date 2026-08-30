package verify

const ciEffortObservationBranch = "agent/ci-effort-observation-v1"

func init() {
	branchScopeAllowlist[ciEffortObservationBranch] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/ci-effort-observation.yml",
		"examples/ci-effort-observation",
		"examples/language-syntax-roundtrip/corpus.json",
		"internal/verify/scope_ci_effort_observation_v1.go",
		"internal/verify/scope_ci_effort_observation_v1_test.go",
		"internal/meta/languageassurance/verticalsliceclosureshadow/contract.go",
		"internal/meta/languageassurance/verticalsliceclosureshadow/evidence/denominator.json",
		"internal/meta/languagereadiness/languagesyntax/conformance/evaluate_test.go",
		"internal/meta/languagereadiness/languagesyntax/model.go",
		"internal/meta/languagereadiness/languagesyntax/registry.go",
		"scripts/ci-effort-observation",
	}
}
