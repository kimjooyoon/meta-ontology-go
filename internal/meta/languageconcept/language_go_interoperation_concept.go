package languageconcept

var languageGoInteroperationConcept = Concept{
	ID:             "language-go-interoperation",
	Problem:        "generated Go can parse while its exported type boundary, replay identity, or rejection authority remains unproved",
	PositiveEffect: "a pure metaprogram reifies existing generator output as Go AST and types, normalizes the exported API, and rejects unknown or ambient boundaries",
	MetaOperation:  "reify-go-projection-and-prove-type-identity",
	Rarity:         "UNCOMMON_COMBINATION",
	Stage:          "OPERATING",
	NoveltyClaim:   false,
	CodeBindings: []string{
		"internal/generator",
		"internal/meta/languagereadiness/languagegointeroperation",
		"internal/meta/languagereadiness/languagegointeroperationbinding",
		"cmd/language-go-interoperation-witness",
		"cmd/language-go-interoperation-readiness-binding",
		"examples/language-go-interoperation",
	},
	MetricBindings: languageGoInteroperationMetricBindings(),
	UseCases: []UseCase{
		{ID: "project-gooo-to-go-api", Trigger: "CI projects eight fixed SemanticIR fixtures through the existing generator",
			ExpectedOutcome: "SATISFIED_8_OF_8_WITH_SOURCE_MAP_AND_API_REPLAY"},
		{ID: "consume-go-1.27-boundary", Trigger: "CI reifies generic methods, generic aliases, and assignment inference with Go 1.27",
			ExpectedOutcome: "SATISFIED_8_OF_8_WITH_5_GENERIC_METHODS_AND_2_ALIASES"},
		{ID: "reject-go-boundary-unknowns", Trigger: "a boundary is invalid, imports ambient authority, exports nothing, or has an unknown payload",
			ExpectedOutcome: "REJECTED_8_OF_8_WITH_ZERO_INVALID_ACCEPTANCES"},
	},
}
