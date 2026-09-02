package verify

const devMainSync20260902Branch = "agent/dev-main-sync-20260902"

func init() {
	branchScopeAllowlist[devMainSync20260902Branch] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/ci.yml",
		".github/workflows/self-improvement-minimal-loop.yml",
		".github/workflows/transformation-effect.yml",
		"bootstrap/repository-projector/command",
		"cmd/feedback-predecessor-witness",
		"examples/language-syntax-roundtrip",
		"examples/self-improvement-minimal-loop",
		"internal/detection/linecaps",
		"internal/meta/feedbackpredecessor",
		"internal/meta/generation",
		"internal/meta/languageassurance/verticalsliceclosureshadow",
		"internal/meta/languagereadiness/languagesyntax/conformance/evaluate_test.go",
		"internal/meta/languagereadiness/languagesyntax/model.go",
		"internal/meta/languagereadiness/languagesyntax/registry.go",
		"internal/meta/metabinding",
		"internal/meta/selfimprovementloop",
		"internal/meta/sourcepolicy",
		"internal/verify",
		"scripts/ci-proof",
		"scripts/feedback-predecessor-ci",
		"scripts/self-improvement-minimal-loop",
		"scripts/verify",
	}
}
