package coupling

import (
	"reflect"
	"testing"
)

func TestPresentationAndExpectedMutationIsolation(t *testing.T) {
	row := testCorpus()[0]
	base := Evaluate(row.Input)
	mutations := []func(*Input){
		func(input *Input) { input.FixtureID = "fixture-label/changed" },
		func(input *Input) {
			input.SemanticBefore.Nodes[0].Name = "Renamed"
			input.SemanticBefore.Nodes[0].Aliases = []string{"AliasChanged"}
		},
		func(input *Input) {
			input.SemanticAfter.Nodes[0].Name = "Renamed"
			input.SemanticAfter.Nodes[0].Aliases = []string{"AliasChanged"}
		},
		func(input *Input) {
			input.Registry[0].PackageLabel = "other"
			input.Registry[0].FileLabel = "other.go"
			input.Registry[0].SourceSpan = "99:1-99:2"
		},
	}
	for index, mutate := range mutations {
		input := cloneInput(row.Input)
		mutate(&input)
		got := Evaluate(input)
		if got.InputDigest != base.InputDigest || got.CanonicalOutputDigest != base.CanonicalOutputDigest || got.ReplayDigest != base.ReplayDigest || got.Decision != base.Decision || got.Reason != base.Reason {
			t.Fatalf("presentation mutation %d affected authority result: base=%+v got=%+v", index, base, got)
		}
	}
	mutatedCase := row
	mutatedCase.Name = "case-label-changed"
	mutatedCase.Expected = FixtureExpectation{
		Decision: DecisionFailClosed, Reason: ReasonDigestMismatch,
		AcceptedSurfaces: []string{"expected-only"}, ChangedSurfaces: []string{"expected-only"}, ReceiptSurfaces: []string{"expected-only"},
		ObservationCounts: ObservationCounts{RegistryBindings: 99, ResourceReceipts: 99}, Resources: ResourceObservation{CPUCoreNS: 99, PeakMemoryBytes: 98, WorkUnits: 97},
	}
	if got := Evaluate(mutatedCase.Input); !reflect.DeepEqual(got, base) {
		t.Fatalf("case name or expected-only mutation affected actual result: base=%+v got=%+v", base, got)
	}
	if got := CanonicalInputDigest(row.Input); got == CanonicalInputDigest(func() Input {
		input := cloneInput(row.Input)
		input.SemanticBefore.Nodes[0].ID = "urn:gooo:entity:renamed"
		return input
	}()) {
		t.Fatal("stable identity mutation did not affect input digest")
	}
	authority := cloneInput(row.Input)
	authority.AuthoritySourceAfter += "\n"
	if CanonicalInputDigest(authority) == base.InputDigest {
		t.Fatal("authoritative source mutation did not affect input digest")
	}
}
