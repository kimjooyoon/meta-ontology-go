package verify

func init() {
	branchScopeAllowlist["agent/luna-meta-07-proof-choice-algebra"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/proof-choice-algebra.yml",
		"examples/proof-choice-algebra",
		"internal/meta/proofchoicealgebra",
		"internal/meta/proofchoicejudge",
		"internal/verify/scope_proof_choice_algebra.go",
		"scripts/proof-choice-algebra",
	}
}
