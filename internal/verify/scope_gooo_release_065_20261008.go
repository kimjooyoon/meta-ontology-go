package verify

func init() {
	branchScopeAllowlist["agent/gooo-release-065-20261008"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/gooo-release-publish.yml",
		"cmd/gooo/package_construction_proposal_test.go",
		"cmd/gooo/version.go",
		"cmd/gooo/version_test.go",
		"docs/external/gooo-release-publication-v3.md",
		"docs/language-direction.ko.md",
		"docs/research/release-065-dogfood-20261008/deterministic.json",
		"docs/research/release-065-dogfood-20261008/interpretation-before-fix.json",
		"docs/research/release-065-dogfood-20261008/interpretation.json",
		"docs/research/release-065-dogfood-20261008/model.json",
		"docs/research/release-065-dogfood-20261008/replay.json",
		"docs/research/release-065-dogfood-20261008/summary.json",
		"examples/assembly-explainer/README.md",
		"examples/package-diagnostic-replay/README.md",
		"internal/bodyexecution/record_observation.go",
		"internal/bodyexecution/record_proposal_observation_test.go",
		"internal/meta/languagereadiness/toolchaincli/assert_json.go",
		"internal/meta/languagereadiness/toolchaincli/assert_positive.go",
		"internal/meta/languagereadiness/toolchaincli/fixture_output_test.go",
		"internal/meta/languagereadiness/toolchainrelease/fixture_test.go",
		"internal/verify/scope_gooo_release_065_20261008.go",
	}
}
