package verify

const publicTrustSurface20260904Branch = "agent/public-trust-surface-20260904-v19"

func init() {
	branchScopeAllowlist[publicTrustSurface20260904Branch] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/dependabot.yml",
		".github/workflows/dependency-review.yml",
		".github/workflows/public-trust-surface.yml",
		"README.md",
		"SECURITY.md",
		"examples/public-trust-surface",
		"examples/language-syntax-roundtrip/corpus.json",
		"internal/meta/languagereadiness/languagesyntax/registry.go",
		"internal/meta/languagereadiness/languagesyntax/conformance/evaluate_test.go",
		"internal/meta/publictrust",
		"internal/verify/scope_public_trust_surface_20260904.go",
		"scripts/public-trust-surface",
	}
}
