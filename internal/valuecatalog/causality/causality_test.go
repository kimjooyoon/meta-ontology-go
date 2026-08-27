package causality

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestEvaluateSuccess(t *testing.T) {
	receipt, err := Evaluate(fixtureReport(t, ModeSuccess), ModeSuccess)
	if err != nil {
		t.Fatal(err)
	}
	if err := Validate(receipt); err != nil {
		t.Fatal(err)
	}
	if receipt.Metrics.DischargedClaimTotal != 9 || receipt.Metrics.DirectMissingClaimTotal != 0 || receipt.Metrics.DependencyBlockedClaimTotal != 0 {
		t.Fatalf("unexpected success metrics: %+v", receipt.Metrics)
	}
	if receipt.Decision.Value != "PASS" || receipt.Decision.SemanticPromotionAuthorized {
		t.Fatalf("unexpected success decision: %+v", receipt.Decision)
	}
}

func TestEvaluateUnknownSeparatesDirectAndBlocked(t *testing.T) {
	receipt, err := Evaluate(fixtureReport(t, ModeUnknown), ModeUnknown)
	if err != nil {
		t.Fatal(err)
	}
	if err := Validate(receipt); err != nil {
		t.Fatal(err)
	}
	metrics := receipt.Metrics
	if metrics.DirectMissingClaimTotal != 1 || metrics.DependencyBlockedClaimTotal != 8 || metrics.ObservedBlockingEdgeTotal != 11 || metrics.MaximumCausePathDepth != 4 {
		t.Fatalf("unexpected unknown metrics: %+v", metrics)
	}
	if receipt.Resolutions[0].Kind != "DIRECT_MISSING" {
		t.Fatalf("root classification: %+v", receipt.Resolutions[0])
	}
	for _, resolution := range receipt.Resolutions[1:] {
		if resolution.Kind != "DEPENDENCY_BLOCKED" || resolution.Coordinate.Reason != "UPSTREAM_CLAIM_OPEN" {
			t.Fatalf("blocked classification: %+v", resolution)
		}
	}
}

func TestValidateRejectsGraphMutation(t *testing.T) {
	receipt, err := Evaluate(fixtureReport(t, ModeSuccess), ModeSuccess)
	if err != nil {
		t.Fatal(err)
	}
	receipt.Graph.Edges[0].Kind = "MUTATED"
	receipt.Graph.Digest, err = graphDigest(receipt.Graph)
	if err != nil {
		t.Fatal(err)
	}
	receipt.Subject.GraphDigest = receipt.Graph.Digest
	receipt.ReceiptDigest, err = receiptDigest(receipt)
	if err != nil {
		t.Fatal(err)
	}
	if err := Validate(receipt); err == nil {
		t.Fatal("mutated graph accepted")
	}
}

func TestEvaluateRejectsMixedResolution(t *testing.T) {
	data := fixtureReport(t, ModeUnknown)
	var report inputReport
	if err := json.Unmarshal(data, &report); err != nil {
		t.Fatal(err)
	}
	report.OperationClaimTransitions[ClaimTotal].Event = "EVIDENCE_ACCEPTED"
	data, err := json.Marshal(report)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Evaluate(data, ""); err == nil {
		t.Fatal("mixed resolution accepted")
	}
}

func fixtureReport(t *testing.T, mode string) []byte {
	t.Helper()
	report := inputReport{Schema: InputReportSchema}
	for index, axis := range claimAxes {
		claimID := "gooo.claim.operation-spec." + axis
		report.OperationClaimTransitions = append(report.OperationClaimTransitions, inputTransition{
			Sequence:         index + 1,
			ClaimID:          claimID,
			DeclarationDigest: fmt.Sprintf("%064x", index+1),
			Event:            "CLAIM_REGISTERED",
			Before:           "UNRECORDED",
			After:            "OPEN",
			Coordinate:       Coordinate{Stage: "DECLARE", Step: axis, Reason: "CLAIM_REGISTERED"},
			TransitionDigest: fmt.Sprintf("%064x", index+101),
		})
	}
	for index, axis := range claimAxes {
		event := "EVIDENCE_ACCEPTED"
		after := "DISCHARGED"
		coordinate := Coordinate{Stage: "VERIFY", Step: axis, Reason: "EVIDENCE_ACCEPTED"}
		if mode == ModeUnknown {
			event = "EVIDENCE_UNAVAILABLE"
			after = "OPEN"
			coordinate = Coordinate{Stage: "RESOLVE", Step: "resolve-operation-spec", Reason: "VALUE_PROGRAM_UNKNOWN"}
		}
		report.OperationClaimTransitions = append(report.OperationClaimTransitions, inputTransition{
			Sequence:                 index + ClaimTotal + 1,
			ClaimID:                  "gooo.claim.operation-spec." + axis,
			DeclarationDigest:        fmt.Sprintf("%064x", index+1),
			Event:                    event,
			Before:                   "OPEN",
			After:                    after,
			Coordinate:               coordinate,
			EvidenceDigest:           fmt.Sprintf("%064x", index+201),
			PreviousTransitionDigest: fmt.Sprintf("%064x", index+200),
			TransitionDigest:         fmt.Sprintf("%064x", index+301),
		})
	}
	report.OperationClaimTransitionHead = report.OperationClaimTransitions[len(report.OperationClaimTransitions)-1].TransitionDigest
	data, err := json.Marshal(report)
	if err != nil {
		t.Fatal(err)
	}
	return data
}
