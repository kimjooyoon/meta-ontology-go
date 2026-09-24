package generation

import "testing"

func TestObserveExperienceMemoryPreservesUnknownOnScopeChange(t *testing.T) {
	receipt := ExperienceMemoryReceipt{
		SchemaVersion: ExperienceMemoryReceiptSchema,
		ExperienceID:  "generation-memory",
		Fingerprint: ExperienceMemoryFingerprint{
			ScenarioID:      "scope-replay",
			SourceDigest:    "sha256:source",
			FixtureDigest:   "sha256:fixture",
			ToolchainDigest: "sha256:toolchain",
			EvaluatorDigest: "sha256:evaluator",
			ScopeDigest:     "sha256:old-scope",
		},
		OutcomeDigest: "sha256:outcome",
		Status:        ExperienceMemoryClosed,
	}
	expected := receipt.Fingerprint
	expected.ScopeDigest = "sha256:new-scope"
	observed := ObserveExperienceMemory(receipt, expected)
	if observed.Status != ExperienceMemoryUnknown || observed.MissingStageIndex != 0 {
		t.Fatalf("expected first UNKNOWN at fingerprint: %#v", observed)
	}
}
