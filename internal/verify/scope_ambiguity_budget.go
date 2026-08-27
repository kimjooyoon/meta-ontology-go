package verify

func init() {
	branchScopeAllowlist["agent/luna-meta-10-ambiguity-budget"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/ambiguity-budget.yml",
		"cmd/ambiguity-budget-verifier",
		"cmd/ambiguity-budget-witness",
		"docs/language/deterministic-ambiguity-budget.md",
		"examples/ambiguity-budget",
		"examples/language-semantic-model/corpus.json",
		"examples/language-syntax-roundtrip/corpus.json",
		"examples/toolchain-conformance/corpus.json",
		"examples/vertical-slice-closure-shadow",
		"internal/meta/ambiguitybudget",
		"internal/meta/ambiguitybudgetjudge",
		"internal/meta/languageconcept",
		"internal/meta/languagereadiness/languagesemantic",
		"internal/meta/languagereadiness/languagesemanticbinding",
		"internal/meta/languagereadiness/languagesyntax",
		"internal/meta/languagereadiness/toolchainconformance",
		"internal/verify/scope_ambiguity_budget.go",
		"scripts/ambiguity-budget",
	}
}
