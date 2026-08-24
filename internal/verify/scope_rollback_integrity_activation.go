package verify

func init() {
	branchScopeAllowlist["agent/rollback-integrity-activation"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/self-improvement-contract.yml",
		"cmd/rollback-integrity-activation-witness",
		"examples/language-syntax-roundtrip/corpus.json",
		"examples/rollback-integrity-activation",
		"internal/meta/languageassurance/rollbackintegrityactivation",
		"internal/meta/languageassurance/digest.go",
		"internal/meta/languageassurance/evaluate_test.go",
		"internal/meta/languageassurance/registry.go",
		"internal/meta/languageassurance/source_authority_activation.go",
		"internal/meta/languagereadiness/languagesyntax",
		"internal/meta/languagereadiness/languagesemantic",
		"examples/language-semantic-model",
		"internal/meta/languagereadiness/toolchainconformance",
		"examples/toolchain-conformance",
		"internal/verify/scope_rollback_integrity_activation.go",
	}
}
