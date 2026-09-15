package verify

func init() {
	branchScopeAllowlist["agent/rollback-integrity-shadow"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/self-improvement-contract.yml",
		"cmd/rollback-integrity-shadow-witness",
		"examples/rollback-integrity-shadow",
		"internal/meta/languageassurance/rollbackintegrityshadow",
		"internal/verify/scope_rollback_integrity_shadow.go",
	}
}
