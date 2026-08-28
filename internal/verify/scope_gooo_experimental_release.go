package verify

func init() {
	branchScopeAllowlist["agent/gooo-experimental-release"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/gooo-release-publish.yml",
		"docs/external/gooo-release-publication-v1.md",
		"examples/gooo-release-publication",
		"internal/meta/languagereadiness/languagesyntax/registry.go",
		"internal/verify/scope_gooo_experimental_release.go",
	}
}
