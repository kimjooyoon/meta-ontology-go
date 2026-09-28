package verify

import "testing"

func TestSourceSubjectWitnessCounterexampleScope(t *testing.T) {
	paths, ok := BranchScope(sourceSubjectWitnessCounterexampleBranch)
	if !ok || len(paths) != 9 {
		t.Fatalf("source witness counterexample branch was not registered exactly: known=%t paths=%d", ok, len(paths))
	}
	allowed := []string{
		"scripts/source-subject-witness/indicator_state.go",
		"scripts/source-subject-witness/ledger_validate.go",
		"internal/verify/scope_source_subject_witness_counterexample_v1.go",
	}
	if err := CheckPathScopeForBranch(allowed, sourceSubjectWitnessCounterexampleBranch); err != nil {
		t.Fatalf("representative source witness paths were rejected: %v", err)
	}
	if err := CheckPathScopeForBranch([]string{"scripts/source-subject-witness/source.go"}, sourceSubjectWitnessCounterexampleBranch); err == nil {
		t.Fatal("unregistered source witness file was accepted")
	}
}
