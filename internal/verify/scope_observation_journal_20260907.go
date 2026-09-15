package verify

func init() {
	branchScopeAllowlist["agent/observation-boundary-journal-20260907"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		"internal/verify/scope_observation_journal_20260907.go",
		"scripts/meta-execution/operations.go",
		"scripts/meta-execution/run.go",
		"scripts/meta-execution/observation_journal.go",
		"scripts/meta-execution/observation_journal_test.go",
	}
}
