package verify

func init() {
	branchScopeAllowlist["agent/luna-meta-04-capability-scoped-expansion"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/capability-scoped-expansion.yml",
		"cmd/capability-scoped-expansion-provider",
		"cmd/capability-scoped-expansion-witness",
		"docs/language/capability-scoped-expansion.md",
		"examples/capability-scoped-expansion",
		"internal/meta/capabilityscopedexpansion",
		"internal/verify/scope_capability_scoped_expansion.go",
	}
}
