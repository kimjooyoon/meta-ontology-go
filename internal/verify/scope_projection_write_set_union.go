package verify

func init() {
	branchScopeAllowlist["agent/projection-write-set-union"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/repository-projection.yml",
		"internal/verify/scope_projection_write_set_union.go",
		"scripts/authorized-write-set",
	}
}
