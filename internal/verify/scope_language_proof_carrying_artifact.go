package verify

func init() {
	branchScopeAllowlist["agent/luna-meta-15-proof-carrying-artifact"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/language-proof-carrying-artifact.yml",
		"cmd/language-proof-carrying-artifact",
		"cmd/language-proof-carrying-artifact-verifier",
		"examples/language-proof-carrying-artifact",
		"internal/meta/languageconcept",
		"internal/meta/languageproofartifact",
		"internal/meta/languageproofartifactverifier",
		"internal/verify/scope_language_proof_carrying_artifact.go",
		"scripts/language-proof-carrying-artifact",
	}
}
