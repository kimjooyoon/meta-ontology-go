package verify

func init() {
	branchScopeAllowlist["agent/minimal-claim-resolution-contract"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/claim-resolution-tuple.yml",
		"cmd/gooo",
		"docs/language/claim-resolution-tuple.md",
		"examples/claim-resolution-tuple",
		"internal/claimresolution",
		"internal/verify/scope_claim_resolution_tuple.go",
	}
}
