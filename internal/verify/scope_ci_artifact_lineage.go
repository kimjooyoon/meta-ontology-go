package verify

func init() {
	branchScopeAllowlist["agent/ci-artifact-lineage"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/ci.yml",
		"examples/ci-artifact-lineage",
		"internal/verify/scope_ci_artifact_lineage.go",
		"scripts/ci-proof/artifacts.js",
		"scripts/ci-proof/artifacts_test.js",
	}
}
