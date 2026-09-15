package verify

func init() {
	branchScopeAllowlist["agent/gooo-attestation-resolution"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/self-improvement-attestation-resolution.yml",
		"internal/meta/selfimprovementattestation",
		"internal/verify/scope_gooo_attestation_resolution.go",
		"scripts/self-improvement-attestation",
	}
}
