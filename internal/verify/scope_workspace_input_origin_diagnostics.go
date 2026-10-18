package verify

func init() {
	branchScopeAllowlist["agent/workspace-input-origin-diagnostics-20261018"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		"cmd/gooo/package_execute_input_diagnostic_test.go",
		"docs/workspace-source-inspection.md",
		"internal/bodyexecution/composition_case_display.go",
		"internal/bodyexecution/composition_cases.go",
		"internal/packageruntime/workspaceexecution/execute.go",
		"internal/packageruntime/workspaceexecution/input_diagnostic_test.go",
		"internal/verify/scope_workspace_input_origin_diagnostics.go",
	}
}
