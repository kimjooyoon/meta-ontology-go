package verify

func init() {
	branchScopeAllowlist["agent/core-conformance-registry-20260905"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/core-conformance-registry.yml",
		".github/workflows/transformation-effect.yml",
		"examples/language-syntax-roundtrip/corpus.json",
		"internal/meta/languagereadiness/languagesyntax",
		"internal/meta/languageassurance/verticalsliceclosureshadow",
		"internal/verify/scope_gooo_core_conformance_registry.go",
	}
}
