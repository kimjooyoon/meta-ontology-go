package verify

func init() {
	paths := []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/transformation-effect.yml",
		"cmd/external-capability-execution-witness",
		"examples/external-capability-execution",
		"internal/meta/externalcapabilityexecution",
		"internal/verify/scope_external_capability_execution.go",
	}
	branches := []string{"agent/external-capability-execution", "agent/external-assurance-eligibility",
		"agent/gooo-capability-boundary"}
	for _, branch := range branches {
		branchScopeAllowlist[branch] = paths
	}
}
