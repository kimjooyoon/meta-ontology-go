package verify

import "testing"

func TestSourceSubjectWitnessLedgerFixScope(t *testing.T) {
	paths, ok := BranchScope(sourceSubjectWitnessLedgerFixBranch)
	if !ok || len(paths) != 8 {
		t.Fatalf("source witness ledger fix branch was not registered exactly: known=%t paths=%d", ok, len(paths))
	}
	allowed := []string{
		"scripts/source-subject-witness/build.go",
		"scripts/source-subject-witness/counts.go",
		"scripts/source-subject-witness/ledger_validate.go",
		"internal/verify/scope_source_subject_witness_ledger_fix_v1.go",
	}
	if err := CheckPathScopeForBranch(allowed, sourceSubjectWitnessLedgerFixBranch); err != nil {
		t.Fatalf("source witness ledger paths were rejected: %v", err)
	}
	if err := CheckPathScopeForBranch([]string{"scripts/source-subject-witness/source.go"}, sourceSubjectWitnessLedgerFixBranch); err == nil {
		t.Fatal("unregistered source witness file was accepted")
	}
}
