package feedbackpredecessor

import "testing"

func TestSelectEmitsNonPromotingBaselineForUnsuccessfulRun(t *testing.T) {
	input := predecessorFixture()
	input.Candidates[0].Conclusion = "failure"
	report, err := Select(input)
	if err != nil {
		t.Fatal(err)
	}
	if report.Decision != DecisionLower || report.Reason != ReasonUnsuccessful ||
		report.Resolution != ResolutionClass || report.NextOperation != OperationReevaluate ||
		report.PromotionAuthorized || report.Selected != nil || !Consumable(report) {
		t.Fatalf("baseline = %#v", report)
	}
	if report.Summary.SuccessfulCandidates != 0 || report.Summary.ReceiptBoundCandidates != 1 ||
		report.Indicators[0].Value != 0 || report.Indicators[1].Value != 10000 {
		t.Fatalf("baseline evidence = %#v", report)
	}
}
