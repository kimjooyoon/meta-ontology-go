package verify

const foundationRegressionRepairBranch = "agent/foundation-regression-repair-20260831"

func init() {
	branchScopeAllowlist[foundationRegressionRepairBranch] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/governance-denominator-v2-migration.json",
		".github/workflows/ci-guardian.yml",
		"internal/verify/scope_foundation_regression_repair_20260831.go",
		"scripts/ci-proof/foundation_authorization.js",
		"scripts/ci-proof/foundation_authorization_test.js",
		"scripts/ci-proof/guardian.js",
		"scripts/ci-proof/guardian_test.js",
	}
}
