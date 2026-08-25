package verify

func init() {
	branchScopeAllowlist["agent/feedback-predecessor-baseline"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/ci.yml",
		"cmd/feedback-predecessor-witness",
		"examples/feedback-predecessor-cycle",
		"internal/meta/feedbackpredecessor",
		"internal/verify/scope_feedback_predecessor_baseline.go",
	}
}
