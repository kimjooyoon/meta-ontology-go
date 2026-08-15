package shadow

import (
	"reflect"
	"testing"
)

const (
	phaseACorpusDigest         = "8448b309f64a05c06f75f03352d7516dcb296182af4b922532c28677353ca01e"
	phaseAExpectedVectorDigest = "c48741ac3ba78be5cbd4ede9df04c962e32da0ba2dc761be79c2829749aad213"
)

func TestCorpusMatchesIndependentOracle(t *testing.T) {
	corpus, err := LoadCorpus()
	if err != nil {
		t.Fatal(err)
	}
	if len(corpus.Cases) < 20 {
		t.Fatalf("corpus has %d cases, want at least 20", len(corpus.Cases))
	}
	seenPartitions := map[string]bool{}
	for _, fixture := range corpus.Cases {
		if fixture.Name == "" || fixture.Partition == "" {
			t.Fatalf("fixture has incomplete identity: %#v", fixture)
		}
		seenPartitions[fixture.Partition] = true
		got := Evaluate(fixture)
		if !reflect.DeepEqual(got, fixture.Expected) {
			t.Errorf("case %q mismatch:\n got  %#v\n want %#v", fixture.Name, got, fixture.Expected)
		}
		if got.ExecutionAuthorized {
			t.Errorf("case %q authorized execution", fixture.Name)
		}
		if got.Status == FullSuiteFallback && (got.ExecutionAuthorized || len(got.SelectedCommandIDs) != 0 || len(got.SelectedGuardIDs) != 0 || len(got.SelectedWorkIDs) != 0 || len(got.SelectedArgv) != 0) {
			t.Errorf("case %q fallback leaked selection or authorization: %#v", fixture.Name, got)
		}
	}
	for _, stage := range []string{StageInput, StageSnapshot, StageRegistry, StagePlan, StagePlanProof, StageProofFail, StageProofUnknown, StageLaneUnknown, StageLaneIneligible, StageSelective} {
		found := false
		for _, fixture := range corpus.Cases {
			if fixture.Expected.Stage == stage {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("corpus does not cover stage %s", stage)
		}
	}
	if !seenPartitions["argv is never executed"] {
		t.Fatal("corpus lacks command-injection-looking argv partition")
	}
}

func TestCorrectionRecordBindsSupersededAndCorrectedEvidence(t *testing.T) {
	record, err := LoadCorrection()
	if err != nil {
		t.Fatal(err)
	}
	if record.ReasonCode != "EXECUTION_AUTHORITY_CONTRADICTION" {
		t.Fatalf("correction reason = %q", record.ReasonCode)
	}
	if record.Superseded.CorpusDigest != "e79ba3696eec2bb67c915398a1f652f523b9b98a3227bee2e4e2c4b9f2f8120e" || record.Superseded.ExpectedVectorDigest != "a9661672d1dcf30df297b8aae90d2b7138ef7126dccf9f45b495a5399dd82c58" {
		t.Fatalf("superseded correction evidence = %#v", record.Superseded)
	}
	if record.Corrected.CorpusDigest != "36359077392431f4e4136baeb022b78f87fdf7c69a0dbab18ca38e3e92ae6954" || record.Corrected.ExpectedVectorDigest != "fe260bba00c58fb3ab761910c253905dfb749be60ed655b03810bebddc2b3ef5" {
		t.Fatalf("corrected correction evidence = %#v", record.Corrected)
	}
}

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
	if record.Corrected.CorpusDigest != CorpusDigest() || record.Corrected.ExpectedVectorDigest != phaseAExpectedVectorDigest {
		t.Fatalf("second correction evidence = %#v", record.Corrected)
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
