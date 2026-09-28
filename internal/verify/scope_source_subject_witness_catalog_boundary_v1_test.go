package verify

import "testing"

func TestSourceSubjectWitnessCatalogBoundaryScope(t *testing.T) {
	paths, ok := BranchScope(sourceSubjectWitnessCatalogBoundaryBranch)
	if !ok || len(paths) != 6 {
		t.Fatalf("source witness catalog boundary branch was not registered exactly: known=%t paths=%d", ok, len(paths))
	}
	allowed := []string{
		"scripts/source-subject-witness/function_binding.go",
		"scripts/source-subject-witness/function_catalog_boundary_test.go",
		"internal/verify/scope_source_subject_witness_catalog_boundary_v1.go",
	}
	if err := CheckPathScopeForBranch(allowed, sourceSubjectWitnessCatalogBoundaryBranch); err != nil {
		t.Fatalf("representative source witness paths were rejected: %v", err)
	}
	if err := CheckPathScopeForBranch([]string{"scripts/source-subject-witness/source.go"}, sourceSubjectWitnessCatalogBoundaryBranch); err == nil {
		t.Fatal("unregistered source witness file was accepted")
	}
}
