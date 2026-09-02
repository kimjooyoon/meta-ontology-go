package verify

const guardianTrustedObserverRecovery20260903Branch = "agent/guardian-trusted-observer-recovery-20260903-rerun"

func init() {
	branchScopeAllowlist[guardianTrustedObserverRecovery20260903Branch] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/ci-guardian.yml",
		"internal/verify/scope_guardian_trusted_observer_recovery_20260903.go",
		"scripts/ci-proof/guardian.js",
		"scripts/ci-proof/guardian_test.js",
	}
}
