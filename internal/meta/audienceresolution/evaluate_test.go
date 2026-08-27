package audienceresolution

import "testing"

func TestEvaluateProjectsNestedAudiencesFromOneLedger(t *testing.T) {
	receipt := Evaluate(fixtureInput(t))
	if receipt.Decision != "PASS" || receipt.Resolution != "EXACT" || receipt.Reason != "AUDIENCE_RESOLUTION_OBSERVED" {
		t.Fatalf("receipt decision=%s resolution=%s reason=%s", receipt.Decision, receipt.Resolution, receipt.Reason)
	}
	if receipt.Summary.Coordinates.Satisfied != 12 || receipt.Summary.Coordinates.Total != 12 || receipt.Summary.Coordinates.BasisPoints != 10000 {
		t.Fatalf("summary=%+v", receipt.Summary)
	}
	if len(receipt.Views) != 3 || receipt.Views[0].Satisfied != 4 || receipt.Views[0].Total != 4 ||
		receipt.Views[1].Satisfied != 8 || receipt.Views[1].Total != 8 || receipt.Views[2].Satisfied != 12 || receipt.Views[2].Total != 12 {
		t.Fatalf("views=%+v", receipt.Views)
	}
	if receipt.Views[0].OmittedCoordinateCount != 8 || receipt.Views[1].OmittedCoordinateCount != 4 || receipt.Views[2].OmittedCoordinateCount != 0 {
		t.Fatalf("omission counts=%d/%d/%d", receipt.Views[0].OmittedCoordinateCount, receipt.Views[1].OmittedCoordinateCount, receipt.Views[2].OmittedCoordinateCount)
	}
	for _, view := range receipt.Views {
		if view.Decision != "PASS" || view.Reason != receipt.Reason {
			t.Fatalf("view contradicts receipt: %+v", view)
		}
	}
	if err := ValidateReceipt(receipt); err != nil {
		t.Fatalf("independent validation failed: %v", err)
	}
}

func TestMissingCoordinateLowersEveryAudienceDecision(t *testing.T) {
	input := fixtureInput(t)
	input.Ledger.Records = input.Ledger.Records[:len(input.Ledger.Records)-1]
	input.Replay = input.Ledger
	receipt := Evaluate(input)
	if receipt.Decision != "FAIL_CLOSED" || receipt.Resolution != "LOWER_RESOLUTION" || receipt.Reason != "AUDIENCE_EVIDENCE_MISSING" {
		t.Fatalf("receipt=%+v", receipt)
	}
	if receipt.Views[0].Decision != "FAIL_CLOSED" || receipt.Views[1].Decision != "FAIL_CLOSED" || receipt.Views[2].Decision != "FAIL_CLOSED" {
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
	input.Ledger.Records[0].Decision = "FAIL_CLOSED"
	input.Replay = input.Ledger
	receipt := Evaluate(input)
	if receipt.Decision != "FAIL_CLOSED" || receipt.Resolution != "INVARIANT_ONLY" || receipt.Reason != "AUDIENCE_DECISION_CONTRADICTION" {
		t.Fatalf("receipt=%+v", receipt)
	}
	for _, view := range receipt.Views {
		if view.Decision != "FAIL_CLOSED" {
			t.Fatalf("contradictory view escaped fail-closed: %+v", view)
		}
	}
	if err := ValidateReceipt(receipt); err != nil {
		t.Fatalf("contradiction receipt should remain independently valid: %v", err)
	}
}

func TestReplayDivergenceCannotBecomePass(t *testing.T) {
	input := fixtureInput(t)
	input.Replay.Records[0].Reason = "different replay"
	receipt := Evaluate(input)
	if receipt.Decision != "FAIL_CLOSED" || receipt.Resolution != "INVARIANT_ONLY" || receipt.Summary.Coordinates.Satisfied != 11 {
		t.Fatalf("receipt=%+v", receipt)
	}
}
