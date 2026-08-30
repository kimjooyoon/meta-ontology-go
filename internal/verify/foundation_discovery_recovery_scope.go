package verify

func init() {
	branchScopeAllowlist["agent/foundation-discovery-recovery-20260830"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/ci.yml",
		"cmd/feedback-predecessor-witness",
		"internal/detection/linecaps",
		"internal/meta/feedbackpredecessor",
		"internal/meta/generation",
		"internal/meta/metabinding/registry.go",
		"internal/meta/sourcepolicy",
		"internal/verify/foundation_discovery_recovery_scope.go",
		"internal/verify/foundation_promotion.go",
		"scripts/ci-proof",
		"scripts/feedback-predecessor-ci",
		"scripts/verify",
	}
}
