package languagereadiness

import (
	"encoding/json"
	"strconv"
	"testing"
)

const improvementTestRegistryDigest = "sha256:0000000000000000000000000000000000000000000000000000000000000000"

func TestEvaluateImprovementTransitionUsesOnlyComparableIntegers(t *testing.T) {
	before := improvementTestSnapshot(7)
	after := improvementTestSnapshot(8)

	got := EvaluateImprovementTransition(before, after)
	if got.Decision != ImprovementImproved || got.ReasonCode != "IMPROVEMENT_PROVEN" {
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
		t.Fatalf("contract shape = comparable:%t indicators:%d proofs:%d", got.Comparable, len(got.Indicators), len(got.Proofs))
	}
}

func TestEvaluateImprovementTransitionDoesNotInferUnknownEvidence(t *testing.T) {
	before := improvementTestSnapshot(7)
	after := improvementTestSnapshot(8)
	after.Evidence[23].Status = "LIKELY_SATISFIED"

	got := EvaluateImprovementTransition(before, after)
	if got.Decision != ImprovementLowerResolution || got.ReasonCode != "AFTER_EVIDENCE_STATUS_UNKNOWN" {
		t.Fatalf("decision = %q, reason = %q", got.Decision, got.ReasonCode)
	}
}

func TestEvaluateImprovementTransitionRejectsMovingDenominators(t *testing.T) {
	before := improvementTestSnapshot(7)
	after := improvementTestSnapshot(8)
	after.RegistryDigest = "sha256:1111111111111111111111111111111111111111111111111111111111111111"

	got := EvaluateImprovementTransition(before, after)
	if got.Decision != ImprovementLowerResolution || got.ReasonCode != "REGISTRY_DIGEST_MISMATCH" {
		t.Fatalf("decision = %q, reason = %q", got.Decision, got.ReasonCode)
	}
}

func TestEvaluateImprovementTransitionSeparatesNoChangeAndRegression(t *testing.T) {
	noChange := EvaluateImprovementTransition(improvementTestSnapshot(7), improvementTestSnapshot(7))
	if noChange.Decision != ImprovementNoChange || noChange.CompletedDelta != 0 || noChange.BasisPointsDelta != 0 {
		t.Fatalf("no change = %#v", noChange)
	}

	regressed := EvaluateImprovementTransition(improvementTestSnapshot(7), improvementTestSnapshot(6))
	if regressed.Decision != ImprovementRegressed || regressed.CompletedDelta != -1 || regressed.Regressions != 1 {
		t.Fatalf("regression = %#v", regressed)
	}
}

func TestEvaluateImprovementTransitionReplaysDeterministically(t *testing.T) {
	before := improvementTestSnapshot(7)
	after := improvementTestSnapshot(8)
	first := EvaluateImprovementTransition(before, after)
	second := EvaluateImprovementTransition(before, after)
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

func improvementTestSnapshot(completed int64) ImprovementSnapshot {
	evidence := make([]ImprovementEvidence, 24)
	for index := range evidence {
		status := ImprovementNotSatisfied
		if int64(index) < completed {
			status = ImprovementSatisfied
		}
		evidence[index] = ImprovementEvidence{
			ID:     "obligation-" + strconv.Itoa(index+1),
			Status: status,
		}
	}
	return ImprovementSnapshot{
		ContractSchema: ImprovementSnapshotSchema,
		RegistryDigest: improvementTestRegistryDigest,
		Completed:      completed,
		Total:          int64(len(evidence)),
		BasisPoints:    completed * 10_000 / int64(len(evidence)),
		Evidence:       evidence,
	}
}
