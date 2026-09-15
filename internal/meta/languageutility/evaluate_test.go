package languageutility

import "testing"

func TestEvaluateQuantifiesUtilityWithoutClaimingCompleteness(t *testing.T) {
	contract := fixtureContract()
	report, err := Evaluate(contract, fixtureObservation(contract))
	if err != nil {
		t.Fatal(err)
	}
	if report.Decision != "PROGRESS_OBSERVED" || report.Resolution != "EXACT" {
		t.Fatalf("decision = %s/%s", report.Decision, report.Resolution)
	}
	got := report.Summary
	if got.ClosedCells != 39 || got.CellsTotal != 42 || got.RemainingCells != 3 ||
		got.ProgressBasisPoints != 9285 || got.CompleteUseCases != 4 || got.UseCasesTotal != 6 {
		t.Fatalf("summary = %#v", got)
	}
	if got.UnknownCells != 0 || got.RefutedCells != 0 || got.UtilityComplete || got.PromotionComplete {
		t.Fatalf("completeness = %#v", got)
	}
	if len(report.Proofs) != 3 || report.Proofs[2].Choice != "regression" ||
		report.Proofs[2].Closed != 9 || report.Proofs[2].Total != 12 {
		t.Fatalf("proofs = %#v", report.Proofs)
	}
}

func TestEvaluateIsDeterministic(t *testing.T) {
	contract := fixtureContract()
	observation := fixtureObservation(contract)
	first, _ := Evaluate(contract, observation)
	second, _ := Evaluate(contract, observation)
	firstRaw, _ := MarshalReport(first)
	secondRaw, _ := MarshalReport(second)
	if string(firstRaw) != string(secondRaw) {
		t.Fatal("language utility report replay differs")
	}
}
