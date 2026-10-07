package workspaceexecution

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime"
)

func savedWorkspaceResult(t *testing.T, result Result) Result {
	t.Helper()
	raw, err := json.Marshal(result)
	if err != nil {
		t.Fatal(err)
	}
	var saved Result
	if err := bodyexecution.DecodeExecutionReceipt(raw, &saved); err != nil {
		t.Fatal(err)
	}
	return saved
}

func TestReplayWorkspaceSourceFillUsesNewInputsAndRetainsPartialResults(t *testing.T) {
	source, err := os.ReadFile("../../../examples/body-codegen/source-ir-fill.gooo.fixture")
	if err != nil {
		t.Fatal(err)
	}
	manifest := packageruntime.Manifest{Schema: packageruntime.ManifestSchema,
		Entry:    packageruntime.EntrySpec{PackagePath: "example/fill", Activity: "Lift"},
		Packages: []packageruntime.PackageSpec{{Path: "example/fill", Name: "source_ir_fill", Sources: []packageruntime.Source{{Filename: "fill.gooo", Content: string(source)}}}}}
	suite, err := bodyexecution.DecodeCompositionCases([]byte(`{"schema":"gooo/body-composition-cases/v1","cases":[{"inputs":{"example/fill:Lift":9},"expected":{"example/fill:Lift":10}}]}`))
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	prior, err := ExecuteWorkspace(ctx, manifest, suite, "", "")
	if err != nil {
		t.Fatal(err)
	}
	prior = savedWorkspaceResult(t, prior)
	suite.Cases[0].Inputs["example/fill:Lift"] = json.RawMessage(`9007199254740993`)
	suite.Cases[0].Expected["example/fill:Lift"] = json.RawMessage(`9007199254740994`)
	again, err := ReplayWorkspace(ctx, manifest, prior, suite, "")
	if err != nil || again.Runtime.FinitePassed != 1 || again.Replay == nil || again.Replay.BodyFillsReplayed != 1 || again.Replay.ModelCalls != 0 {
		t.Fatal("new-input replay failed", err, again.Replay)
	}
	if again.Runtime.RuntimeSuiteSHA256 == prior.Runtime.RuntimeSuiteSHA256 || !again.Runtime.RuntimeReplayed || again.Composition.GeneratedSHA256 != prior.Composition.GeneratedSHA256 {
		t.Fatal("replay did not use the saved program on the new suite")
	}
	suite.Cases[0].Expected["example/fill:Lift"] = json.RawMessage(`1`)
	again, err = ReplayWorkspace(ctx, manifest, savedWorkspaceResult(t, again), suite, "")
	if err != nil || again.Runtime.FinitePassed != 0 || again.Runtime.FiniteTotal != 1 || again.Replay.ModelCalls != 0 {
		t.Fatal("failed finite expectation did not remain observable", err)
	}
	for _, field := range []string{"source", "fill", "missing", "duplicate", "projection"} {
		bad := savedWorkspaceResult(t, prior)
		changed := manifest
		switch field {
		case "source":
			changed.Packages = append([]packageruntime.PackageSpec(nil), manifest.Packages...)
			changed.Packages[0].Sources = append([]packageruntime.Source(nil), manifest.Packages[0].Sources...)
			changed.Packages[0].Sources[0].Content += "\n"
		case "fill":
			bad.BodyFills[0].Generation.Report.BodyFill.SelectedCaseResults[0].Actual++
		case "missing":
			bad.BodyFills = nil
		case "duplicate":
			bad.BodyFills = append(bad.BodyFills, bad.BodyFills[0])
		case "projection":
			bad.Composition.Source += "\n"
		}
		if _, err := ReplayWorkspace(ctx, changed, bad, suite, "missing-go-binary"); err == nil || strings.Contains(err.Error(), "missing-go-binary") {
			t.Fatal("changed construction reached native execution", field, err)
		}
	}
}

func TestReplayWorkspaceImportsInputOnlyAndCancellation(t *testing.T) {
	manifest := importedActivityWorkspace()
	suite, err := bodyexecution.DecodeCompositionInputs([]byte(`{"schema":"gooo/body-composition-inputs/v1","inputs":[{"example/core:Normalize":"한글 and English"}]}`))
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	prior, err := ExecuteWorkspace(ctx, manifest, suite, "", "")
	if err != nil {
		t.Fatal(err)
	}
	prior = savedWorkspaceResult(t, prior)
	again, err := ReplayWorkspace(ctx, manifest, prior, suite, "")
	if err != nil || again.Runtime.FiniteTotal != 0 || again.Runtime.InputSeparation.Status != "UNKNOWN" || again.Replay.ModelCalls != 0 {
		t.Fatal("input-only replay invented correctness", err)
	}
	manifest.Packages[1].Sources[0].Content += "\n"
	if _, err := ReplayWorkspace(ctx, manifest, prior, suite, "missing-go-binary"); err == nil {
		t.Fatal("changed imported source replayed")
	}
	canceled, cancel := context.WithCancel(ctx)
	cancel()
	if _, err := ReplayWorkspace(canceled, manifest, prior, suite, ""); err != context.Canceled {
		t.Fatal("caller cancellation was lost", err)
	}
}
