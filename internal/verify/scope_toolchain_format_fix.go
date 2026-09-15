package verify

func init() {
	branchScopeAllowlist["agent/toolchain-format-fix"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/transformation-effect.yml",
		"cmd/gooo",
		"cmd/language-readiness-witness",
		"cmd/toolchain-format-fix-witness",
		"docs/language/toolchain-cli-readiness.md",
		"docs/language/toolchain-format-fix-readiness.md",
		"docs/language/toolchain-format-fix.md",
		"examples/language-concept-catalog/README.md",
		"examples/toolchain-format-fix",
		"internal/formatfix",
		"internal/meta/languageconcept",
		"internal/meta/languagereadiness/artifact",
		"internal/meta/languagereadiness/external_evidence.go",
		"internal/meta/languagereadiness/toolchain_cli_evidence.go",
		"internal/meta/languagereadiness/toolchain_format_fix_evidence.go",
		"internal/meta/languagereadiness/toolchain_format_fix_promotion.go",
		"internal/meta/languagereadiness/toolchain_format_fix_test.go",
		"internal/meta/languagereadiness/toolchaincli",
		"internal/meta/languagereadiness/toolchainformatfix",
		"internal/verify/scope_toolchain_format_fix.go",
	}
}
