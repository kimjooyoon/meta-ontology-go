package verify

func init() {
	branchScopeAllowlist["agent/vertical-slice-closure-eligibility"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/transformation-effect.yml",
		"cmd/vertical-slice-closure-eligibility-witness",
		"examples/vertical-slice-closure-eligibility",
		"internal/meta/languageassurance/verticalsliceclosureeligibility",
		"internal/verify/scope_vertical_slice_closure_eligibility.go",
	}
}
