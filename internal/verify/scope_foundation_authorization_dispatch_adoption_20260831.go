package verify

const foundationAuthorizationDispatchAdoptionBranch = "agent/foundation-authorization-dispatch-adoption-20260831"

func init() {
	branchScopeAllowlist[foundationAuthorizationDispatchAdoptionBranch] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/foundation-authorization.json",
		".github/governance-denominator-v6-foundation-authorization-dispatch.json",
		".github/workflows/ci-guardian.yml",
		".github/workflows/ci-effort-observation.yml",
		".github/workflows/ci.yml",
		"examples/ci-time-causality",
		"examples/receipt-schema-migration-v3",
		"internal/verify/scope_foundation_authorization_dispatch_adoption_20260831.go",
		"scripts/ci-effort-observation",
		"scripts/ci-proof/foundation_authorization.js",
		"scripts/ci-proof/foundation_authorization_test.js",
		"scripts/ci-proof/guardian.js",
		"scripts/ci-proof/guardian_test.js",
		"scripts/ci-proof/source_migration_v3_test.js",
		"scripts/ci-proof/time_causality_v011_test.js",
	}
}
