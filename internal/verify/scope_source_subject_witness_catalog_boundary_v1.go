package verify

const sourceSubjectWitnessCatalogBoundaryBranch = "agent/source-subject-witness-catalog-boundary-v2"

func init() {
	branchScopeAllowlist[sourceSubjectWitnessCatalogBoundaryBranch] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		"internal/verify/scope_source_subject_witness_catalog_boundary_v1.go",
		"internal/verify/scope_source_subject_witness_catalog_boundary_v1_test.go",
		"scripts/source-subject-witness/function_binding.go",
		"scripts/source-subject-witness/function_catalog_boundary_test.go",
	}
}
