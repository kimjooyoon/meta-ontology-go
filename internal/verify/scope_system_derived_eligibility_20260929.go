package verify

func init() {
	branchScopeAllowlist["agent/system-derived-eligibility-20260929"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/self-improvement-candidate-authorization.yml",
		".github/workflows/self-improvement-execution-contract.yml",
		".github/workflows/self-improvement-execution-grant.yml",
		"examples/self-improvement-execution-grant/README.md",
		"examples/self-improvement-execution-grant/grant.gooo",
		"examples/self-improvement/AUTHORIZATION.md",
		"examples/self-improvement/authorization.gooo",
		"internal/meta/selfimprovementcandidate/authorization.go",
		"internal/meta/selfimprovementcandidate/authorization_test.go",
		"internal/meta/selfimprovementexecutiongrant/canonical.go",
		"internal/meta/selfimprovementexecutiongrant/evaluate.go",
		"internal/meta/selfimprovementexecutiongrant/evaluate_test.go",
		"internal/meta/selfimprovementexecutiongrant/model.go",
		"internal/meta/selfimprovementexecutiongrant/verify.go",
		"internal/verify/scope_system_derived_eligibility_20260929.go",
		"scripts/self-improvement-candidate-authorization/main.go",
		"scripts/self-improvement-execution-grant/cases.go",
		"scripts/self-improvement-execution-grant/input.go",
		"scripts/self-improvement-execution-grant/live.go",
		"scripts/self-improvement-execution-grant/main.go",
		"scripts/self-improvement-execution-grant/verify.go",
	}
}
