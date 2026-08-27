package verify

func init() {
	branchScopeAllowlist["agent/luna-meta-22-claim-dependency-causality"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/language-claim-dependency.yml",
		"docs/research/claim-dependency-causality.md",
		"examples/language-claim-dependency",
		"internal/meta/claimdependency",
		"internal/meta/claimdependencyjudge",
		"internal/verify/scope_gooo_claim_dependency.go",
		"scripts/language-claim-dependency",
		"scripts/language-claim-dependency-evidence",
		"scripts/language-claim-dependency-judge",
		"scripts/language-claim-dependency-intervention",
	}
}
