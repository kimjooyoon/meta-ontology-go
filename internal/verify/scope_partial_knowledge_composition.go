package verify

func init() {
	branchScopeAllowlist["agent/luna-meta-14-partial-knowledge-composition"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		"cmd/partial-knowledge-composition-verifier",
		"cmd/partial-knowledge-composition-witness",
		"docs/research/partial-knowledge-composition.md",
		"examples/partial-knowledge-composition",
		"internal/meta/partialknowledgecomposition",
		"internal/verify/scope_partial_knowledge_composition.go",
	}
}
