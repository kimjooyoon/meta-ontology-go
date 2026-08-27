package verify

func init() {
	branchScopeAllowlist["agent/luna-meta-29-minimal-causal-explanation"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/metric-counterfactual.yml",
		"examples/minimal-causal-explanation",
		"internal/meta/minimalcausalexplanation",
		"internal/verify/scope_minimal_causal_explanation.go",
		"scripts/minimal-causal-explanation",
	}
}
