package verify

func init() {
	branchScopeAllowlist["agent/luna-meta-21-audience-resolution"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		"cmd/audience-resolution-witness",
		"examples/audience-resolution",
		"internal/meta/audienceresolution",
		"internal/verify/scope_audience_resolution.go",
	}
}
