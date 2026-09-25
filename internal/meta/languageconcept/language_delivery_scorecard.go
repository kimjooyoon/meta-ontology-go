package languageconcept

import delivery "github.com/kimjooyoon/meta-ontology-go/internal/meta/languagedelivery"

var languageDeliveryScorecardConcept = Concept{
	ID:             "language-delivery-scorecard",
	Problem:        "Internal readiness can be mistaken for user-visible language completeness.",
	PositiveEffect: "A fixed delivery contract projects one evidence set at reader-specific resolutions and preserves known gaps.",
	MetaOperation:  delivery.MetaOperation,
	Rarity:         "UNCOMMON_COMBINATION",
	Stage:          "OPERATING",
	NoveltyClaim:   false,
	CodeBindings: []string{
		"internal/meta/languagedelivery",
		"cmd/language-delivery-scorecard",
		"examples/language-delivery-scorecard",
	},
	MetricBindings: delivery.MetricIDs(),
	UseCases: []UseCase{
		{ID: "known-delivery-gaps", Trigger: "five fixed obligations have no executable receipt", ExpectedOutcome: "INCOMPLETE_31_OF_36"},
		{ID: "reader-resolution", Trigger: "user, tool author, and governor inspect the same facts", ExpectedOutcome: "9_OF_12_19_OF_24_31_OF_36"},
		{ID: "unknown-source-receipt", Trigger: "an upstream decision is not explicit PASS", ExpectedOutcome: "FAIL_CLOSED_LOWER_RESOLUTION"},
	},
}
