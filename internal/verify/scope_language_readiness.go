package verify

func init() {
	branchScopeAllowlist["agent/quantified-language-readiness-v30"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		"internal/meta/languagereadiness",
		"internal/verify/scope_language_readiness.go",
	}
	branchScopeAllowlist["agent/quantified-improvement-transition-v31"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		"internal/meta/languagereadiness/improvement",
		"internal/verify/scope_language_readiness.go",
	}
}
