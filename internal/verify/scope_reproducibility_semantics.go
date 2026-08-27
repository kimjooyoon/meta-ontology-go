package verify

func init() {
	branchScopeAllowlist["agent/luna-meta-12-reproducibility-semantics"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		"examples/reproducibility-semantics",
		"internal/meta/reproducibilitysemantics",
		"internal/verify/scope_reproducibility_semantics.go",
		"scripts/reproducibility-semantics",
		"scripts/semantic-conformance.sh",
	}
}
