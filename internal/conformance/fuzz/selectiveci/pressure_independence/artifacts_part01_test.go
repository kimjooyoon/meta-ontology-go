package pressureindependence

import (
	"testing"
)

type artifactRole struct {
	name string
	data []byte
	set  func(*Input, string)
	get  func(Input) string
}

func TestSyntheticArtifactBindings(t *testing.T) {
	input := mustCorpusInput(t, "two-independent-groups-pass")
	bindSyntheticArtifacts(&input)
	if !verifyArtifactBindings(input) {
		t.Fatal("correct five-way artifact binding was rejected")
	}
	if got := Evaluate(input); got.Decision != DecisionPass {
		t.Fatalf("correct five-way binding = %#v", got)
	}
	for _, role := range syntheticArtifactRoles() {
		expected := artifactDigest(role.data)
		if role.get(input) != expected {
			t.Fatalf("%s digest was not recomputed from immutable bytes", role.name)
		}
		mutated := append([]byte(nil), role.data...)
		mutated[0] ^= 1
		mutatedDigest := artifactDigest(mutated)
		if mutatedDigest == expected {
			t.Fatalf("%s one-byte mutation did not change digest", role.name)
		}
		mutatedInput := input
		role.set(&mutatedInput, mutatedDigest)
		if got := Evaluate(mutatedInput); got.Decision != DecisionUnknown || got.Reason != ReasonStaleDigest {
			t.Fatalf("%s mutated artifact = %#v", role.name, got)
		}
		arbitraryDigest := artifactDigest([]byte("shape-valid-arbitrary-" + role.name))
		if !validDigest(arbitraryDigest) || arbitraryDigest == expected {
			t.Fatalf("%s arbitrary digest was not shape-valid and distinct", role.name)
		}
		arbitraryInput := input
		role.set(&arbitraryInput, arbitraryDigest)
		if got := Evaluate(arbitraryInput); got.Decision != DecisionUnknown || got.Reason != ReasonStaleDigest {
			t.Fatalf("%s arbitrary digest = %#v", role.name, got)
		}
		t.Logf("%s=%s", role.name, expected)
	}
}
func TestArtifactSemanticFieldsBindInput(t *testing.T) {
	base := mustCorpusInput(t, "two-independent-groups-pass")
	mutations := []struct {
		name   string
		mutate func(*Input)
	}{
		{name: "requested_K", mutate: func(input *Input) { input.RequestedK++ }},
		{name: "minimum_independent", mutate: func(input *Input) { input.MinimumIndependent++ }},
		{name: "fixture_id", mutate: func(input *Input) { input.FixtureID = "other-fixture" }},
	}
	for _, mutation := range mutations {
		input := base
		mutation.mutate(&input)
		if got := Evaluate(input); got.Decision != DecisionUnknown || got.Reason != ReasonStaleDigest {
			t.Fatalf("%s mutation = %#v", mutation.name, got)
		}
	}
}
