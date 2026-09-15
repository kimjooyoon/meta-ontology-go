package verify

func init() {
	branchScopeAllowlist["agent/workgraph-cli-contract"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/gooo-release-readiness.yml",
		"docs/external/gooo-release-evidence-v1.md",
		"internal/verify/scope_gooo_release_readiness.go",
	}
}
