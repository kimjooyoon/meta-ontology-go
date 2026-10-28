package bodyexecution

import (
	"context"
	"encoding/json"
	"os"
	"testing"
)

func TestConstructedRecordNativeCompositionAndActualPositions(t *testing.T) {
	source, err := os.ReadFile("../../examples/constructed-record-values/main.gooo")
	if err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile("../../examples/constructed-record-values/cases.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases CompositionCases
	if err := json.Unmarshal(raw, &cases); err != nil || len(cases.Cases) != 4 {
		t.Fatal("independent authored case set changed", err)
	}
	prior, err := GenerateCompositionWithOptions(context.Background(), "constructed.gooo", source, cases,
		CompositionOptions{EntryActivity: "Read"})
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(prior)
	if err != nil {
		t.Fatal(err)
	}
	var saved Composition
	if err := json.Unmarshal(encoded, &saved); err != nil {
		t.Fatal(err)
	}
	run, err := ExecuteComposition(context.Background(), "constructed.gooo", source, saved, cases, nativeTool())
	if err != nil || run.FinitePassed != 4 || run.FiniteTotal != 4 || !run.RuntimeReplayed {
		t.Fatalf("constructed record composition/replay: %v stage=%s finite=%d/%d build=%s", err, run.Stage, run.FinitePassed, run.FiniteTotal, run.Build.Diagnostics)
	}
	if len(run.ConstructionSites) != 1 || len(run.ReadSites) != 1 || run.ConstructionSites[0].ElementType != run.ReadSites[0].ElementType {
		t.Fatal("constructor and reader disagree on declared element identity", run.ConstructionSites, run.ReadSites)
	}
	for i, want := range []int{0, 1, 1, 0} {
		trace := run.Traces[i]
		if len(trace.Constructions) != 1 || !trace.Constructions[0].Completed || len(trace.Constructions[0].Elements) != 2 || len(trace.Reads) != 1 {
			t.Fatal("record constructor lost actual member positions", i, trace.Constructions, trace.Reads)
		}
		read := trace.Reads[0]
		if read.Origin != nil || read.Construction == nil || read.Construction.Instance != 0 || read.Construction.Index != want {
			t.Fatal("equal-valued records collapsed into another origin", i, read)
		}
	}
	if run.ReadObservation == nil || run.ReadObservation.ConstructedReads != 4 {
		t.Fatal("constructed record reads absent", run.ReadObservation)
	}
}
