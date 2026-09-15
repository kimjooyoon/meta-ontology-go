package verify

func init() {
	branchScopeAllowlist["agent/readiness-guarded-capability"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/transformation-effect.yml",
		"cmd/language-readiness-witness",
		"docs/language/guarded-promotion-capability.md",
		"internal/meta/languageconcept",
		"internal/meta/languagereadiness",
		"internal/verify/scope_readiness_guarded_capability.go",
	}
}
