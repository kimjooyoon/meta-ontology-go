package verify

import "testing"

func TestJudgeRequiresAReceiptAndRawObservations(t *testing.T) {
	if _, err := Judge(nil, "fixture.gooo", nil, nil, nil, nil, nil); err == nil {
		t.Fatal("judge accepted an empty receipt")
	}
}
