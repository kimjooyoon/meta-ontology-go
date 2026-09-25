package verify

func init() {
	branchScopeAllowlist["agent/gooo-policy-foundation-binding"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/transformation-effect.yml",
		"cmd/external-capability-execution-witness/authorization-foundation",
		"examples/external-capability-execution/authorization/foundation.json",
		"internal/meta/externalcapabilityexecution/authorizationfoundation",
		"internal/verify/scope_external_capability_authorization_foundation.go",
	}
}
