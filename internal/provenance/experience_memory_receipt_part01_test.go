package provenance

import "testing"

func completeExperienceMemoryReceipt() ExperienceMemoryReceipt {
	receipt := NewExperienceMemoryReceipt()
	receipt.ExperienceID = "experience-001"
	receipt.Fingerprint = ExperienceMemoryFingerprint{
		ScenarioID:      "memory-replay",
		SourceDigest:    "sha256:source",
		FixtureDigest:   "sha256:fixture",
		ToolchainDigest: "sha256:toolchain",
		EvaluatorDigest: "sha256:evaluator",
		ScopeDigest:     "sha256:scope",
	}
	receipt.OutcomeDigest = "sha256:outcome"
	receipt.Status = ExperienceMemoryClosed
	return receipt
}

func TestExperienceMemoryStaleFingerprintBecomesUnknown(t *testing.T) {
	receipt := completeExperienceMemoryReceipt()
	expected := receipt.Fingerprint
	expected.SourceDigest = "sha256:new-source"
	observed := receipt.Observe(expected)
	if observed.Status != ExperienceMemoryUnknown {
		t.Fatalf("expected UNKNOWN, got %q", observed.Status)
	}
	if observed.MissingStageIndex != 0 || observed.MissingStage != "fingerprint" {
		t.Fatalf("unexpected UNKNOWN coordinates: %#v", observed)
	}
	if err := observed.Validate(); err != nil {
		t.Fatalf("validate observed receipt: %v", err)
	}
}

func TestExperienceMemoryExactFingerprintRemainsClosed(t *testing.T) {
	receipt := completeExperienceMemoryReceipt()
	observed := receipt.Observe(receipt.Fingerprint)
	if observed.Status != ExperienceMemoryClosed {
		t.Fatalf("expected CLOSED, got %q", observed.Status)
	}
	if err := observed.Validate(); err != nil {
		t.Fatalf("validate observed receipt: %v", err)
	}
}

func TestExperienceMemoryPreservesFirstUnknownStage(t *testing.T) {
	receipt := completeExperienceMemoryReceipt()
	receipt.Status = ExperienceMemoryUnknown
	_ = receipt.PreserveFirstUnknown(1, "outcome", "receipt missing")
	_ = receipt.PreserveFirstUnknown(2, "scope", "scope changed")
	if receipt.MissingStageIndex != 1 || receipt.MissingStage != "outcome" {
		t.Fatalf("first UNKNOWN stage was not preserved: %#v", receipt)
	}
}
