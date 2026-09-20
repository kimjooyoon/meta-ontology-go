package verify

const sourceSubjectWitnessCounterexampleBranch = "agent/source-subject-witness-counterexample-v1"

func init() {
	branchScopeAllowlist[sourceSubjectWitnessCounterexampleBranch] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		"internal/verify/scope_source_subject_witness_counterexample_v1.go",
		"internal/verify/scope_source_subject_witness_counterexample_v1_test.go",
		"scripts/source-subject-witness/build.go",
		"scripts/source-subject-witness/digest.go",
		"scripts/source-subject-witness/indicator_state.go",
		"scripts/source-subject-witness/ledger_validate.go",
		"scripts/source-subject-witness/unsatisfied_indicator_test.go",
	}
}
