package verify

const devMainTopologyRepair20260902Branch = "agent/dev-main-topology-repair-20260902"

func init() {
	branchScopeAllowlist[devMainTopologyRepair20260902Branch] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		"internal/verify/scope_dev_main_topology_repair_20260902.go",
	}
	branchScopeAllowlist["agent/runtime-repair-candidate-validation-20260915"] = []string{
		"cmd/gooo",
		"internal/valueexecution",
	}
}
