package verify

func init() {
	branchScopeAllowlist["agent/language-syntax-corpus-conformance-20260920"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		"docs/language/language-syntax-roundtrip.md",
		"examples/language-syntax-roundtrip/README.md",
		"examples/language-syntax-roundtrip/corpus.json",
		"internal/meta/languagereadiness/languagesyntax/conformance/evaluate_test.go",
		"internal/meta/languagereadiness/languagesyntax/model.go",
		"internal/meta/languagereadiness/languagesyntax/registry.go",
		"internal/verify/scope_language_syntax_corpus_conformance_20260920.go",
	}
}
