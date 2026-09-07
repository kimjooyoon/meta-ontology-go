package verify

func init() {
	branchScopeAllowlist["agent/manifest-observation-ownership-20260907"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		"internal/meta/generation/observation_path.go",
		"internal/meta/generation/observation_path_test.go",
		"internal/verify/scope_manifest_observation_20260907.go",
		"scripts/meta-execution/run.go",
		"scripts/meta-receipts/run.go",
	}
}
