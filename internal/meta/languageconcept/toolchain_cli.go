package languageconcept

var toolchainCLIConcept = Concept{
	ID:             "toolchain-cli",
	Problem:        "A command surface can exist without proving its executable exit, output, replay, or effect contract.",
	PositiveEffect: "A fixed meta evaluator now binds the built gooo binary to deterministic success and rejection receipts.",
	MetaOperation:  "evaluate-toolchain-cli-contract",
	Rarity:         "UNCOMMON_COMBINATION",
	Stage:          "OPERATING",
	NoveltyClaim:   false,
	CodeBindings: []string{
		"cmd/gooo",
		"internal/toolchaincli",
		"internal/meta/languagereadiness/toolchaincli",
		"cmd/toolchain-cli-witness",
		"examples/toolchain-cli",
	},
	MetricBindings: toolchainCLIMetricBindings(),
	UseCases: []UseCase{
		{ID: "inspect-cli-identity", Trigger: "CI invokes text and structured version contracts", ExpectedOutcome: "2_OF_2_IDENTITY_PATHS"},
		{ID: "execute-language-commands", Trigger: "CI checks syntax, semantics, and roundtrip through the built binary", ExpectedOutcome: "4_OF_4_LANGUAGE_OPERATIONS"},
		{ID: "reject-cli-boundaries", Trigger: "CI invokes six missing or unknown command boundaries", ExpectedOutcome: "6_OF_6_REJECTED_WITH_ZERO_WRITES"},
	},
}
