package verify

func init() {
	branchScopeAllowlist["agent/gooo-body-profile-discovery-20261018"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/ci.yml",
		".github/workflows/gooo-release-publish.yml",
		"cmd/gooo/version.go",
		"cmd/gooo/version_body_limits_test.go",
		"cmd/gooo/version_build.go",
		"cmd/gooo/version_test.go",
		"docs/ci-work-domain.ko.md",
		"docs/external/gooo-release-publication-v3.md",
		"docs/native-record-values.md",
		"docs/releases/0.6.36-dev.md",
		"internal/bodycodegen/profile_limits.go",
		"internal/bodycodegen/pure_call_definition_limit_test.go",
		"internal/bodycodegen/pure_calls.go",
		"internal/bodycodegen/record_types.go",
		"internal/meta/languagereadiness/toolchainrelease/platform_collection_map_reference.go",
		"internal/meta/languagereadiness/toolchainrelease/platform_collection_map_reference_test.go",
		"internal/meta/languagereadiness/toolchainrelease/platform_package_profile.go",
		"internal/meta/languagereadiness/toolchainrelease/platform_package_profile_reference.go",
		"internal/verify/scope_gooo_body_profile_discovery.go",
		"scripts/ci-effort-observation/operation_guard_test.go",
		"scripts/ci-effort-observation/validate.go",
	}
}
