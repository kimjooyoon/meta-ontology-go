package verify

func init() {
	branchScopeAllowlist["agent/luna-meta-21-audience-resolution"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/audience-resolution.yml",
		"cmd/audience-resolution-consumer",
		"cmd/audience-resolution-witness",
		"examples/audience-resolution",
		"internal/meta/audienceresolution",
		"internal/meta/audienceresolutionconsumer",
		"internal/verify/scope_audience_resolution.go",
	}
}
