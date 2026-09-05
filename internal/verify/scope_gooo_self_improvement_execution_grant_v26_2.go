package verify

func init() {
	branchScopeAllowlist["agent/v26.2-exact-canonical-grant-fixture-20260905"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/self-improvement-candidate-authorization.yml",
		".github/workflows/self-improvement-execution-contract.yml",
		".github/workflows/self-improvement-execution-grant.yml",
		"examples/self-improvement-execution-grant",
		"internal/meta/selfimprovementexecutiongrant",
		"internal/verify/scope_gooo_self_improvement_execution_grant_v26_2.go",
		"scripts/self-improvement-execution-grant",
	}
}
