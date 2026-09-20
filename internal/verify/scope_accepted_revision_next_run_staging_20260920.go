package verify

func init() {
	branchScopeAllowlist["agent/accepted-revision-next-run-staging-20260920"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		"cmd/gooo/check_part02_test.go",
		"cmd/gooo/main_part01.go",
		"cmd/gooo/stage_accepted_revision.go",
		"cmd/gooo/stage_accepted_revision_test.go",
		"cmd/gooo/usage.go",
		"examples/domain-observation/README.md",
		"examples/domain-observation/observe.sh",
		"internal/valueexecution/source_revision_next_run_staging.go",
		"internal/valueexecution/source_revision_next_run_staging_test.go",
		"internal/verify/scope_accepted_revision_next_run_staging_20260920.go",
	}
}
