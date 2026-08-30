package verify

const temporalTransitionTicketV1Branch = "agent/temporal-transition-ticket-v1"

func init() {
	branchScopeAllowlist[temporalTransitionTicketV1Branch] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/temporal-transition-ticket-v1.yml",
		".github/workflows/transformation-effect.yml",
		"contracts",
		"examples/temporal-transition-ticket",
		"examples/language-syntax-roundtrip/corpus.json",
		"fixtures/temporal-transition-ticket",
		"internal/meta/languagereadiness/languagesyntax/conformance/evaluate_test.go",
		"internal/meta/languagereadiness/languagesyntax/model.go",
		"internal/meta/languagereadiness/languagesyntax/registry.go",
		"internal/verify/scope_temporal_transition_ticket_v1.go",
		"internal/verify/scope_temporal_transition_ticket_v1_test.go",
		"scripts/bind-activity-resolutions.sh",
		"scripts/conform-temporal-transition-ticket-v1.sh",
		"scripts/evaluate-temporal-transition-ticket-v1.sh",
	}
}
