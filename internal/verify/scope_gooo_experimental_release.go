package verify

func init() {
	branchScopeAllowlist["agent/gooo-experimental-release"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/gooo-release-publish.yml",
		"docs/external/gooo-release-publication-v1.md",
		"examples/gooo-release-publication",
		"examples/language-syntax-roundtrip/corpus.json",
		"internal/meta/languageassurance/verticalsliceclosureshadow/contract.go",
		"internal/meta/languageassurance/verticalsliceclosureshadow/denominator.go",
		"internal/meta/languageassurance/verticalsliceclosureshadow/evidence/denominator.json",
		"internal/meta/languagereadiness/languagesyntax/registry.go",
		"internal/meta/languagereadiness/languagesyntax/conformance/evaluate_test.go",
		"internal/meta/languagereadiness/languagesyntax/model.go",
		"internal/verify/scope_gooo_experimental_release.go",
	}
	branchScopeAllowlist["agent/gooo-release-asset-set"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/gooo-release-publish.yml",
		"internal/verify/scope_gooo_experimental_release.go",
	}
}
