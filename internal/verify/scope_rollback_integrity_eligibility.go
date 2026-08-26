package verify

func init() {
	branchScopeAllowlist["agent/rollback-integrity-eligibility"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/self-improvement-contract.yml",
		"cmd/rollback-integrity-eligibility-witness",
		"examples/rollback-integrity-eligibility",
		"internal/meta/languageassurance/rollbackintegrityeligibility",
		"internal/verify/scope_rollback_integrity_eligibility.go",
	}
}
