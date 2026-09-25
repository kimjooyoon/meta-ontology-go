package verify

func init() {
	branchScopeAllowlist["agent/repair-handoff-dogfood-20260920"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/repair-handoff-dogfood.yml",
		"examples/repair-handoff/main.gooo",
		"examples/repair-handoff/repair-candidate.json",
		"examples/repair-handoff/README.md",
		"internal/verify/scope_repair_handoff_dogfood_20260920.go",
	}
}
