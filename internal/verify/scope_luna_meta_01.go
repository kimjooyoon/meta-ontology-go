package verify

func init() {
	branchScopeAllowlist["agent/luna-meta-01-claim-lifecycle-calculus"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/ci.yml",
		".github/workflows/claim-lifecycle-calculus.yml",
		"docs/research/claim-lifecycle-calculus.md",
		"examples/claim-lifecycle-calculus",
		"internal/verify/scope_luna_meta_01.go",
		"scripts/claim-lifecycle-calculus",
		"scripts/claim-lifecycle-calculus-judge",
	}
}
