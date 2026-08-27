package verify

func init() {
	branchScopeAllowlist["agent/luna-meta-03-meta-operation-provenance"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/meta-operation-provenance.yml",
		"examples/meta-operation-provenance",
		"internal/meta/operationprovenance",
		"internal/verify/scope_meta_operation_provenance.go",
		"scripts/meta-operation-provenance",
	}
}
