package languageconcept

var languageDeterministicQueryConcept = Concept{
	ID:             "language-deterministic-query",
	Problem:        "language and metric bindings cannot be inspected reproducibly when a query is only an ambient procedure",
	PositiveEffect: "reified bounded query plans return canonical receipts while keeping candidate facts non-authoritative",
	MetaOperation:  "execute-reified-deterministic-query-plan",
	Rarity:         "UNCOMMON_COMBINATION",
	Stage:          "OPERATING",
	NoveltyClaim:   false,
	CodeBindings: []string{
		"internal/query",
		"internal/meta/languagereadiness/languagedeterministicquery",
		"internal/meta/languagereadiness/languagedeterministicquerybinding",
		"cmd/language-deterministic-query-witness",
		"cmd/language-deterministic-query-readiness-binding",
		"examples/language-deterministic-query",
	},
	MetricBindings: languageDeterministicQueryMetricBindings(),
	UseCases: []UseCase{
		{
			ID:              "query-meta-bindings",
			Trigger:         "CI needs to inspect the code and metric bindings of the deterministic query concept",
			ExpectedOutcome: "SATISFIED_32_OF_32_WITH_CANONICAL_REPLAY",
		},
		{
			ID:              "replay-query-plan",
			Trigger:         "the same reified query plan is executed against an insertion-permuted meta graph",
			ExpectedOutcome: "IDENTICAL_REQUEST_RESULT_AND_GRAPH_RECEIPTS",
		},
		{
			ID:              "reject-query-unknowns",
			Trigger:         "a query selects an unknown layer or endpoint, or observes only a candidate fact",
			ExpectedOutcome: "FAIL_CLOSED_WITH_ZERO_AUTHORITY_PROMOTIONS",
		},
	},
}
