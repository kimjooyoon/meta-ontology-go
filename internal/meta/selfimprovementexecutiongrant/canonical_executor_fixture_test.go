package selfimprovementexecutiongrant

import (
	"reflect"
	"testing"
)

func TestCanonicalExecutorFixtureAdversarialCopyIsPure(t *testing.T) {
	original := CanonicalExecutorGrantFixture{
		Verification: CanonicalExecutorVerification{
			Counts: map[string]int{"CLOSED": 1},
			Cases: []CanonicalExecutorVerificationCase{{MissingFields: []string{"candidate_input_digest"}}},
		},
		Manifest: CanonicalExecutorBindingManifest{ArtifactNames: []string{"canonical-executor-grant-request.json"}},
	}
	mutated := cloneCanonicalExecutorFixture(original)
	mutated.Manifest.ArtifactNames[0] = "substituted.json"
	mutated.Verification.Counts["CLOSED"] = 99
	mutated.Verification.Cases[0].MissingFields[0] = "source_artifact"
	if original.Manifest.ArtifactNames[0] != "canonical-executor-grant-request.json" || original.Verification.Counts["CLOSED"] != 1 || original.Verification.Cases[0].MissingFields[0] != "candidate_input_digest" {
		t.Fatal("adversarial fixture copy mutated the source fixture")
	}
}

func TestCanonicalExecutorFixtureVerificationRepeatsWithoutMutation(t *testing.T) {
	fixture := CanonicalExecutorGrantFixture{
		Verification: CanonicalExecutorVerification{Counts: map[string]int{}},
		Manifest:     CanonicalExecutorBindingManifest{ArtifactNames: []string{"canonical-executor-grant-request.json"}},
	}
	original := cloneCanonicalExecutorFixture(fixture)
	first := VerifyCanonicalExecutorGrantFixture(PolicyProgram{}, fixture)
	second := VerifyCanonicalExecutorGrantFixture(PolicyProgram{}, fixture)
	if !reflect.DeepEqual(first, second) {
		t.Fatal("repeated canonical fixture verification produced different results")
	}
	if !reflect.DeepEqual(fixture, original) {
		t.Fatal("canonical fixture verification mutated its input")
	}
}
