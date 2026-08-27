package verify

func init() {
	branchScopeAllowlist["agent/luna-meta-28-evidence-quorum"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/evidence-quorum.yml",
		"cmd/evidence-quorum-witness",
		"cmd/evidence-quorum-source-channel",
		"cmd/evidence-quorum-reconstructor",
		"cmd/evidence-quorum-artifact-observer",
		"cmd/evidence-quorum-counterexample",
		"examples/evidence-quorum",
		"internal/meta/evidencequorumchannel",
		"internal/meta/evidencequorumconsumer",
		"internal/meta/evidencequorumpolicy",
		"internal/meta/evidencequorumwire",
		"internal/meta/languageconcept",
		"internal/verify/scope_evidence_quorum.go",
		"scripts/evidence-quorum",
	}
}
