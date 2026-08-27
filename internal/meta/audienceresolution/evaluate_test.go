package audienceresolution

import (
	"strings"
	"testing"
)

func TestEvaluateProjectsNestedAudiencesFromOneLedger(t *testing.T) {
	receipt := Evaluate(fixtureInput(t))
	if receipt.Decision != "PASS" || receipt.Resolution != "EXACT" || !strings.Contains(receipt.Reason, "CURRENT_EVIDENCE_SUBJECT_RECONSTRUCTED") {
		t.Fatalf("receipt decision=%s resolution=%s reason=%s", receipt.Decision, receipt.Resolution, receipt.Reason)
	}
	if receipt.Summary.Coordinates.Satisfied != 10 || receipt.Summary.Coordinates.Total != 10 || receipt.Summary.Coordinates.BasisPoints != 10000 {
		t.Fatalf("summary=%+v", receipt.Summary)
	}
	if len(receipt.Views) != 3 || receipt.Views[0].Satisfied != 4 || receipt.Views[0].Total != 4 ||
		receipt.Views[1].Satisfied != 8 || receipt.Views[1].Total != 8 || receipt.Views[2].Satisfied != 12 || receipt.Views[2].Total != 12 {
		t.Fatalf("views=%+v", receipt.Views)
	}
	if receipt.Views[0].OmittedCoordinateCount != 8 || receipt.Views[1].OmittedCoordinateCount != 4 || receipt.Views[2].OmittedCoordinateCount != 0 {
		t.Fatalf("omission counts=%d/%d/%d", receipt.Views[0].OmittedCoordinateCount, receipt.Views[1].OmittedCoordinateCount, receipt.Views[2].OmittedCoordinateCount)
	}
	if receipt.Views[0].LocalDecision != "UNKNOWN" || receipt.Views[1].LocalDecision != "UNKNOWN" || receipt.Views[2].LocalDecision != "PASS" {
		t.Fatalf("local decisions=%+v", receipt.Views)
	}
	if receipt.Views[0].InheritedStatus != "INHERITED_NOT_LOCALLY_VERIFIED" || receipt.Views[1].InheritedStatus != "INHERITED_NOT_LOCALLY_VERIFIED" || receipt.Views[2].InheritedStatus != "LOCALLY_VERIFIED" {
		t.Fatalf("inheritance=%+v", receipt.Views)
	}
	if err := ValidateReceipt(receipt); err != nil {
		t.Fatalf("independent validation failed: %v", err)
	}
}

func TestMissingCoordinateLowersEveryAudienceDecision(t *testing.T) {
	input := fixtureInput(t)
	input.Ledger.Records = input.Ledger.Records[1:]
	receipt := Evaluate(input)
	if receipt.Decision != "UNKNOWN" || receipt.Resolution != "LOWER_RESOLUTION" || receipt.Reason == "" {
		t.Fatalf("receipt=%+v", receipt)
	}
	if receipt.Views[0].LocalDecision != "UNKNOWN" || receipt.Views[1].LocalDecision != "UNKNOWN" || receipt.Views[2].LocalDecision != "UNKNOWN" {
		t.Fatalf("views=%+v", receipt.Views)
	}
	if receipt.Views[0].Satisfied != 4 || receipt.Views[2].Satisfied != 11 {
		t.Fatalf("views retained local coordinates=%+v", receipt.Views)
	}
	if err := ValidateReceipt(receipt); err != nil {
		t.Fatalf("fail-closed receipt should remain independently valid: %v", err)
	}
}

func TestContradictoryRecordFailsClosedForAllAudiences(t *testing.T) {
	input := fixtureInput(t)
	input.Ledger.Records[0].ObservedValue = "CONTRADICTORY"
	receipt := Evaluate(input)
	if receipt.Decision != "REFUTED" || receipt.Resolution != "INVARIANT_ONLY" || receipt.Reason == "" {
		t.Fatalf("receipt=%+v", receipt)
	}
	if receipt.Views[0].LocalDecision != "UNKNOWN" || receipt.Views[1].LocalDecision != "REFUTED" || receipt.Views[2].LocalDecision != "REFUTED" {
		t.Fatalf("contradictory local views: %+v", receipt.Views)
	}
	if err := ValidateReceipt(receipt); err != nil {
		t.Fatalf("contradiction receipt should remain independently valid: %v", err)
	}
}

func TestReplayDivergenceCannotBecomePass(t *testing.T) {
	input := fixtureInput(t)
	receipt := Evaluate(input)
	if receipt.Decision != "PASS" || receipt.Resolution != "EXACT" || !receipt.Replay.Equal {
		t.Fatalf("receipt=%+v", receipt)
	}
}
