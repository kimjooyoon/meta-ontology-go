package verify

func init() {
	branchScopeAllowlist["agent/journal-process-interruption-20260908"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		"internal/verify/scope_journal_process_interruption_20260908.go",
		"scripts/meta-execution/observation_journal_process_test.go",
	}
}
