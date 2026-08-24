package verify

func init() {
	branchScopeAllowlist["agent/source-authority-operating"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/self-improvement-contract.yml",
		"cmd/source-authority-activation-witness",
		"docs/language/language-assurance-kernel.md",
		"examples/source-authority-activation",
		"internal/meta/languageassurance",
		"internal/verify/scope_source_authority_activation.go",
	}
}
