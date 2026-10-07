package verify

func init() {
	branchScopeAllowlist["agent/gooo-assembly-policy-20261008"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		"cmd/gooo/body_compose.go",
		"cmd/gooo/body_compose_policy.go",
		"cmd/gooo/body_compose_policy_test.go",
		"docs/language/body-codegen.md",
		"examples/assembly-policy/README.md",
		"examples/assembly-policy/cases.json",
		"examples/assembly-policy/unsolved.gooo.fixture",
		"internal/bodycodegen/record_assembly.go",
		"internal/bodycodegen/record_assembly_replay.go",
		"internal/bodycodegen/record_assembly_types.go",
		"internal/bodycodegen/record_policy.go",
		"internal/bodycodegen/record_policy_selection_test.go",
		"internal/bodycodegen/record_policy_test.go",
		"internal/bodycodegen/record_policy_types.go",
		"internal/bodyexecution/composition_policy_test.go",
		"internal/bodyexecution/composition_source_fill.go",
		"internal/bodyexecution/composition_source_search.go",
		"internal/verify/scope_gooo_assembly_policy_20261008.go",
	}
}
