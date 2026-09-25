package verify

const sourceSubjectWitnessLedgerFixBranch = "agent/source-subject-witness-ledger-fix-v1"

func init() {
	branchScopeAllowlist[sourceSubjectWitnessLedgerFixBranch] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		"internal/verify/scope_source_subject_witness_ledger_fix_v1.go",
		"internal/verify/scope_source_subject_witness_ledger_fix_v1_test.go",
		"scripts/source-subject-witness/build.go",
		"scripts/source-subject-witness/counts.go",
		"scripts/source-subject-witness/function_catalog_boundary_test.go",
		"scripts/source-subject-witness/ledger_validate.go",
	}
}
