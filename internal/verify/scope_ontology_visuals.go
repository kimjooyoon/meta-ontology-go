package verify

func init() {
	branchScopeAllowlist["agent/ontology-visuals"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		"README.md",
		"docs/assets/ontology-visuals",
		"internal/verify/scope_ontology_visuals.go",
	}
}
