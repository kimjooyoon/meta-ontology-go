package verify

func init() {
	branchScopeAllowlist["agent/toolchain-cross-platform-release"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/transformation-effect.yml",
		"cmd/language-readiness-witness",
		"cmd/toolchain-release-platform-witness",
		"cmd/toolchain-release-witness",
		"docs/language/toolchain-cross-platform-release-readiness.md",
		"docs/language/toolchain-cross-platform-release.md",
		"examples/toolchain-cross-platform-release",
		"internal/meta/languageconcept",
		"internal/meta/languagereadiness/artifact",
		"internal/meta/languagereadiness/external_evidence.go",
		"internal/meta/languagereadiness/toolchain_cross_platform_release_evidence.go",
		"internal/meta/languagereadiness/toolchain_cross_platform_release_promotion.go",
		"internal/meta/languagereadiness/toolchainrelease",
		"internal/verify/scope_toolchain_cross_platform_release.go",
	}
}
