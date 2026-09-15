package languageconcept

var toolchainConformanceConcept = Concept{
	ID:             "toolchain-conformance",
	Problem:        "Independent green tools can disagree on schema, commit identity, evidence, or effect authority.",
	PositiveEffect: "A versioned metaprogram closes nine exact-head tool receipts and rejects bounded drift in memory.",
	MetaOperation:  "close-toolchain-conformance-ledger",
	Rarity:         "UNCOMMON_COMBINATION",
	Stage:          "OPERATING",
	NoveltyClaim:   false,
	CodeBindings: []string{
		"internal/meta/languagereadiness/toolchainconformance",
		"cmd/toolchain-conformance-witness",
		"examples/toolchain-conformance",
		".github/workflows/transformation-effect.yml",
		"docs/language/toolchain-conformance.md",
	},
	MetricBindings: toolchainConformanceMetricBindings(),
	UseCases: []UseCase{
		{ID: "same-head-surface-closure", Trigger: "CI joins nine versioned tool receipts", ExpectedOutcome: "9_OF_9_SURFACES_152_OF_152_CASES"},
		{ID: "in-memory-drift-rejection", Trigger: "CI applies thirteen bounded receipt mutations", ExpectedOutcome: "13_OF_13_FAIL_CLOSED"},
		{ID: "effect-authority-boundary", Trigger: "a receipt reports writes or mutation authority", ExpectedOutcome: "ZERO_WRITES_ZERO_AUTHORITIES"},
	},
}
