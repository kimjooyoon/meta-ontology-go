package extractor

import "testing"

func TestProfileObservationRecordHelpersRetainRenderedCapacity(t *testing.T) {
	requireNamedCapacities(t, "../../../..", []string{
		"internal/syntax#SupportedEntityFieldsProfiles",
		"internal/syntax#EntityFieldsProfile.canonicalProfile",
		"internal/meta/syntaxregistration#registrationHintProfiles",
		"internal/meta/syntaxregistration#latestExerciseProfile",
		"internal/meta/syntaxregistration#bindProfileObservationLayout",
		"internal/meta/syntaxregistration#bindContractSource",
		"internal/meta/syntaxregistration#inspectProfileHintSourceForCapability",
		"internal/meta/syntaxregistration#replayProfileHintSource",
	})
}
