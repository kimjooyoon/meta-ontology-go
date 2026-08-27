package verify

func init() {
	branchScopeAllowlist["agent/luna-meta-13-observer-effect-ledger"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/observer-effect-ledger.yml",
		"docs/research/observer-effect-ledger.md",
		"examples/observer-effect-ledger",
		"internal/meta/observereffect",
		"internal/verify/scope_observer_effect_ledger.go",
		"scripts/observer-effect-judge",
		"scripts/observer-effect-ledger",
	}
}
