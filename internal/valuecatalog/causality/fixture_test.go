package causality

import (
	"encoding/json"
	"fmt"
	"testing"
)

func fixtureReport(t *testing.T, mode string) []byte {
	t.Helper()
	report := inputReport{Schema: InputReportSchema}
	for index, axis := range claimAxes {
		claimID := "gooo.claim.operation-spec." + axis
		report.OperationClaimTransitions = append(report.OperationClaimTransitions, inputTransition{
			Sequence:          index + 1,
			ClaimID:           claimID,
			DeclarationDigest: fmt.Sprintf("%064x", index+1),
			Event:             "CLAIM_REGISTERED",
			Before:            "UNRECORDED",
			After:             "OPEN",
			Coordinate:        Coordinate{Stage: "DECLARE", Step: axis, Reason: "CLAIM_REGISTERED"},
			TransitionDigest:  fmt.Sprintf("%064x", index+101),
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
