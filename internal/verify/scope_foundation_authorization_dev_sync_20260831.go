package verify

const foundationAuthorizationDevSyncBranch = "agent/foundation-authorization-dev-sync-20260831"

func init() {
	branchScopeAllowlist[foundationAuthorizationDevSyncBranch] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/foundation-authorization.json",
		".github/workflows/ci-guardian.yml",
		".github/workflows/ci.yml",
		"internal/verify/foundation_authorization.go",
		"internal/verify/scope_foundation_authorization_dev_sync_20260831.go",
		"scripts/ci-proof/foundation_authorization.js",
		"scripts/ci-proof/foundation_authorization_test.js",
		"scripts/ci-proof/guardian.js",
	}
}
