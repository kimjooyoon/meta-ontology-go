package verify

func init() {
	branchScopeAllowlist["agent/gooo-standalone-package-tools-20261007"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		"cmd/gooo/package_execute.go",
		"cmd/gooo/package_execute_observation.go",
		"cmd/gooo/package_execute_standalone_test.go",
		"docs/language-direction.ko.md",
		"docs/language/workspace-manifest.md",
		"examples/assembly-explainer/README.md",
		"examples/assembly-explainer/cases.json",
		"examples/assembly-explainer/gooo.workspace.json",
		"examples/assembly-explainer/inputs.json",
		"examples/assembly-explainer/main.gooo.fixture",
		"internal/bodyexecution/composition_cases.go",
		"internal/bodyexecution/composition_input_separation.go",
		"internal/bodyexecution/composition_observation.go",
		"internal/bodyexecution/composition_observation_test.go",
		"internal/packageruntime/workspaceexecution/execute.go",
		"internal/packageruntime/workspaceexecution/program.go",
		"internal/packageruntime/workspaceexecution/standalone_test.go",
		"internal/verify/scope_gooo_standalone_package_tools_20261007.go",
	}
}
