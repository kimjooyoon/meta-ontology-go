package verify

func init() {
	branchScopeAllowlist["agent/core-conformance-registry-20260905"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/ci.yml",
		".github/workflows/core-conformance-registry.yml",
		".github/workflows/transformation-effect.yml",
		"examples/language-syntax-roundtrip/corpus.json",
		"internal/meta/languagereadiness/languagesyntax",
		"internal/meta/languageassurance/verticalsliceclosureshadow",
		"internal/meta/selfimprovementcontinuation/build.go",
		"internal/meta/selfimprovementcontinuation/model.go",
		"internal/meta/selfimprovementcontinuation/verify.go",
		"internal/meta/selfimprovementexecutiongrant/evaluate.go",
		"internal/meta/selfimprovementexecutiongrant/model.go",
		"internal/meta/selfimprovementexecutiongrant/verify.go",
		"internal/meta/selfimprovementcandidate/constants.go",
		"internal/verify/scope_gooo_core_conformance_registry.go",
	}
}
