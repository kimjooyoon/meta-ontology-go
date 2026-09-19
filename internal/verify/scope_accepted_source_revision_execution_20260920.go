package verify

func init() {
	branchScopeAllowlist["agent/accepted-source-revision-execution-20260920"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		"cmd/gooo/check_part02_test.go",
		"cmd/gooo/emit_dispatch.go",
		"cmd/gooo/main_part01.go",
		"cmd/gooo/run_accepted_revision.go",
		"cmd/gooo/run_accepted_revision_test.go",
		"cmd/gooo/usage.go",
		"examples/domain-observation/README.md",
		"examples/domain-observation/observe.sh",
		"internal/valueexecution/source_revision_execution.go",
		"internal/valueexecution/source_revision_execution_test.go",
		"internal/verify/scope_accepted_source_revision_execution_20260920.go",
	}
}
