package languageconcept

import release "github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/toolchainrelease"

var toolchainCrossPlatformReleaseConcept = Concept{
	ID:             "toolchain-cross-platform-release",
	Problem:        "A build that succeeds on one host can be mistaken for a releasable language toolchain.",
	PositiveEffect: "Three fixed native runners emit source-bound, replay-equal binaries and archives before readiness credit.",
	MetaOperation:  release.MetaOperation,
	Rarity:         "UNCOMMON_COMBINATION",
	Stage:          "OPERATING",
	NoveltyClaim:   false,
	CodeBindings: []string{
		"cmd/gooo",
		"cmd/toolchain-release-platform-witness",
		"cmd/toolchain-release-witness",
		"internal/meta/languagereadiness/toolchainrelease",
		"examples/toolchain-cross-platform-release",
		".github/workflows/transformation-effect.yml",
	},
	MetricBindings: release.MetricIDs(),
	UseCases: []UseCase{
		{ID: "three-native-platforms", Trigger: "CI builds and runs the exact head on three fixed x64 runners", ExpectedOutcome: "3_OF_3_EXACT_PLATFORM_RECEIPTS"},
		{ID: "reproducible-release-candidate", Trigger: "each runner builds and packages the CLI twice", ExpectedOutcome: "6_BINARY_AND_6_ARCHIVE_BUILDS_BYTE_EQUAL"},
		{ID: "uncertain-platform-receipt", Trigger: "a receipt is missing, duplicated, stale, or has an unknown decision", ExpectedOutcome: "FAIL_CLOSED_WITHOUT_READINESS_CREDIT"},
	},
}
