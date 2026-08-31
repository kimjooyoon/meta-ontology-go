package verify

const sourceSubjectWitnessProducerBranch = "agent/source-subject-witness-producer-v1"

func init() {
	branchScopeAllowlist[sourceSubjectWitnessProducerBranch] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		"internal/verify/scope_source_subject_witness_producer_v1.go",
		"internal/verify/scope_source_subject_witness_producer_v1_test.go",
		"scripts/source-subject-witness/indicator_state.go",
		"scripts/source-subject-witness/indicator_state_test.go",
	}
}
