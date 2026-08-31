package verify

const executableGuardianScopeAdoptionBranch = "agent/executable-guardian-scope-adoption-20260831"

func init() {
	branchScopeAllowlist[executableGuardianScopeAdoptionBranch] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/governance-denominator-v5-executable-guardian-scope.json",
		".github/workflows/ci-guardian.yml",
		"internal/verify/scope_executable_guardian_scope_adoption_20260831.go",
		"scripts/ci-proof/foundation_authorization.js",
		"scripts/ci-proof/foundation_authorization_test.js",
		"scripts/ci-proof/guardian.js",
		"scripts/ci-proof/guardian_test.js",
	}
}
