package verify

func init() {
	branchScopeAllowlist["agent/artifact-producer-lifecycle-20261018"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/ci.yml",
		"docs/ci-work-domain.ko.md",
		"internal/verify/scope_artifact_producer_lifecycle.go",
		"scripts/ci-proof/artifacts.js",
	}
}
