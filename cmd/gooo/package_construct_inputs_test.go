package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime/workspaceexecution"
)

func TestPackageConstructInputsAndSavedObservation(t *testing.T) {
	root := "../../examples/package-caller-construction/"
	r, raw := runPackageConstructionFixture(t, "--construction-cases", root+"construction-cases.json",
		"--inputs", root+"inputs.json", "--attempts", "6", root+"gooo.workspace.json")
	input, err := os.ReadFile(root + "inputs.json")
	if err != nil {
		t.Fatal(err)
	}
	if r.Decision != "OBSERVED" || r.InputsDigest != workspaceDigest(input) || r.CasesDigest != "" ||
		r.Result.Construction.Decision != "COMPLETE_FINITE" || r.Result.Evaluation.ConstructionReplayed ||
		r.Result.Evaluation.Runtime.FiniteTotal != 0 || r.Result.Evaluation.Runtime.FinitePassed != 0 {
		t.Fatal("actual inputs were scored or not bound", r)
	}
	var values, diagnostics bytes.Buffer
	if code := writePackageConstruction(r, nil, false, &values, &diagnostics); code != exitOK ||
		values.String() != "3\n-10\n-9007199254740994\n1\n" || diagnostics.Len() != 0 {
		t.Fatal(code, values.String(), diagnostics.String())
	}
	for _, trace := range r.Result.Evaluation.Runtime.Traces {
		for _, d := range trace.Deliveries {
			if len(d.Expected) != 0 || d.Passed != nil {
				t.Fatal("input-only execution fabricated an oracle", d)
			}
		}
	}
	saved := filepath.Join(t.TempDir(), "observation.json")
	if err := os.WriteFile(saved, raw, 0600); err != nil {
		t.Fatal(err)
	}
	replay, next := runPackageConstructionFixture(t, "--receipt", saved, "--inputs", root+"inputs.json", root+"gooo.workspace.json")
	originalHistory, _ := json.Marshal(r.Result.Construction)
	replayedHistory, _ := json.Marshal(replay.Result.Construction)
	if replay.Decision != "OBSERVED" || replay.ReplayedFrom != workspaceDigest(raw) ||
		!replay.Result.Evaluation.ConstructionReplayed || replay.Result.Evaluation.NewModelCalls != 0 ||
		!bytes.Equal(originalHistory, replayedHistory) {
		t.Fatal("observation replay changed construction", replay.Decision)
	}
	if err := os.WriteFile(saved, next, 0600); err != nil {
		t.Fatal(err)
	}
	scored, _ := runPackageConstructionFixture(t, "--receipt", saved, "--cases", root+"evaluation-cases.json", root+"gooo.workspace.json")
	if scored.Decision != "COMPLETE_FINITE" || scored.InputsDigest != "" || scored.CasesDigest == "" ||
		scored.Result.Evaluation.Runtime.FinitePassed != 4 || scored.Result.Evaluation.Runtime.FiniteTotal != 4 ||
		scored.Result.Evaluation.NewModelCalls != 0 || scored.ReplayedFrom != workspaceDigest(next) {
		t.Fatal("labelled evaluation could not follow an unscored observation", scored.Decision)
	}
	// Observation receipts must still reconstruct the original attempts.
	replay.Result.Construction.Attempts[0].LocalPassed++
	tampered, err := json.Marshal(replay)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(saved, tampered, 0600); err != nil {
		t.Fatal(err)
	}
	values.Reset()
	diagnostics.Reset()
	code := run([]string{"package", "construct", "--json", "--receipt", saved,
		"--inputs", root + "inputs.json", root + "gooo.workspace.json"}, &values, &diagnostics)
	if code != exitFailure || !strings.Contains(values.String(), `"decision":"FAIL_CLOSED"`) {
		t.Fatal("input-only replay skipped history verification", code, diagnostics.String())
	}
}

func TestPackageConstructPartialInputsRemainUnscored(t *testing.T) {
	root := "../../examples/package-caller-construction/"
	r, _ := runPackageConstructionFixture(t, "--construction-cases", root+"construction-cases.json",
		"--inputs", root+"inputs.json", "--attempts", "5", root+"gooo.workspace.json")
	if r.Decision != "OBSERVED" || r.Result.Construction.Decision != "PARTIAL_FINITE" ||
		len(r.Result.Construction.Attempts) != 5 || r.Result.Evaluation.Runtime.FiniteTotal != 0 {
		t.Fatal("unscored inputs changed the construction decision", r.Decision)
	}
}

func TestPackageConstructInputOptionsAndSchema(t *testing.T) {
	for _, mode := range [][]string{
		{}, {"--inputs", "x", "--cases", "y"}, {"--inputs", "x", "--inputs", "y"}, {"--inputs", ""},
	} {
		args := append([]string{"--receipt", "saved", "workspace"}, mode...)
		if _, _, err := parsePackageConstructArgs(args); err == nil {
			t.Fatal("ambiguous or missing input mode accepted", args)
		}
	}
	root := "../../examples/package-caller-construction/"
	var out, diagnostics bytes.Buffer
	code := run([]string{"package", "construct", "--json", "--construction-cases", root + "construction-cases.json",
		"--inputs", root + "evaluation-cases.json", "--attempts", "6", root + "gooo.workspace.json"}, &out, &diagnostics)
	if code != exitFailure || !strings.Contains(out.String(), `"decision":"FAIL_CLOSED"`) {
		t.Fatal("scored cases accepted as actual-only inputs", code, out.String(), diagnostics.String())
	}
}

func TestPackageConstructPlainInputsRetainFaults(t *testing.T) {
	for _, delivery := range []bodyexecution.CompositionDelivery{
		{Fault: &bodyexecution.CompositionFault{Kind: "ZERO_DIVISOR"}},
		{BlockedBy: []string{"upstream"}},
	} {
		delivery.ActivityID = "entry"
		r := packageConstructionReceipt{Decision: "OBSERVED", InputsDigest: "input",
			Result: &workspaceexecution.ConstructionResult{
				Program: workspaceexecution.Program{Entry: workspaceexecution.ActivityRef{LoweredName: "Main"}},
				Construction: bodyexecution.JointConstruction{Selected: bodyexecution.Composition{
					Plan: bodyexecution.CompositionPlan{Activities: []bodyexecution.CompositionActivity{{ID: "entry", Name: "Main"}}}}},
				Evaluation: bodyexecution.JointEvaluation{Runtime: bodyexecution.CompositionRuntime{
					Traces: []bodyexecution.CompositionTrace{{Deliveries: []bodyexecution.CompositionDelivery{delivery}}}}},
			}}
		var out, diagnostics bytes.Buffer
		if code := writePackageConstruction(r, nil, false, &out, &diagnostics); code != exitFailure ||
			out.Len() != 0 || diagnostics.Len() == 0 {
			t.Fatal("native fault was printed as a successful value", code, out.String(), diagnostics.String())
		}
		out.Reset()
		diagnostics.Reset()
		if code := writePackageConstruction(r, nil, true, &out, &diagnostics); code != exitOK ||
			!strings.Contains(out.String(), `"decision":"OBSERVED"`) {
			t.Fatal("JSON observation did not retain a native outcome", code, out.String())
		}
	}
}

func TestPackageConstructActualNativeFault(t *testing.T) {
	root := "../../examples/package-caller-construction/"
	dir := t.TempDir()
	for _, name := range []string{"gooo.workspace.json", "app.gooo.fixture", "budget.gooo.fixture"} {
		raw, err := os.ReadFile(root + name)
		if err != nil {
			t.Fatal(err)
		}
		if name == "app.gooo.fixture" {
			raw = bytes.Replace(raw, []byte("if plan.exhausted { return -plan.next }\nreturn plan.next"),
				[]byte("return plan.next / (input.limit - input.used)"), 1)
		}
		if err := os.WriteFile(filepath.Join(dir, name), raw, 0600); err != nil {
			t.Fatal(err)
		}
	}
	files := map[string]string{
		"feedback.json": `{"schema":"gooo/body-composition-cases/v1","cases":[{"inputs":{"app/retry:Main":{"used":0,"limit":8}},"expected":{"app/retry:Main":0}}]}`,
		"inputs.json":   `{"schema":"gooo/body-composition-inputs/v1","inputs":[{"app/retry:Main":{"used":8,"limit":8}}]}`,
	}
	for name, value := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(value), 0600); err != nil {
			t.Fatal(err)
		}
	}
	r, raw := runPackageConstructionFixture(t, "--construction-cases", filepath.Join(dir, "feedback.json"),
		"--inputs", filepath.Join(dir, "inputs.json"), "--attempts", "1", filepath.Join(dir, "gooo.workspace.json"))
	if r.Decision != "OBSERVED" || r.Result.Evaluation.Runtime.FiniteTotal != 0 ||
		!r.Result.Evaluation.Runtime.RuntimeReplayed || !bytes.Contains(raw, []byte(`"kind":"ZERO_DIVISOR"`)) {
		t.Fatal("native zero divisor disappeared from unscored observations", r.Decision)
	}
	var out, diagnostics bytes.Buffer
	if code := writePackageConstruction(r, nil, false, &out, &diagnostics); code != exitFailure ||
		out.Len() != 0 || !strings.Contains(diagnostics.String(), "ZERO_DIVISOR") {
		t.Fatal("native fault became a successful plain value", code, out.String(), diagnostics.String())
	}
}
