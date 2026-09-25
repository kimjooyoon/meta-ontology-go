package verify

func init() {
	branchScopeAllowlist["agent/repair-handoff-consumption-20260920"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		"cmd/gooo/check_part02_test.go",
		"cmd/gooo/consume_repair.go",
		"cmd/gooo/consume_repair_test.go",
		"cmd/gooo/emit_dispatch.go",
		"cmd/gooo/main_part01.go",
		"cmd/gooo/usage.go",
		"internal/valueexecution/repair_handoff.go",
		"internal/valueexecution/repair_handoff_test.go",
		"internal/verify/scope_repair_handoff_20260920.go",
	}
}
