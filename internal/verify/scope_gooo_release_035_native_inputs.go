package verify

func init() {
	branchScopeAllowlist["agent/gooo-release-035-native-inputs"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/gooo-release-publish.yml",
		"cmd/gooo/package_source_metadata.go",
		"internal/packageruntime/source_metadata.go",
		"internal/meta/languagereadiness/toolchainrelease/platform_package_profile_program.go",
		"cmd/gooo/version.go",
		"cmd/gooo/version_test.go",
		"docs/external/gooo-release-publication-v3.md",
		"docs/releases/0.6.35-dev.md",
		"internal/meta/languagereadiness/toolchainrelease/platform_language.go",
		"internal/meta/languagereadiness/toolchainrelease/platform_package_profile_runtime.go",
		"internal/meta/languagereadiness/toolchainrelease/platform_named_inputs.go",
		"internal/meta/languagereadiness/toolchainrelease/platform_named_inputs_test.go",
		"internal/verify/scope_gooo_release_035_native_inputs.go",
	}
}
