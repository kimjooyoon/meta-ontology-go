package bodyexecution

import (
	"context"
	"encoding/json"
	"os"
	"testing"
)

func TestCompositionTextOperationsNativeAssemblyAndReplay(t *testing.T) {
	source, err := os.ReadFile("../../examples/text-operations/source.gooo.fixture")
	if err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile("../../examples/text-operations/cases.json")
	if err != nil {
		t.Fatal(err)
	}
	suite, err := DecodeCompositionCases(raw)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	prior, err := GenerateCompositionWithOptions(ctx, "filenames.gooo", source, suite, CompositionOptions{EntryActivity: "Classify"})
	if err != nil {
		t.Fatal(err)
	}
	r := prior.Steps[0].Generation.Report.RecordAssembly
	if r == nil || r.Passed != 5 || r.Total != 5 || r.FieldsPassed != 15 || r.FieldsTotal != 15 ||
		r.SelectedMask != 7 || len(r.Attempts) != 8 || r.ModelCalls != 0 {
		t.Fatal("source filename choices did not complete", r)
	}
	native, err := ExecuteComposition(ctx, "filenames.gooo", source, prior, suite, nativeTool())
	if err != nil || native.FinitePassed != 12 || native.FiniteTotal != 12 || !native.RuntimeReplayed {
		t.Fatal("native filename classification", err, native)
	}
	encoded, err := json.Marshal(prior)
	if err != nil {
		t.Fatal(err)
	}
	restored, err := DecodeComposition(encoded)
	if err != nil {
		t.Fatal(err)
	}
	if err := VerifyComposition(ctx, "filenames.gooo", source, restored); err != nil {
		t.Fatal("saved text composition differs", err)
	}
	replay, err := ExecuteComposition(ctx, "filenames.gooo", source, restored, suite, nativeTool())
	if err != nil || replay.FinitePassed != native.FinitePassed || !replay.RuntimeReplayed {
		t.Fatal("saved native text execution differs", err, replay)
	}
}
