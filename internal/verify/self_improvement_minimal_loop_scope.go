package verify

func init() {
	branchScopeAllowlist["agent/self-improvement-minimal-loop"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/self-improvement-minimal-loop.yml",
		"examples/self-improvement-minimal-loop",
		"internal/meta/selfimprovementloop",
		"internal/verify/self_improvement_minimal_loop_scope.go",
		"scripts/self-improvement-minimal-loop",
	}
}
