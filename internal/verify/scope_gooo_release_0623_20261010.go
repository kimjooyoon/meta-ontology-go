package verify

func init() {
	branchScopeAllowlist["agent/gooo-release-0623-20261010"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/gooo-release-publish.yml",
		"README.md",
		"cmd/gooo/version.go",
		"cmd/gooo/version_test.go",
		"docs/external/gooo-release-publication-v3.md",
		"docs/native-body-composition.md",
		"docs/composition-plan-inspection.md",
		"docs/releases/0.6.23-dev.md",
		"docs/research/usability-release-20261010.md",
		"examples/composition-inputs/README.md",
		"internal/meta/languagereadiness/toolchainrelease/fixture_test.go",
		"internal/meta/languagereadiness/toolchainrelease/platform_language.go",
		"internal/meta/languagereadiness/toolchainrelease/platform_plan_inputs.go",
		"internal/meta/languagereadiness/toolchainrelease/platform_plan_inspection.go",
		"internal/meta/languagereadiness/toolchainrelease/platform_plan_input_execution.go",
		"internal/meta/languagereadiness/toolchainrelease/platform_plan_input_scoring.go",
		"internal/meta/languagereadiness/toolchainrelease/platform_plan_inputs_test.go",
		"internal/meta/languagereadiness/toolchainrelease/platform_plan_input_fixtures_test.go",
		"internal/meta/languagereadiness/toolchainrelease/testdata",
		"internal/verify/scope_gooo_release_0623_20261010.go",
	}
}
