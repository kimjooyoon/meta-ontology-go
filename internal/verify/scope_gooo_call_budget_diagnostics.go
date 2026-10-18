package verify

func init() {
	branchScopeAllowlist["agent/gooo-call-budget-diagnostics-20261018"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		"cmd/gooo/body_codegen.go",
		"cmd/gooo/body_codegen_budget_test.go",
		"docs/language/body-codegen.md",
		"internal/bodycodegen/pure_call_budget.go",
		"internal/bodycodegen/pure_call_budget_test.go",
		"internal/bodycodegen/pure_calls.go",
		"internal/verify/scope_gooo_call_budget_diagnostics.go",
	}
}
