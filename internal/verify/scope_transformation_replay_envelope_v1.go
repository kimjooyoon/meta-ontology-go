package verify

func init() {
	branchScopeAllowlist["agent/transformation-replay-envelope-v1"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/transformation-effect.yml",
		"internal/meta/transformationeffect",
		"internal/verify/scope_transformation_replay_envelope_v1.go",
		"scripts/self-improvement-cycle",
		"scripts/transformation-effect",
	}
}
