package verify

func init() {
	branchScopeAllowlist["agent/vertical-slice-closure-shadow"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/transformation-effect.yml",
		"cmd/vertical-slice-closure-shadow-witness",
		"examples/vertical-slice-closure-shadow",
		"internal/meta/languageassurance/verticalsliceclosureshadow",
		"internal/verify/scope_vertical_slice_closure_shadow.go",
	}
}
