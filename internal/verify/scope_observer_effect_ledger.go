package verify

func init() {
	branchScopeAllowlist["agent/luna-meta-13-observer-effect-ledger"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/metric-transition.yml",
		".github/workflows/observer-effect-ledger.yml",
		".github/workflows/self-improvement-cycle.yml",
		".github/workflows/self-improvement-language-observation.yml",
		".github/workflows/source-subject-witness.yml",
		".github/workflows/transformation-effect.yml",
		"docs/research/observer-effect-ledger.md",
		"examples/language-semantic-model",
		"examples/language-syntax-roundtrip",
		"examples/observer-effect-ledger",
		"internal/meta/languagereadiness/languagesemantic",
		"internal/meta/languagereadiness/languagesyntax",
		"internal/meta/observereffect",
		"internal/verify/scope_observer_effect_ledger.go",
		"scripts/observer-effect-judge",
		"scripts/observer-effect-ledger",
	}
}
