package verify

func init() {
	branchScopeAllowlist["agent/gooo-body-profile-discovery-20261018"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/ci.yml",
		"cmd/gooo/version_body_limits_test.go",
		"cmd/gooo/version_build.go",
		"docs/ci-work-domain.ko.md",
		"docs/native-record-values.md",
		"internal/bodycodegen/profile_limits.go",
		"internal/bodycodegen/record_types.go",
		"internal/verify/scope_gooo_body_profile_discovery.go",
	}
}
