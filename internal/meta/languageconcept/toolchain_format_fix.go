package languageconcept

var toolchainFormatFixConcept = Concept{
	ID:             "toolchain-format-fix",
	Problem:        "Formatting and fixing can mutate files without a versioned semantic plan or fixed-point proof.",
	PositiveEffect: "One built binary emits canonical source and read-only plans that a metaprogram applies in memory to an explicit fixed point.",
	MetaOperation:  "evaluate-toolchain-format-fix",
	Rarity:         "UNCOMMON_COMBINATION",
	Stage:          "OPERATING",
	NoveltyClaim:   false,
	CodeBindings: []string{
		"cmd/gooo",
		"internal/formatter",
		"internal/formatfix",
		"internal/meta/languagereadiness/toolchainformatfix",
		"cmd/toolchain-format-fix-witness",
		"examples/toolchain-format-fix",
	},
	MetricBindings: toolchainFormatFixMetricBindings(),
	UseCases: []UseCase{
		{ID: "format-canonical-source", Trigger: "CI invokes text, JSON, and check modes", ExpectedOutcome: "3_OF_3_FORMAT_PATHS"},
		{ID: "plan-without-writing", Trigger: "CI observes changed, fixed-point, and text fix plans", ExpectedOutcome: "3_OF_3_FIX_PATHS_WITH_ZERO_WRITES"},
		{ID: "reject-unknown-boundaries", Trigger: "CI invokes six malformed, missing, unknown, or write-authority paths", ExpectedOutcome: "6_OF_6_REJECTED"},
	},
}
