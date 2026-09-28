package verify

func init() {
	branchScopeAllowlist["agent/accepted-revision-provenance-20260920"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		"examples/domain-observation/README.md",
		"examples/domain-observation/observe.sh",
		"internal/valueexecution/source_revision_execution.go",
		"internal/valueexecution/source_revision_execution_test.go",
		"internal/verify/scope_accepted_revision_provenance_20260920.go",
	}
}
