package verify

func init() {
	branchScopeAllowlist["agent/gooo-attestation-summary"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/self-improvement-attestation-resolution.yml",
		"internal/meta/selfimprovementattestation",
		"internal/verify/scope_gooo_attestation_summary.go",
		"scripts/self-improvement-attestation",
	}
}
