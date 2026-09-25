package verify

import "testing"

func TestSourceSubjectWitnessObservedScope(t *testing.T) {
	paths, ok := BranchScope(sourceSubjectWitnessObservedBranch)
	if !ok || len(paths) != 14 {
		t.Fatalf("source witness branch was not registered exactly: known=%t paths=%d", ok, len(paths))
	}
	allowed := []string{
		"scripts/source-subject-witness/source.go",
		"scripts/source-subject-witness/indicator_state.go",
		"internal/verify/scope_source_subject_witness_observed_v1.go",
		".github/ci-governance.json",
	}
	if err := CheckPathScopeForBranch(allowed, sourceSubjectWitnessObservedBranch); err != nil {
		t.Fatalf("representative source witness paths were rejected: %v", err)
	}
	if err := CheckPathScopeForBranch([]string{"scripts/source-subject-witness/other.go"}, sourceSubjectWitnessObservedBranch); err == nil {
		t.Fatal("unregistered source witness file was accepted")
	}
	if err := CheckPathScopeForBranch([]string{"scripts/unrelated/main.go"}, sourceSubjectWitnessObservedBranch); err == nil {
		t.Fatal("unrelated path was accepted")
	}
}
