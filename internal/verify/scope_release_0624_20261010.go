package verify

func init() {
	branchScopeAllowlist["agent/release-0624-20261010"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/gooo-release-publish.yml",
		"README.md",
		"cmd/gooo/version.go",
		"cmd/gooo/version_test.go",
		"docs/external/gooo-release-publication-v3.md",
		"docs/releases/0.6.24-dev.md",
		"docs/research/typed-path-release-20261010.md",
		"internal/meta/languagereadiness/toolchainrelease/fixture_test.go",
		"internal/meta/languagereadiness/toolchainrelease/platform_language.go",
		"internal/meta/languagereadiness/toolchainrelease/platform_scalar_preflight_model.go",
		"internal/meta/languagereadiness/toolchainrelease/platform_typed_paths.go",
		"internal/meta/languagereadiness/toolchainrelease/platform_typed_paths_test.go",
		"internal/meta/languagereadiness/toolchainrelease/platform_typed_preflight.go",
		"internal/meta/languagereadiness/toolchainrelease/platform_typed_validate.go",
		"internal/meta/languagereadiness/toolchainrelease/platform_typed_model_binding.go",
		"internal/meta/languagereadiness/toolchainrelease/platform_typed_runtime.go",
		"internal/meta/languagereadiness/toolchainrelease/testdata/typed-path-unary.json.gz",
		"internal/meta/languagereadiness/toolchainrelease/testdata/typed-path-record.json.gz",
		"internal/meta/languagereadiness/toolchainrelease/testdata/typed-path-construct.json.gz",
		"internal/meta/languagereadiness/toolchainrelease/testdata/typed-path-replay.json.gz",
		"internal/meta/languagereadiness/toolchainrelease/testdata/typed-path-README.md",
		"internal/verify/scope_release_0624_20261010.go",
	}
}
