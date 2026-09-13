package verify

func init() {
	branchScopeAllowlist["agent/pr-authoritative-contract-20260911"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/self-improvement-contract.yml",
		"internal/verify/scope_pr_authoritative_contract_20260911.go",
		"scripts/self-improvement-contract/workflow_authority_test.go",
		"scripts/self-improvement-contract/workflow_authority_cases_test.go",
	}
}
