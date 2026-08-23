package verify

func init() {
	branchScopeAllowlist["agent/readiness-rollback-fixed-point"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/transformation-effect.yml",
		"cmd/language-readiness-witness/proposal-compat",
		"cmd/language-readiness-witness/rollback-fixed-point",
		"docs/language/rollback-fixed-point-recovery.md",
		"internal/meta/languageconcept",
		"internal/meta/languagereadiness/proposalcompat",
		"internal/meta/languagereadiness/rollbackfixedpoint",
		"internal/meta/transformationeffect/validate_ledger.go",
		"internal/verify/scope_readiness_rollback_fixed_point.go",
	}
}
