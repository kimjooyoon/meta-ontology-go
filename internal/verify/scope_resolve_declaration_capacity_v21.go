package verify

const resolveDeclarationCapacityV21Branch = "agent/resolve-declaration-capacity-v21"

func init() {
	branchScopeAllowlist[resolveDeclarationCapacityV21Branch] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		"cmd/gooo/compatibility_generate_part01.go",
		"internal/meta/repositoryprojection/extractor/cases_test.go",
		"internal/verify/scope_resolve_declaration_capacity_v21.go",
		"scripts/self-improvement-compiler-compatibility/receipt.go",
	}
}
