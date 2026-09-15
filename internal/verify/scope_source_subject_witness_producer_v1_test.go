package verify

import "testing"

func TestSourceSubjectWitnessProducerScope(t *testing.T) {
	paths, ok := BranchScope(sourceSubjectWitnessProducerBranch)
	if !ok || len(paths) != 6 {
		t.Fatalf("source witness producer branch was not registered exactly: known=%t paths=%d", ok, len(paths))
	}
	allowed := []string{
		"scripts/source-subject-witness/indicator_state.go",
		"scripts/source-subject-witness/indicator_state_test.go",
		"internal/verify/scope_source_subject_witness_producer_v1.go",
		".github/ci-governance.json",
	}
	if err := CheckPathScopeForBranch(allowed, sourceSubjectWitnessProducerBranch); err != nil {
		t.Fatalf("representative producer paths were rejected: %v", err)
	}
	if err := CheckPathScopeForBranch([]string{"scripts/source-subject-witness/source.go"}, sourceSubjectWitnessProducerBranch); err == nil {
		t.Fatal("unregistered source witness file was accepted")
	}
	if err := CheckPathScopeForBranch([]string{"scripts/unrelated/main.go"}, sourceSubjectWitnessProducerBranch); err == nil {
		t.Fatal("unrelated path was accepted")
	}
}
