package verify

const opentofuExactReuseBranch = "agent/opentofu-exact-reuse-v1"

func init() {
	branchScopeAllowlist[opentofuExactReuseBranch] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		"internal/meta/opentofuobservation/counterexamples.go",
		"internal/meta/opentofuobservation/counterexamples_test.go",
		"internal/meta/opentofuobservation/evaluate.go",
		"internal/meta/opentofuobservation/evaluate_test.go",
		"internal/meta/opentofuobservation/model.go",
		"internal/meta/opentofuobservation/report.go",
		"internal/meta/opentofuobservation/validate.go",
		"internal/meta/opentofuobservation/verify.go",
		"internal/verify/scope_opentofu_exact_reuse_v1.go",
		"internal/verify/scope_opentofu_exact_reuse_v1_test.go",
		"scripts/opentofu-observation",
	}
}
