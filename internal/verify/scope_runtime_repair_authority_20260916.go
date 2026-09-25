package verify

func init() {
	branchScopeAllowlist["agent/runtime-repair-candidate-validation-20260915"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		"cmd/gooo/emit_dispatch.go",
		"docs/governance/problem-solving-scope.md",
		"internal/valueexecution/repair.go",
		"internal/valueexecution/repair_test.go",
		"internal/verify/scope_runtime_repair_authority_20260916.go",
	}
}
