package shadow

import "testing"

func TestBothCorrectionRecordsRemainAvailable(t *testing.T) {
	if _, err := LoadCorrection(); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadSecondCorrection(); err != nil {
		t.Fatal(err)
	}
}
