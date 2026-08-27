package verify

func init() {
	branchScopeAllowlist["agent/luna-meta-05-phase-separation-witness"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/transformation-effect.yml",
		"examples/phase-separation-witness",
		"internal/meta/phaseseparation",
		"internal/verify/scope_phase_separation_witness.go",
		"scripts/phase-separation-adjudicator",
		"scripts/phase-separation-witness",
	}
}
