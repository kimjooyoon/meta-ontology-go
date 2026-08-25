package languageconcept

import execution "github.com/kimjooyoon/meta-ontology-go/internal/meta/languagesourceexecution"

var languageSourceExecutionConcept = Concept{
	ID:             "language-source-execution",
	Problem:        "A parsed activity contract can exist without a user-executable language transition.",
	PositiveEffect: "Gooo executes its own ontology transition and emits deterministic, fail-closed receipts.",
	MetaOperation:  "execute-source-activity",
	Rarity:         "UNCOMMON_COMBINATION",
	Stage:          "OPERATING",
	NoveltyClaim:   false,
	CodeBindings: []string{
		"internal/sourceexecution",
		"cmd/gooo",
		"internal/meta/languagesourceexecution",
		"cmd/language-source-execution-witness",
		"examples/language-source-execution",
	},
	MetricBindings: execution.MetricIDs(),
	UseCases: []UseCase{
		{ID: "execute-billing", Trigger: "a user selects the declared PayOrder activity", ExpectedOutcome: "PASS_1_SOURCE_EXECUTION_4_EVENTS"},
		{ID: "deterministic-replay", Trigger: "the same source and entry execute twice", ExpectedOutcome: "EXACT_BYTE_REPLAY_1_OF_1"},
		{ID: "unknown-entry", Trigger: "the selected activity is absent", ExpectedOutcome: "FAIL_CLOSED_SOURCE_ENTRY_UNKNOWN"},
	},
}
