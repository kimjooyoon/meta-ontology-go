package verify

func init() {
	branchScopeAllowlist["agent/gooo-producer-attestation"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/self-improvement-language-observation.yml",
		"internal/verify/scope_gooo_producer_attestation.go",
	}
}
