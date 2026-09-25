package verify

func init() {
	branchScopeAllowlist["agent/native-test-cost-trace-20260911"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		"internal/verify/scope_native_test_cost_trace_20260911.go",
		"scripts/meta-execution/collapse_execution_test.go",
		"scripts/meta-execution/collapse_execution_trace_test.go",
	}
}
