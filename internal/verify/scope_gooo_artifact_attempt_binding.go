package verify

func init() {
	branchScopeAllowlist["agent/gooo-artifact-attempt-binding"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/self-improvement-language-candidate.yml",
		"internal/meta/selfimprovementtransport",
		"internal/verify/scope_gooo_artifact_attempt_binding.go",
		"scripts/self-improvement-transport-selector",
	}
}
