package verify

func init() {
	branchScopeAllowlist["agent/language-readiness-foundation-seed"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/transformation-effect.yml",
		"cmd/language-readiness-witness/foundation-seed",
		"internal/meta/languagereadiness/artifact/foundationseed",
		"internal/meta/languagereadiness/languagesemantic/registry_definition.go",
		"internal/verify/scope_language_readiness_foundation_seed.go",
	}
}
