package verify

func init() {
	branchScopeAllowlist["agent/luna-meta-31-conflict-free-registry-projection"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/conflict-free-registry-projection.yml",
		"docs/language/conflict-free-registry-projection.md",
		"examples/language-semantic-model/concept.manifest.json",
		"examples/language-syntax-roundtrip/concept.manifest.json",
		"examples/toolchain-conformance/concept.manifest.json",
		"internal/meta/registryprojection",
		"internal/verify/scope_conflict_free_registry_projection.go",
		"scripts/conflict-free-registry-projection",
		"scripts/conflict-free-registry-projection-consumer",
	}
}
