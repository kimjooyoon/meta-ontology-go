package verify

func init() {
	branchScopeAllowlist["agent/structured-input-diagnostics-20261018"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		"cmd/gooo/package_execute.go",
		"cmd/gooo/package_execute_input_diagnostic_test.go",
		"cmd/gooo/package_replay.go",
		"docs/workspace-source-inspection.md",
		"internal/bodyexecution/composition_case_display.go",
		"internal/bodyexecution/composition_cases.go",
		"internal/bodyexecution/input_diagnostic.go",
		"internal/packageruntime/workspaceexecution/execute.go",
		"internal/packageruntime/workspaceexecution/input_diagnostic_test.go",
		"internal/verify/scope_structured_input_diagnostics.go",
	}
}
