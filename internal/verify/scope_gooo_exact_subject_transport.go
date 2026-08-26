package verify

func init() {
	branchScopeAllowlist["agent/gooo-exact-subject-transport"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/self-improvement-language-candidate.yml",
		".github/workflows/self-improvement-language-observation.yml",
		"examples/self-improvement/TRANSPORT.md",
		"examples/self-improvement/transport.gooo",
		"internal/meta/selfimprovementtransport",
		"internal/verify/scope_gooo_exact_subject_transport.go",
		"scripts/self-improvement-transport",
	}
}
