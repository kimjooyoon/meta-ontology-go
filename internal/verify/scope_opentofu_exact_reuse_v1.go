package verify

const opentofuExactReuseBranch = "agent/opentofu-exact-reuse-v1"

func init() {
	branchScopeAllowlist[opentofuExactReuseBranch] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		"internal/meta/opentofuobservation",
		"internal/verify/scope_opentofu_exact_reuse_v1.go",
		"internal/verify/scope_opentofu_exact_reuse_v1_test.go",
		"scripts/opentofu-observation",
	}
}
