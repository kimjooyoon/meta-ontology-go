package generation

import (
	"strings"
	"testing"
)

func TestGenerationProvenanceReplayPart01PreservesFrontier(t *testing.T) {
	digest := "sha256:" + strings.Repeat("0", 64)
	pass, err := BuildGenerationProvenanceReplayPart01(
		"execution-plan-1",
		digest,
		digest,
		digest,
		generationProvenancePassPart01,
		"GENERATION_AND_REVERSE_OBSERVED",
		-1,
	)
	if err != nil {
		t.Fatalf("pass build error = %v", err)
	}
	if !pass.ValidPart01() {
		t.Fatalf("pass replay should be valid: %#v", pass)
	}

	unknown, err := BuildGenerationProvenanceReplayPart01(
		"execution-plan-1",
		digest,
		"",
		"",
		generationProvenanceUnknownPart01,
		"MISSING_REVERSE_EVIDENCE",
		2,
	)
	if err != nil {
		t.Fatalf("unknown build error = %v", err)
	}
	if !unknown.ValidPart01() || unknown.MissingStageIndex != 2 {
		t.Fatalf("unknown replay should preserve frontier: %#v", unknown)
	}

	tampered := pass
	tampered.GeneratedDigest = "sha256:" + strings.Repeat("1", 64)
	if tampered.ValidPart01() {
		t.Fatal("tampered generated digest unexpectedly validated")
	}

	tampered = unknown
	tampered.MissingStageIndex = -1
	if tampered.ValidPart01() {
		t.Fatal("tampered unknown frontier unexpectedly validated")
	}
}
