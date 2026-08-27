package claimdependencyjudge

import "testing"

func TestRawConsumerRejectsMissingSourceAndObservation(t *testing.T) {
	if _, err := Judge(nil, "missing.gooo", nil, nil, nil); err == nil {
		t.Fatal("raw consumer accepted missing evidence")
	}
}
