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
	branchScopeAllowlist["agent/language-readiness-artifact-v32"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/transformation-effect.yml",
		"cmd/language-readiness-witness",
		"internal/meta/languagereadiness/artifact",
		"internal/meta/languagereadiness/improvement/adapter.go",
		"internal/verify/scope_language_readiness.go",
	}
	branchScopeAllowlist["agent/language-readiness-improvement-v33"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/transformation-effect.yml",
		"cmd/language-readiness-witness",
		"internal/meta/languageconcept",
		"internal/meta/languagereadiness/artifact",
		"internal/verify/scope_language_readiness.go",
	}
	branchScopeAllowlist["agent/readiness-autonomy-change-proposal"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/transformation-effect.yml",
		"cmd/language-readiness-witness",
		"internal/meta/languageconcept",
		"internal/meta/languagereadiness",
		"internal/verify/scope_language_readiness.go",
	}
	branchScopeAllowlist["agent/readiness-guarded-promotion"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/metric-counterfactual.yml",
		".github/workflows/transformation-effect.yml",
		"cmd/guarded-promotion-witness",
		"docs/language/guarded-promotion.md",
		"internal/meta/languagereadiness/guardedpromotion",
		"internal/verify/scope_language_readiness.go",
	}
	branchScopeAllowlist["agent/readiness-squash-predecessor"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/metric-counterfactual.yml",
		"docs/language/proposal-synthesis-continuity.md",
		"internal/meta/languagereadiness/proposalpromotion",
		"internal/meta/metricstrategy/proposalpredecessor",
		"internal/verify/scope_language_readiness.go",
	}
	branchScopeAllowlist["agent/readiness-baseline-resolution"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		"cmd/language-readiness-witness/predecessor-selection",
		"docs/language/readiness-ancestor-resolution.md",
		"internal/meta/languagereadiness/artifact/predecessorresolution",
		"internal/verify/scope_language_readiness.go",
	}
	branchScopeAllowlist["agent/readiness-push-continuity"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/metric-counterfactual.yml",
		"docs/language/readiness-push-continuity.md",
		"internal/verify/readiness_push_continuity_test.go",
		"internal/verify/scope_language_readiness.go",
	}
}
