package verify

func init() {
	branchScopeAllowlist["agent/independent-policy-execution-observer-20260920"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		"cmd/meta-policy-compilation-consumer/revision_receipt_observation.go",
		"cmd/meta-policy-compilation-consumer/revision_receipt_wire.go",
		"cmd/meta-policy-compilation-witness/revision_independent_observation.go",
		"cmd/meta-policy-compilation-witness/revision_independent_observation_test.go",
		"cmd/meta-policy-compilation-witness/revision_independent_process.go",
		"cmd/meta-policy-compilation-witness/revision_independent_report.go",
		"docs/language/meta-policy-compilation.md",
		"examples/meta-policy-compilation/counterexample-execution/README.md",
		"internal/meta/policycompilation/revision_independent_execution.go",
		"internal/meta/policycompilation/revision_independent_public_observation_test.go",
		"internal/meta/policycompilation/revision_observation_model.go",
		"internal/verify/scope_independent_policy_execution_observer_20260920.go",
	}
}
