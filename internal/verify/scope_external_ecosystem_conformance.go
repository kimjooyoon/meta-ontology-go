package verify

func init() {
	branchScopeAllowlist["agent/external-ecosystem-conformance"] = []string{
		".github/workflows/transformation-effect.yml",
		"cmd/external-ecosystem-conformance-witness",
		"examples/external-ecosystem-conformance",
		"internal/meta/externalecosystemconformance",
		"internal/verify/scope_external_ecosystem_conformance.go",
	}
}
