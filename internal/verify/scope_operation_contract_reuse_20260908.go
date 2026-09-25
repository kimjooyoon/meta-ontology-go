package verify

func init() {
	branchScopeAllowlist["agent/operation-contract-parse-reuse-20260908"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		"internal/verify/scope_operation_contract_reuse_20260908.go",
		"internal/meta/generation/operation_input_contract.go",
		"internal/meta/generation/operation_contract_cache.go",
		"internal/meta/generation/operation_contract_cache_test.go",
	}
}
