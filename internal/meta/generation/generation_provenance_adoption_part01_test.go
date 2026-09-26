package generation

import (
	"strings"
	"testing"
)

func TestGenerationProvenanceAdoptionRequiresCompleteReplay(t *testing.T) {
	digest := "sha256:" + strings.Repeat("0", 64)
	replay, err := BuildGenerationProvenanceReplayPart01(
		"execution-plan-1",
		digest,
		digest,
		digest,
		generationProvenancePassPart01,
		"GENERATION_AND_REVERSE_OBSERVED",
		-1,
	)
	if err != nil {
		t.Fatalf("replay build error = %v", err)
	}
	adoption, err := EvaluateGenerationProvenanceAdoptionPart01(replay)
	if err != nil {
		t.Fatalf("adoption evaluation error = %v", err)
	}
	if adoption.Decision != generationProvenanceAdoptPart01 || !adoption.ValidPart01() {
		t.Fatalf("complete replay was not adopted: %#v", adoption)
	}
}

func TestGenerationProvenanceAdoptionPreservesUnknownFrontier(t *testing.T) {
	digest := "sha256:" + strings.Repeat("0", 64)
	replay, err := BuildGenerationProvenanceReplayPart01(
		"execution-plan-1",
		digest,
		"",
		"",
		generationProvenanceUnknownPart01,
		"MISSING_REVERSE_EVIDENCE",
		2,
	)
	if err != nil {
		t.Fatalf("replay build error = %v", err)
	}
	adoption, err := EvaluateGenerationProvenanceAdoptionPart01(replay)
	if err != nil {
		t.Fatalf("adoption evaluation error = %v", err)
	}
	if adoption.Decision != generationProvenanceAdoptionUnknownPart01 || adoption.MissingStageIndex != 2 {
		t.Fatalf("unknown replay frontier changed: %#v", adoption)
	}
}

func TestGenerationProvenanceAdoptionRejectsTampering(t *testing.T) {
	digest := "sha256:" + strings.Repeat("0", 64)
	replay, err := BuildGenerationProvenanceReplayPart01(
		"execution-plan-1",
		digest,
		digest,
		digest,
		generationProvenancePassPart01,
		"GENERATION_AND_REVERSE_OBSERVED",
		-1,
	)
	if err != nil {
		t.Fatalf("replay build error = %v", err)
	}
	adoption, err := EvaluateGenerationProvenanceAdoptionPart01(replay)
	if err != nil {
		t.Fatalf("adoption evaluation error = %v", err)
	}
	adoption.ReverseDigest = "sha256:" + strings.Repeat("1", 64)
	if adoption.ValidPart01() {
		t.Fatal("tampered adoption unexpectedly validated")
	}
}
