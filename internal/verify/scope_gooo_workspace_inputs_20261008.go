package verify

func init() {
	branchScopeAllowlist["agent/gooo-workspace-inputs-20261008"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		"cmd/gooo/body_compose_separation_test.go",
		"docs/native-body-composition.md",
		"docs/research/workspace-inputs-20261008/execution.json",
		"docs/research/workspace-inputs-20261008/summary.json",
		"examples/workspace-input-observations/README.md",
		"examples/workspace-input-observations/cases.json",
		"examples/workspace-input-observations/gooo.workspace.json",
		"examples/workspace-input-observations/main.gooo.fixture",
		"internal/bodyexecution/composition_earlier_inputs.go",
		"internal/bodyexecution/composition_input_separation.go",
		"internal/packageruntime/workspaceexecution/execute.go",
		"internal/packageruntime/workspaceexecution/input_separation.go",
		"internal/packageruntime/workspaceexecution/input_separation_profiles_test.go",
		"internal/packageruntime/workspaceexecution/input_separation_test.go",
		"internal/verify/scope_gooo_workspace_inputs_20261008.go",
	}
}
