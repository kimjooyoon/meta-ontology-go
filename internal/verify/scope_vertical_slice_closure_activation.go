package verify

func init() {
	branchScopeAllowlist["agent/vertical-slice-closure-activation"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/self-improvement-contract.yml",
		"cmd/vertical-slice-closure-activation-witness",
		"docs/language/language-semantic-model.md",
		"docs/language/language-semantic-readiness-binding.md",
		"docs/language/language-syntax-roundtrip.md",
		"examples/language-semantic-model",
		"examples/language-syntax-roundtrip/README.md",
		"examples/language-syntax-roundtrip/corpus.json",
		"examples/vertical-slice-closure-activation",
		"internal/meta/languageassurance/digest.go",
		"internal/meta/languageassurance/evaluate_test.go",
		"internal/meta/languageassurance/registry.go",
		"internal/meta/languageassurance/source_authority_activation.go",
		"internal/meta/languageassurance/verticalsliceclosureactivation",
		"internal/meta/languagereadiness/languagesemantic",
		"internal/meta/languagereadiness/languagesemanticbinding",
		"internal/meta/languagereadiness/languagesyntax",
		"internal/verify/scope_vertical_slice_closure_activation.go",
	}
}
