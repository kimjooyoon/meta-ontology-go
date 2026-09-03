package verify

func init() {
	branchScopeAllowlist["agent/guardian-successor-protocol-adoption-20260903"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/ci-guardian.yml",
		"internal/verify/scope_guardian_successor_protocol_adoption_20260903.go",
		"scripts/ci-proof/foundation_authorization_protocol.js",
		"scripts/ci-proof/guardian.js",
		"scripts/ci-proof/guardian_successor.js",
		"scripts/ci-proof/guardian_successor_test.js",
	}
}
