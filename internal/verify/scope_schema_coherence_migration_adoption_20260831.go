package verify

const schemaCoherenceMigrationAdoptionBranch = "agent/schema-coherence-migration-adoption-20260831"

func init() {
	branchScopeAllowlist[schemaCoherenceMigrationAdoptionBranch] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/governance-denominator-v4-schema-coherence.json",
		".github/workflows/ci-guardian.yml",
		".github/workflows/ci.yml",
		"internal/verify/scope_schema_coherence_migration_adoption_20260831.go",
		"scripts/ci-proof/foundation_authorization.js",
		"scripts/ci-proof/foundation_authorization_test.js",
		"scripts/ci-proof/guardian.js",
		"scripts/ci-proof/guardian_test.js",
	}
}
