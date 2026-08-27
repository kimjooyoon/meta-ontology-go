package proposalpromotion

import "testing"

func TestGuardrailsFailClosed(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*Source)
	}{
		{"ambiguity", func(source *Source) { source.Selection.AmbiguousCandidates = 1 }},
		{"mutation authority", func(source *Source) {
			source.Selection.SelectedPromotionAuthorized = true
		}},
	}
	for _, test := range tests {
		source := validSource()
		test.mutate(&source)
		receipt := evaluate(testCurrent, testEvidence, source, testObservationEvidence())
		if receipt.Decision != DecisionFailClosed || Validate(receipt, receipt.Repository, testCurrent, testEvidence) == nil {
			t.Fatalf("%s receipt = %+v", test.name, receipt)
		}
	}
}

func TestDigestTamperFailsClosed(t *testing.T) {
	receipt := evaluate(testCurrent, testEvidence, validSource(), testObservationEvidence())
	receipt.Source.Selection.RunAttempt++
	if Validate(receipt, receipt.Repository, testCurrent, testEvidence) == nil {
		t.Fatal("tampered receipt accepted")
	}
}
