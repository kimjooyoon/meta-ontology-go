package verify

func init() {
	branchScopeAllowlist["agent/candidate-leakage-contract"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/self-improvement-contract.yml",
		"cmd/candidate-leakage-witness",
		"examples/candidate-leakage-shadow",
		"internal/meta/languageassurance/candidateleakage",
		"internal/verify/scope_candidate_leakage.go",
	}
}
