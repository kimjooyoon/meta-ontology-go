package shadow

import (
	"reflect"
	"testing"
)

const (
	phaseACorpusDigest         = "1749e4a01627483ca9b3f6ecb20e83244abb729b24501600dbe2ee553d295ca3"
	phaseAExpectedVectorDigest = "c48741ac3ba78be5cbd4ede9df04c962e32da0ba2dc761be79c2829749aad213"
	priorCorpusDigest          = "8448b309f64a05c06f75f03352d7516dcb296182af4b922532c28677353ca01e"
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
