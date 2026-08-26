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
	for _, branch := range []string{"agent/external-capability-execution", "agent/external-assurance-eligibility"} {
		branchScopeAllowlist[branch] = paths
	}
}
