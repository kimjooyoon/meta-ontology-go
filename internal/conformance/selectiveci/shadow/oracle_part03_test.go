package shadow

import (
	"testing"
)

func TestSecondCorrectionRecordBindsLaneRegistryOmission(t *testing.T) {
	record, err := LoadSecondCorrection()
	if err != nil {
		t.Fatal(err)
	}
	if record.ReasonCode != "LANE_REGISTRY_BINDING_OMISSION" {
		t.Fatalf("second correction reason = %q", record.ReasonCode)
	}
	if record.Superseded.CorpusDigest != "36359077392431f4e4136baeb022b78f87fdf7c69a0dbab18ca38e3e92ae6954" || record.Superseded.ExpectedVectorDigest != "fe260bba00c58fb3ab761910c253905dfb749be60ed655b03810bebddc2b3ef5" {
		t.Fatalf("second correction predecessor evidence = %#v", record.Superseded)
	}
	if record.Corrected.CorpusDigest != priorCorpusDigest || record.Corrected.ExpectedVectorDigest != phaseAExpectedVectorDigest {
		t.Fatalf("second correction evidence = %#v", record.Corrected)
	}
}
func TestThirdCorrectionRecordBindsExpectedDigestIndependence(t *testing.T) {
	record, err := LoadThirdCorrection()
	if err != nil {
		t.Fatal(err)
	}
	if record.ReasonCode != "EXPECTED_VALUE_DIGEST_ECHO" {
		t.Fatalf("third correction reason = %q", record.ReasonCode)
	}
	if record.Superseded.CorpusDigest != priorCorpusDigest || record.Superseded.ExpectedVectorDigest != phaseAExpectedVectorDigest {
		t.Fatalf("third correction predecessor evidence = %#v", record.Superseded)
	}
	if record.Corrected.CorpusDigest != CorpusDigest() || record.Corrected.ExpectedVectorDigest != phaseAExpectedVectorDigest {
		t.Fatalf("third correction evidence = %#v", record.Corrected)
	}
}
func TestCorpusDigestsAreStableAndDistinct(t *testing.T) {
	corpus, err := LoadCorpus()
	if err != nil {
		t.Fatal(err)
	}
	if got := CorpusDigest(); got != phaseACorpusDigest {
		t.Fatalf("corpus digest changed: got %s want %s", got, phaseACorpusDigest)
	}
	if got := ExpectedVectorDigest(corpus); got != phaseAExpectedVectorDigest {
		t.Fatalf("expected vector digest changed: got %s want %s", got, phaseAExpectedVectorDigest)
	}
}
