package verify

const foundationBootstrapDevSyncBranch = "agent/foundation-bootstrap-dev-sync-20260831"

func init() {
	branchScopeAllowlist[foundationBootstrapDevSyncBranch] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/foundation-bootstrap-dev-sync.json",
		".github/workflows/ci-guardian.yml",
		"internal/verify/scope_foundation_bootstrap_dev_sync_20260831.go",
		"scripts/ci-proof/foundation_bootstrap.js",
		"scripts/ci-proof/guardian.js",
	}
}
