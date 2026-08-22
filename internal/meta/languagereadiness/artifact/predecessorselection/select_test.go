package predecessorselection

import "testing"

func TestSelectFailsClosedWithoutCanonicalCandidate(t *testing.T) {
	input := Input{Repository: "owner/repository",
		CurrentHeadSHA: "2222222222222222222222222222222222222222",
		PredecessorSHA: "1111111111111111111111111111111111111111",
		Branch: "dev", Workflow: "Transformation effect ledger"}
	result, err := Select(input)
	if err != nil {
		t.Fatal(err)
	}
	if result.Report.Decision != DecisionFailClosed || result.Report.Reason != ReasonNotFound {
		t.Fatalf("missing predecessor did not fail closed: %+v", result.Report)
	}
}
