package verify

const sourceSubjectWitnessObservedBranch = "agent/source-subject-witness-observed-v1"

func init() {
	branchScopeAllowlist[sourceSubjectWitnessObservedBranch] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		"internal/verify/scope_source_subject_witness_observed_v1.go",
		"internal/verify/scope_source_subject_witness_observed_v1_test.go",
		"scripts/source-subject-witness/build.go",
		"scripts/source-subject-witness/digest.go",
		"scripts/source-subject-witness/function_binding.go",
		"scripts/source-subject-witness/indicator_state.go",
		"scripts/source-subject-witness/indicator_state_test.go",
		"scripts/source-subject-witness/indicators.go",
		"scripts/source-subject-witness/ledger.go",
		"scripts/source-subject-witness/ledger_validate.go",
		"scripts/source-subject-witness/source.go",
		"scripts/source-subject-witness/validate.go",
	}
}
