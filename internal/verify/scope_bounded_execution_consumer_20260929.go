package verify

func init() {
	branchScopeAllowlist["agent/bounded-grant-executor-v1-20260929"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/self-improvement-execution-consumer.yml",
		"examples/self-improvement-execution-consumer",
		"internal/verify/scope_bounded_execution_consumer_20260929.go",
		"scripts/self-improvement-execution-consumer",
	}
}
