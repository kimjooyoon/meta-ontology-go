package verify

func init() {
	branchScopeAllowlist["agent/revise-from-handoff-20260920"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		"cmd/gooo/check_part02_test.go",
		"cmd/gooo/emit_dispatch.go",
		"cmd/gooo/main_part01.go",
		"cmd/gooo/revise_from_handoff.go",
		"cmd/gooo/revise_from_handoff_test.go",
		"cmd/gooo/usage.go",
		"internal/valueexecution/repair_source_revision.go",
		"internal/valueexecution/repair_source_revision_test.go",
		"internal/valueexecution/source_revision.go",
		"internal/verify/scope_revise_from_handoff_20260920.go",
	}
}
