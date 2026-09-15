package languageconcept

import oracle "github.com/kimjooyoon/meta-ontology-go/internal/meta/languagesourcebindingpromotion"

var languageSourceBindingPromotionConcept = Concept{
	ID:             "language-source-binding-promotion",
	Problem:        "A deterministic artifact can remain semantically unbound when its producer validates itself.",
	PositiveEffect: "A dependency-separated oracle projects bounded Gooo source semantics and rejects digest-valid forgeries.",
	MetaOperation:  "independently-project-source",
	Rarity:         "UNCOMMON_COMBINATION",
	Stage:          "OPERATING",
	NoveltyClaim:   false,
	CodeBindings: []string{
		"internal/meta/languagesourcebindingpromotion",
		"cmd/language-source-binding-promotion",
		"examples/language-source-binding-promotion",
		"scripts/language-source-binding-promotion",
	},
	MetricBindings: oracle.MetricIDs(),
	UseCases: []UseCase{
		{ID: "genuine-source-bound", Trigger: "a real execution artifact is compared with raw Gooo source", ExpectedOutcome: "PASS_1_OF_1_EXACT_SOURCE_BINDING"},
		{ID: "resealed-output-forgery", Trigger: "a digest-valid output entity is changed after execution", ExpectedOutcome: "FAIL_CLOSED_ARTIFACT_SOURCE_PROJECTION_MISMATCH"},
		{ID: "unknown-artifact-decision", Trigger: "a digest-valid artifact carries an unknown decision", ExpectedOutcome: "FAIL_CLOSED_ARTIFACT_DECISION_UNKNOWN"},
		{ID: "unsupported-source-projection", Trigger: "source is outside the fixed oracle grammar", ExpectedOutcome: "FAIL_CLOSED_LOWER_RESOLUTION_AT_ORACLE_PARSE"},
	},
}
