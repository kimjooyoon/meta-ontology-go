package verify

func init() {
	branchScopeAllowlist["agent/luna-meta-09-causal-ci-selection"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/causal-ci-selection.yml",
		"docs/causal-ci-selection.md",
		"examples/causal-ci-selection",
		"internal/meta/causalci",
		"internal/verify/scope_luna_meta_09.go",
		"scripts/causal-ci-selection",
	}
}
