package verify

func init() {
	branchScopeAllowlist["agent/luna-meta-12-reproducibility-semantics"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		"docs/language/language-semantic-model.md",
		"docs/language/language-semantic-readiness-binding.md",
		"docs/language/language-syntax-roundtrip.md",
		"examples/language-semantic-model",
		"examples/language-syntax-roundtrip",
		"examples/reproducibility-semantics",
		"examples/toolchain-conformance/corpus.json",
		"internal/meta/languagereadiness/languagesemantic",
		"internal/meta/languagereadiness/languagesemanticbinding",
		"internal/meta/languagereadiness/languagesyntax",
		"internal/meta/languagereadiness/toolchainconformance",
		"internal/meta/reproducibilitysemantics",
		"internal/verify/scope_reproducibility_semantics.go",
		"scripts/reproducibility-semantics",
		"scripts/semantic-conformance.sh",
	}
}
