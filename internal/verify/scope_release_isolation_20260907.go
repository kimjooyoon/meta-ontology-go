package verify

func init() {
	branchScopeAllowlist["agent/release-artifact-execution-isolation-20260907"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/gooo-release-publish.yml",
		"internal/verify/scope_release_isolation_20260907.go",
		"internal/verify/scope_release_isolation_20260907_test.go",
	}
}
