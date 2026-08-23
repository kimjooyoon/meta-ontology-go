package verify

func init() {
	branchScopeAllowlist["agent/promotion-authorized-continuity"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/transformation-effect.yml",
		"cmd/promotion-authorized-continuity",
		"docs/language/promotion-authorized-continuity.md",
		"internal/meta/languagereadiness/promotioncontinuity",
		"internal/verify/scope_promotion_authorized_continuity.go",
	}
}
