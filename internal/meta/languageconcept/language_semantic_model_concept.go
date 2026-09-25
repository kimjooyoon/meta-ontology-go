package languageconcept

var languageSemanticModelConcept = Concept{
	ID:             "language-semantic-model",
	Problem:        "syntax expansion can blur semantic authority and ambient effects",
	PositiveEffect: "normalized IR replay separates meaning, provenance, evidence, and sealed effects",
	MetaOperation:  "prove-staged-semantic-model",
	Rarity:         "UNCOMMON_COMBINATION",
	Stage:          "OPERATING",
	NoveltyClaim:   false,
	CodeBindings: []string{
		"internal/syntax",
		"internal/bidir",
		"internal/semantic",
		"internal/meta/languagereadiness/languagesemantic",
		"cmd/language-semantic-witness",
		"examples/language-semantic-model",
	},
	MetricBindings: languageSemanticModelMetrics,
	UseCases: []UseCase{
		{
			ID:              "staged-semantic-authority-replay",
			Trigger:         "thirteen sources, three authority laws, and two upstream rejections are replayed",
			ExpectedOutcome: "IMPROVED_14_TO_15_OF_24_WITH_18_OF_18_CASES_AND_ZERO_EFFECTS",
		},
	},
}
