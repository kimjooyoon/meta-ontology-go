package verify

func init() {
	branchScopeAllowlist["agent/luna-meta-19-invariant-transformation"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/transformation-effect.yml",
		"cmd/invariant-transformation-witness",
		"docs/language/invariant-preserving-transformation.md",
		"examples/invariant-transformation",
		"internal/meta/invarianttransformation",
		"internal/verify/scope_invariant_transformation.go",
	}
}
