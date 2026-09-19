package verify

func init() {
	branchScopeAllowlist["agent/ci-gofmt-baseline-repair-20260920"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		"internal/valueexecution/source_revision_next_run.go",
		"internal/verify/scope_ci_gofmt_baseline_repair_20260920.go",
	}
}
