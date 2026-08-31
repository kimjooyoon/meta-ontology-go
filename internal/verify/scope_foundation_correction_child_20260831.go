package verify

const foundationCorrectionChildBranch = "agent/foundation-correction-child-20260831"

func init() {
	branchScopeAllowlist[foundationCorrectionChildBranch] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/governance-denominator-v2-migration.json",
		".github/governance-denominator-v3-correction.json",
		".github/workflows/ci-guardian.yml",
		"internal/verify/scope_foundation_correction_child_20260831.go",
		"scripts/ci-proof/foundation_authorization.js",
		"scripts/ci-proof/foundation_authorization_test.js",
		"scripts/ci-proof/guardian.js",
		"scripts/ci-proof/guardian_test.js",
	}
}
