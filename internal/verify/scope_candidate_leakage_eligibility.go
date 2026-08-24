package verify

func init() {
	branchScopeAllowlist["agent/candidate-leakage-eligibility"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/self-improvement-contract.yml",
		"cmd/candidate-leakage-eligibility-witness",
		"examples/candidate-leakage-eligibility",
		"internal/meta/languageassurance/candidateleakageeligibility",
		"internal/verify/scope_candidate_leakage_eligibility.go",
	}
}
