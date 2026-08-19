package bidir

import (
	"reflect"
	"testing"
)

func TestOutputPortOrderSurvivesModelPermutationsAndRepeats(t *testing.T) {
	document := sourceOrderedOutputDocument()
	base, err := Get(document)
	if err != nil {
		t.Fatal(err)
	}
	permuted := base.Clone()
	reverseRelations(permuted.Relations)
	basePorts, _ := orderedSequences(base)
	permutedPorts, _ := orderedSequences(permuted)
	if !reflect.DeepEqual(basePorts, permutedPorts) || sequenceHash(basePorts) != sequenceHash(permutedPorts) {
		t.Fatalf("output port order was not source-authoritative: %v != %v", basePorts, permutedPorts)
	}
	want := []ID{"billing://entity/zebra", "billing://entity/apple"}
	for repeat := 0; repeat < 3; repeat++ {
		for _, model := range []Model{base, permuted} {
			written, err := Put(document, model)
			if err != nil {
				t.Fatal(err)
			}
			if got := outputIDs(written); !reflect.DeepEqual(got, want) {
				t.Fatalf("permuted model changed output order: got %v want %v", got, want)
			}
			if got := outputSpans(written); !reflect.DeepEqual(got, []SourceSpan{{File: "ports.gooo", Start: 30, End: 35}, {File: "ports.gooo", Start: 40, End: 45}}) {
				t.Fatalf("permuted model lost output spans: %#v", got)
			}
		}
	}
	if _, err := Get(document); err != nil {
		t.Fatal(err)
	}
}
func TestInputPortOrderSurvivesModelPermutationsAndRepeats(t *testing.T) {
	document := sourceOrderedInputDocument()
	base, err := Get(document)
	if err != nil {
		t.Fatal(err)
	}
	permuted := base.Clone()
	reverseRelations(permuted.Relations)
	basePorts, _ := orderedSequences(base)
	permutedPorts, _ := orderedSequences(permuted)
	if !reflect.DeepEqual(basePorts, permutedPorts) || sequenceHash(basePorts) != sequenceHash(permutedPorts) {
		t.Fatalf("input port order was not source-authoritative: %v != %v", basePorts, permutedPorts)
	}
	want := []ID{"billing://entity/zebra", "billing://entity/apple"}
	for repeat := 0; repeat < 3; repeat++ {
		for _, model := range []Model{base, permuted} {
			written, err := Put(document, model)
			if err != nil {
				t.Fatal(err)
			}
			if got := inputIDs(written); !reflect.DeepEqual(got, want) {
				t.Fatalf("permuted model changed input order: got %v want %v", got, want)
			}
			if got := inputSpans(written); !reflect.DeepEqual(got, []SourceSpan{{File: "ports.gooo", Start: 10, End: 15}, {File: "ports.gooo", Start: 20, End: 25}}) {
				t.Fatalf("permuted model lost input spans: %#v", got)
			}
		}
	}
}
