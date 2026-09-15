package languageconcept

import metalsp "github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/toolchainlsp"

var toolchainLSPConcept = Concept{
	ID:             "toolchain-lsp",
	Problem:        "An editor server can expose stale or uncertain semantic navigation as exact language meaning.",
	PositiveEffect: "A fixed transcript projects exact parser state and withholds navigation whenever semantic evidence loses resolution.",
	MetaOperation:  metalsp.MetaOperation,
	Rarity:         "UNCOMMON_COMBINATION",
	Stage:          "OPERATING",
	NoveltyClaim:   false,
	CodeBindings: []string{
		"internal/lsp",
		"internal/lsp/coupling",
		"internal/meta/languagereadiness/toolchainlsp",
		"cmd/toolchain-lsp-witness",
		"examples/toolchain-lsp",
	},
	MetricBindings: metalsp.MetricIDs(),
	UseCases: []UseCase{
		{ID: "editor-read-session", Trigger: "CI replays the fixed standard LSP transcript", ExpectedOutcome: "16_OF_16_PROTOCOL_CASES"},
		{ID: "semantic-coupling-navigation", Trigger: "one exact coupling explanation covers the requested position", ExpectedOutcome: "ONE_STANDARD_LOCATION_LINK"},
		{ID: "uncertain-coupling", Trigger: "upstream meaning is unknown, failed, stale, or cancelled", ExpectedOutcome: "ZERO_NAVIGATION_LINKS"},
	},
}
