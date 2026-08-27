package verify

func init() {
	branchScopeAllowlist["agent/gooo-artifact-lifecycle-receipt"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/self-improvement-language-candidate.yml",
		"bootstrap/function-extractor/recipes.json",
		"examples/self-improvement/transport.gooo",
		"internal/meta/languagereadiness/languagesemantic/registry_definition.go",
		"internal/meta/languagereadiness/languagesyntax/conformance/evaluate_test.go",
		"internal/meta/selfimprovementtransport",
		"internal/verify/scope_gooo_artifact_lifecycle_receipt.go",
		"scripts/self-improvement-transport-selector",
	}
}
