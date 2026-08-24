package verify

func init() {
	branchScopeAllowlist["agent/external-ecosystem-conformance-execution"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/transformation-effect.yml",
		"cmd/external-ecosystem-execution-witness",
		"examples/external-ecosystem-execution",
		"internal/meta/externalecosystemexecution",
		"internal/verify/scope_external_ecosystem_execution.go",
	}
}
