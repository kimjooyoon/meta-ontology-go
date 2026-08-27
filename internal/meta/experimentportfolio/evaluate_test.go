package experimentportfolio

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func validInput() Input {
	contract := ExpectedContract()
	receipts := make([]Receipt, 0, len(contract.Candidates))
	for _, candidate := range contract.Candidates {
		receipt, err := ProduceReceipt(strings.Repeat("a", 40), candidate.SourcePath, []byte(candidate.ID), candidate.ID)
		if err != nil {
			panic(err)
		}
		receipts = append(receipts, receipt)
	}
	return Input{SubjectSHA: strings.Repeat("a", 40), Contract: contract, Receipts: receipts}
}

func TestEvaluatePreservesEveryCandidateVectorAndEvidence(t *testing.T) {
	input := validInput()
	report := Evaluate(input)
	if report.Decision != "PORTFOLIO_PRESERVED" || report.Resolution != "EXACT" || report.Reason != "NO_WINNER_NO_AGGREGATE" {
		t.Fatalf("report identity = %#v", report)
	}
	if len(report.Candidates) != ExpectedCandidates || report.Summary.CoordinatesPerCandidate != ExpectedCoordinates {
		t.Fatalf("candidate summary = %#v", report.Summary)
	}
	for index, candidate := range report.Candidates {
		receipt := input.Receipts[index]
		if !reflect.DeepEqual(candidate.Receipt.CoordinateVector, candidate.CoordinateVector) ||
			!reflect.DeepEqual(candidate.Receipt.Counterexamples, candidate.Counterexamples) ||
			!reflect.DeepEqual(candidate.Receipt.UnknownLocations, candidate.UnknownLocations) {
			t.Fatalf("candidate %q did not preserve receipt evidence", candidate.CandidateID)
		}
		if candidate.CounterexampleCount != len(receipt.Counterexamples) || candidate.UnknownLocationCount != len(receipt.UnknownLocations) {
			t.Fatalf("candidate %q count changed", candidate.CandidateID)
		}
	}
	if report.Summary.CounterexampleCounts["derive"] != 2 || report.Summary.CounterexampleCounts["replay"] != 1 || report.Summary.CounterexampleCounts["reflect"] != 0 {
		t.Fatalf("counterexample counts = %#v", report.Summary.CounterexampleCounts)
	}
	for _, proof := range report.Proofs {
		if !proof.Passed {
			t.Fatalf("proof failed: %#v", proof)
		}
	}
	if report.RepositoryWrites != 0 || report.MutationAuthority {
		t.Fatalf("read-only effects = %d/%v", report.RepositoryWrites, report.MutationAuthority)
	}
}

func TestEvaluateRejectsForgedCoordinateDigest(t *testing.T) {
	input := validInput()
	input.Receipts[0].CoordinateVector[0].Numerator = 0
	report := Evaluate(input)
	if report.Decision != "FAIL_CLOSED" || report.Reason != "PORTFOLIO_RECEIPT_DIGEST_INVALID" || report.Summary.Unknowns != 1 {
		t.Fatalf("forged report = %#v", report)
	}
}

func TestEvaluateEmitsNoAggregateScoreOrWinner(t *testing.T) {
	report := Evaluate(validInput())
	data, err := json.Marshal(report)
	if err != nil {
		t.Fatal(err)
	}
	encoded := string(data)
	if strings.Contains(encoded, `"score"`) || strings.Contains(encoded, `"winner"`) || strings.Contains(encoded, `"basis_points"`) {
		t.Fatalf("aggregate field emitted: %s", encoded)
	}
}
