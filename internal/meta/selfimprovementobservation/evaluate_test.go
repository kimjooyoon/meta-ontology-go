package selfimprovementobservation

import "testing"

func TestBuildBindsReadOnlyImprovementInput(t *testing.T) {
	in, opts := validFixture()
	observation := Build(in, opts)
	if observation.Decision != "OBSERVED" || observation.Resolution != "EXACT" {
		t.Fatalf("decision/resolution = %s/%s", observation.Decision, observation.Resolution)
	}
	if observation.Summary.Coordinates.Satisfied != 16 || observation.Summary.Coordinates.Total != 16 {
		t.Fatalf("coordinates = %#v", observation.Summary.Coordinates)
	}
	if observation.Summary.CandidateCount != 0 || observation.Authority != (Authority{}) {
		t.Fatalf("candidate/authority = %d/%#v", observation.Summary.CandidateCount, observation.Authority)
	}
	if len(observation.Artifacts) != 3 || len(observation.Views) != 3 || len(observation.Proofs) != 3 {
		t.Fatalf("artifacts/views/proofs = %d/%d/%d", len(observation.Artifacts), len(observation.Views), len(observation.Proofs))
	}
	if !ValidObservationDigest(observation) {
		t.Fatal("observation digest is not content-bound")
	}
}

func TestBuildIsDeterministic(t *testing.T) {
	in, opts := validFixture()
	first, second := Build(in, opts), Build(in, opts)
	if first.Digest != second.Digest || first.InputDigest != second.InputDigest {
		t.Fatalf("replay drift = %s/%s", first.Digest, second.Digest)
	}
}
