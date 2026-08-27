package languageconcept

var deterministicAmbiguityBudgetConcept = Concept{
	ID:             "deterministic-ambiguity-budget",
	Problem:        "A confidence label can hide unresolved interpretations and the evidence needed to resolve them.",
	PositiveEffect: "An executable integer budget preserves ambiguity cardinalities, lower-resolution transitions, and intervention evidence.",
	MetaOperation:  "measure-deterministic-ambiguity-budget",
	Rarity:         "UNCOMMON_COMBINATION",
	Stage:          "OPERATING",
	NoveltyClaim:   false,
	CodeBindings: []string{
		".github/workflows/ambiguity-budget.yml",
		"cmd/ambiguity-budget-verifier",
		"cmd/ambiguity-budget-witness",
		"docs/language/deterministic-ambiguity-budget.md",
		"examples/ambiguity-budget",
		"examples/language-semantic-model/corpus.json",
		"examples/language-syntax-roundtrip/corpus.json",
		"examples/toolchain-conformance/corpus.json",
		"examples/vertical-slice-closure-shadow/usecases.json",
		"internal/meta/ambiguitybudget",
		"internal/meta/ambiguitybudgetjudge",
		"internal/meta/languagereadiness/languagesemanticbinding",
		"internal/verify/scope_ambiguity_budget.go",
		"scripts/ambiguity-budget",
	},
	MetricBindings: []string{
		"gooo.metric.meta.ambiguity-budget.integer-observations.v2",
		"gooo.metric.meta.ambiguity-budget.lower-resolution-claims.v2",
		"gooo.metric.meta.ambiguity-budget.intervention-separation.v2",
	},
	UseCases: []UseCase{
		{ID: "bounded-ambiguity-observation", Trigger: "four executable integer cases contain zero, boundary, over-budget, and unknown observations", ExpectedOutcome: "CONFORMANCE_PASS_SUBJECT_LOWER_RESOLUTION_4_CASES_2_INTERVENTIONS"},
	},
}
