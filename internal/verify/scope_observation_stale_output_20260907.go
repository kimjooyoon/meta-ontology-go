package verify

func init() {
	branchScopeAllowlist["agent/observation-stale-output-20260907"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		"internal/verify/scope_observation_stale_output_20260907.go",
		"scripts/meta-execution/run.go",
		"scripts/meta-execution/observation_archive.go",
		"scripts/meta-execution/observation_archive_test.go",
	}
}
