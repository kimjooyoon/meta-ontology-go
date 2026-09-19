package verify

func init() {
	branchScopeAllowlist["agent/runtime-plan-contract-repair-20260920"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		"cmd/gooo/run_source_plan_test.go",
		"cmd/gooo/runtime_plan.go",
		"cmd/gooo/runtime_plan_execution_contract.go",
		"internal/verify/scope_runtime_plan_contract_repair_20260920.go",
	}
}
