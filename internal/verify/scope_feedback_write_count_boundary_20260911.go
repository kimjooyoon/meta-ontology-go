package verify

func init() {
	branchScopeAllowlist["agent/feedback-write-count-boundary-20260911"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		"cmd/feedback-semantic-state-witness",
		"internal/meta/feedbackstate",
		"internal/verify/scope_feedback_write_count_boundary_20260911.go",
	}
}
