package verify

func init() {
	branchScopeAllowlist["agent/core-rss-observability-v1"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/transformation-effect.yml",
		"cmd/toolchain-cli-witness",
		"docs/language/toolchain-cli.md",
		"internal/meta/languagereadiness/toolchaincli",
		"internal/toolchaincli",
		"internal/verify/scope_core_rss_observability.go",
	}
}
