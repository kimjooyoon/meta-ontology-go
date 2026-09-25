package verify

func init() {
	branchScopeAllowlist["agent/gooo-independent-artifact-oracle"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/transformation-effect.yml",
		"cmd/language-artifact-oracle",
		"examples/language-artifact-oracle",
		"internal/meta/languageartifactoracle",
		"internal/meta/languageconcept",
		"internal/verify/scope_language_artifact_oracle.go",
		"scripts/language-artifact-oracle",
	}
}
