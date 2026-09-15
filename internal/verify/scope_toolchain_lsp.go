package verify

func init() {
	branchScopeAllowlist["agent/toolchain-lsp"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/transformation-effect.yml",
		"cmd/language-readiness-witness",
		"cmd/toolchain-lsp-witness",
		"docs/language/toolchain-lsp-readiness.md",
		"docs/language/toolchain-lsp.md",
		"examples/toolchain-lsp",
		"internal/meta/languageconcept",
		"internal/meta/languagereadiness/artifact",
		"internal/meta/languagereadiness/external_evidence.go",
		"internal/meta/languagereadiness/toolchain_lsp_evidence.go",
		"internal/meta/languagereadiness/toolchain_lsp_promotion.go",
		"internal/meta/languagereadiness/toolchainlsp",
		"internal/verify/scope_toolchain_lsp.go",
	}
}
