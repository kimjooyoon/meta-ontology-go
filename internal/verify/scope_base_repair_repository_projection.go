package verify

func init() {
	branchScopeAllowlist["agent/base-repair-repository-projection"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/repository-projection.yml",
		"bootstrap/function-extractor",
		"examples/repository-projection-repair",
		"internal/meta/repositoryprojection/extractor",
		"internal/verify/scope_base_repair_repository_projection.go",
	}
}
