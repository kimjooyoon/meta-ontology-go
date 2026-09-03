package verify

func init() {
	branchScopeAllowlist["agent/foundation-authorization-protocol-20260903"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/foundation-authorization-protocol.json",
		".github/workflows/ci.yml",
		"internal/verify/scope_foundation_authorization_protocol_20260903.go",
		"scripts/ci-proof/foundation_authorization_protocol.js",
		"scripts/ci-proof/foundation_authorization_protocol_test.js",
	}
}
