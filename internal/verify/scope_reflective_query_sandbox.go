package verify

func init() {
	branchScopeAllowlist["agent/luna-meta-17-reflective-query-sandbox"] = []string{
		".github/agent-scope-table.md",
		".github/workflows/reflective-query-sandbox.yml",
		"examples/reflective-query-sandbox",
		"internal/meta/reflectivequerysandbox",
		"internal/verify/scope_reflective_query_sandbox.go",
		"scripts/reflective-query-sandbox",
	}
}
