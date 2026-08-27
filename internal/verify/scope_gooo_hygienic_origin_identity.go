package verify

func init() {
	branchScopeAllowlist["agent/luna-meta-06-hygienic-origin-identity"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/language-hygienic-origin-identity.yml",
		"examples/language-hygienic-origin-identity",
		"internal/meta/hygienicoriginidentity",
		"internal/verify/scope_gooo_hygienic_origin_identity.go",
		"scripts/language-hygienic-origin-identity",
	}
}
