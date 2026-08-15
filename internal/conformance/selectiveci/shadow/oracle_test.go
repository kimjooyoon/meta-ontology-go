package shadow

import (
	"reflect"
	"testing"
)

const (
	phaseACorpusDigest         = "e79ba3696eec2bb67c915398a1f652f523b9b98a3227bee2e4e2c4b9f2f8120e"
	phaseAExpectedVectorDigest = "a9661672d1dcf30df297b8aae90d2b7138ef7126dccf9f45b495a5399dd82c58"
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
