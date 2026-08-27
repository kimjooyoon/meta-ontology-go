package verify

func init() {
	branchScopeAllowlist["agent/luna-meta-30-experiment-portfolio"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/experiment-portfolio.yml",
		"cmd/experiment-portfolio-causal-source",
		"cmd/experiment-portfolio-causality",
		"cmd/experiment-portfolio-evaluate",
		"cmd/experiment-portfolio-receipt",
		"docs/language/language-experiment-portfolio.md",
		"docs/language/language-semantic-model.md",
		"docs/language/language-semantic-readiness-binding.md",
		"docs/language/toolchain-conformance.md",
		"examples/experiment-portfolio",
		"examples/language-semantic-model",
		"examples/language-syntax-roundtrip/corpus.json",
		"examples/toolchain-conformance",
		"examples/vertical-slice-closure-shadow",
		"internal/meta/experimentportfolio",
		"internal/meta/languageassurance/verticalsliceclosureshadow",
		"internal/meta/languagereadiness/languagesemantic",
		"internal/meta/languagereadiness/languagesemanticbinding",
		"internal/meta/languagereadiness/languagesyntax",
		"internal/meta/languagereadiness/toolchainconformance",
		"internal/verify/scope_experiment_portfolio.go",
		"scripts/experiment-portfolio",
	}
}
