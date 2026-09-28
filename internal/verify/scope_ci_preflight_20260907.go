package verify

func init() {
	branchScopeAllowlist["agent/ci-scope-preflight-20260907"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/ci.yml",
		"internal/verify/ci_scope_preflight_order_test.go",
		"internal/verify/scope_ci_preflight_20260907.go",
		"internal/verify/scope_ci_preflight_20260907_test.go",
	}
}
