package verify

func init() {
	branchScopeAllowlist["agent/gooo-release-067-20261008"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/gooo-release-publish.yml",
		"README.md",
		"cmd/gooo/version.go",
		"cmd/gooo/version_test.go",
		"docs/external/gooo-release-publication-v3.md",
		"docs/language-direction.ko.md",
		"docs/releases/0.6.7-dev.md",
		"internal/meta/languagereadiness/toolchaincli/assert_json.go",
		"internal/meta/languagereadiness/toolchaincli/assert_positive.go",
		"internal/meta/languagereadiness/toolchaincli/fixture_output_test.go",
		"internal/meta/languagereadiness/toolchainrelease/fixture_test.go",
		"internal/verify/scope_gooo_release_067_20261008.go",
	}
}
