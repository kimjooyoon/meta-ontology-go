package improvement

import (
	"encoding/json"
	"testing"
)

func TestEvaluateUsesOnlyComparableIntegers(t *testing.T) {
	got := Evaluate(testSnapshot(7), testSnapshot(8))
	if got.Decision != Improved || got.ReasonCode != "IMPROVEMENT_PROVEN" {
		t.Fatalf("decision = %q, reason = %q", got.Decision, got.ReasonCode)
	}
	if got.BeforeCompleted != 7 || got.AfterCompleted != 8 || got.Total != 24 {
		t.Fatalf("completed = %d/%d -> %d/%d", got.BeforeCompleted, got.Total, got.AfterCompleted, got.Total)
	}
	if got.CompletedDelta != 1 || got.BasisPointsDelta != 417 {
		t.Fatalf("deltas = completed:%d bps:%d", got.CompletedDelta, got.BasisPointsDelta)
	}
	if got.Gains != 1 || got.Regressions != 0 || got.BeforeUnresolved != 0 || got.AfterUnresolved != 0 {
		t.Fatalf("guardrails = gains:%d regressions:%d unresolved:%d/%d", got.Gains, got.Regressions, got.BeforeUnresolved, got.AfterUnresolved)
	}
	if !got.Comparable || len(got.Indicators) != 5 || len(got.Proofs) != 4 {
		t.Fatalf("shape = comparable:%t indicators:%d proofs:%d", got.Comparable, len(got.Indicators), len(got.Proofs))
	}
}

func TestEvaluateReplaysDeterministically(t *testing.T) {
	first := Evaluate(testSnapshot(7), testSnapshot(8))
	second := Evaluate(testSnapshot(7), testSnapshot(8))
	firstJSON, err := json.Marshal(first)
	if err != nil {
		t.Fatal(err)
	}
	secondJSON, err := json.Marshal(second)
	if err != nil {
		t.Fatal(err)
	}
	if string(firstJSON) != string(secondJSON) || first.Digest != second.Digest {
		t.Fatalf("replay mismatch: %s != %s", first.Digest, second.Digest)
	}
}
