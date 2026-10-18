package verify

func init() {
	branchScopeAllowlist["agent/workspace-input-template-20261018"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		"cmd/gooo/package_inspect.go",
		"cmd/gooo/package_inspect_inputs_test.go",
		"cmd/gooo/usage_package.go",
		"docs/workspace-source-inspection.md",
		"internal/bodyexecution/composition_inspect.go",
		"internal/bodyexecution/composition_inspect_test.go",
		"internal/packageruntime/workspaceexecution/input_template.go",
		"internal/verify/scope_workspace_input_template.go",
	}
}
