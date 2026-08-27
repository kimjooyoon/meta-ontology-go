package verify

func init() {
	branchScopeAllowlist["agent/gooo-artifact-lifecycle-receipt"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/self-improvement-language-candidate.yml",
		"examples/self-improvement/transport.gooo",
		"internal/meta/selfimprovementtransport",
		"internal/verify/scope_gooo_artifact_lifecycle_receipt.go",
		"scripts/self-improvement-transport-selector",
	}
}
