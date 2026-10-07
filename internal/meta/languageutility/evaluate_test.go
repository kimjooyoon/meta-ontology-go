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
	if got.ClosedCells != 45 || got.CellsTotal != 49 || got.RemainingCells != 4 ||
		got.ProgressBasisPoints != 9183 || got.CompleteUseCases != 4 || got.UseCasesTotal != 7 {
		t.Fatalf("summary = %#v", got)
	}
	if got.UnknownCells != 0 || got.RefutedCells != 0 || got.UtilityComplete || got.PromotionComplete {
		t.Fatalf("completeness = %#v", got)
	}
	if len(report.Proofs) != 3 || report.Proofs[2].Choice != "regression" ||
		report.Proofs[2].Closed != 10 || report.Proofs[2].Total != 14 {
		t.Fatalf("proofs = %#v", report.Proofs)
	}
	for _, cell := range report.Cells {
		if cell.UseCaseID != "capability-discovery" {
			continue
		}
		if cell.StageID == "RESOURCE_OBSERVED" && cell.State != StateOpen {
			t.Fatalf("discovery resource cell = %#v, want explicit open state", cell)
		}
		if cell.StageID == "USER_ARTIFACT_VERIFIED" && cell.State != StateClosed {
			t.Fatalf("discovery report artifact cell = %#v, want closed state", cell)
		}
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
