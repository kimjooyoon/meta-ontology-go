package verify

func init() {
	branchScopeAllowlist["agent/toolchain-cli"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/transformation-effect.yml",
		"cmd/language-readiness-witness",
		"cmd/toolchain-cli-witness",
		"docs/language/toolchain-cli-readiness.md",
		"docs/language/toolchain-cli.md",
		"examples/toolchain-cli",
		"internal/toolchaincli",
		"internal/meta/languageconcept",
		"internal/meta/languagereadiness/artifact",
		"internal/meta/languagereadiness/external_evidence.go",
		"internal/meta/languagereadiness/promotion.go",
		"internal/meta/languagereadiness/promotion_evidence.go",
		"internal/meta/languagereadiness/promotion_language_validation.go",
		"internal/meta/languagereadiness/promotion_validation.go",
		"internal/meta/languagereadiness/toolchain_cli_evidence.go",
		"internal/meta/languagereadiness/toolchain_cli_promotion.go",
		"internal/meta/languagereadiness/toolchain_cli_test.go",
		"internal/meta/languagereadiness/toolchaincli",
		"internal/verify/scope_toolchain_cli.go",
	}
}
