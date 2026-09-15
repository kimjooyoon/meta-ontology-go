package verify

func init() {
	branchScopeAllowlist["agent/boundary-unknown-records-20260908"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		"internal/verify/scope_boundary_unknown_20260908.go",
		"scripts/meta-cost-report/model.go",
		"scripts/meta-cost-report/assemble.go",
		"scripts/meta-cost-report/unknown.go",
		"scripts/meta-cost-report/unknown_test.go",
	}
}
