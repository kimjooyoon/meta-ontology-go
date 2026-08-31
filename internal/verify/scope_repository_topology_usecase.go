package verify

func init() {
	branchScopeAllowlist["agent/repository-topology-usecase"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/metric-counterfactual.yml",
		"cmd/repository-topology-witness",
		"docs/language/repository-topology-use-case.md",
		"examples/repository-topology-use-case",
		"internal/meta/repositorytopology",
		"internal/verify/scope_repository_topology_usecase.go",
	}
}
