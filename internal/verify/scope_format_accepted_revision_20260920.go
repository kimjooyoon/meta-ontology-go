package verify

func init() {
	branchScopeAllowlist["agent/format-accepted-revision-20260920"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		"cmd/gooo/run_accepted_revision.go",
		"cmd/gooo/run_accepted_revision_test.go",
		"internal/verify/scope_format_accepted_revision_20260920.go",
	}
}
