package languageconcept

var languagePackageRuntimeConcept = Concept{
	ID:             "language-package-runtime",
	Problem:        "Parsed .gooo files did not form a deterministic multi-package runtime unit.",
	PositiveEffect: "A fixed corpus now proves canonical package initialization and activity-contract resolution.",
	MetaOperation:  "evaluate-language-package-runtime",
	Rarity:         "UNCOMMON_COMBINATION",
	Stage:          "OPERATING",
	NoveltyClaim:   false,
	CodeBindings: []string{
		"internal/packageruntime",
		"internal/meta/languagereadiness/languagepackageruntime",
		"cmd/language-package-runtime-witness",
		"examples/language-package-runtime",
	},
	MetricBindings: languagePackageRuntimeMetricBindings(),
	UseCases: []UseCase{
		{ID: "deterministic-package-graph", Trigger: "four packages with a diamond import graph", ExpectedOutcome: "10_OF_10_POSITIVE_PATHS"},
		{ID: "multi-source-entry-contract", Trigger: "two .gooo sources and an entry activity", ExpectedOutcome: "ENTRY_RESOLVED_WITH_ZERO_EFFECTS"},
		{ID: "invalid-runtime-fail-closed", Trigger: "eight invalid package runtime boundaries", ExpectedOutcome: "8_OF_8_REJECTED"},
	},
}
