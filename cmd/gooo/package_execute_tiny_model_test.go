package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/decisionroute"
)

func TestRunPackageExecuteFillsImportedActivitiesWithOneTinyModel(t *testing.T) {
	t.Setenv("GOOO_LAYA_URL", "")
	t.Setenv("GOOO_LAYA_API_KEY", "")
	root := t.TempDir()
	writeWorkspaceSource(t, root, "core.gooo.fixture", `package core
namespace core
entity Integer id "workspace://integer"
activity Normalize(Integer) -> Integer computes "return __GOOO_BODY_HOLE_value__"
`)
	writeWorkspaceSource(t, root, "app.gooo.fixture", `package app
namespace app
import core "example/core"
activity Main(Integer) -> Integer computes "return __GOOO_BODY_HOLE_value__"
bind core.Normalize.result -> Main.input
`)
	manifestPath := writeWorkspaceManifest(t, root, `{
  "schema": "gooo/package-workspace-manifest/v1",
  "entry": {"package_path": "example/app", "activity": "Main"},
  "packages": [
    {"path": "example/app", "name": "app", "imports": ["example/core"], "sources": ["app.gooo.fixture"]},
    {"path": "example/core", "name": "core", "imports": [], "sources": ["core.gooo.fixture"]}
  ]
}`)
	casesPath := filepath.Join(root, "cases.json")
	if err := os.WriteFile(casesPath, []byte(`{
  "schema": "gooo/body-composition-cases/v1",
  "cases": [{
    "inputs": {"example/core:Normalize": 7},
    "expected": {"example/core:Normalize": 8, "example/app:Main": 8}
  }]
}`), 0600); err != nil {
		t.Fatal(err)
	}
	plansPath := filepath.Join(root, "body-plans.json")
	if err := os.WriteFile(plansPath, []byte(`{
  "schema": "gooo/workspace-body-fill-plans/v1",
  "activities": [
    {"package_path":"example/core","activity":"Normalize","plan":{"schema":"gooo/body-codegen-ir-fill-plan/v1","intent":"Add one.","hole_id":"value","candidates":[{"id":"increment","expression":"input + 1"},{"id":"multiply_one","expression":"input * 1"}],"test_cases":[{"input":-4,"expected":-3},{"input":0,"expected":1},{"input":7,"expected":8}]}},
    {"package_path":"example/app","activity":"Main","plan":{"schema":"gooo/body-codegen-ir-fill-plan/v1","intent":"Preserve the delivered value.","hole_id":"value","candidates":[{"id":"add_zero","expression":"input + 0"},{"id":"multiply_zero","expression":"input * 0"}],"test_cases":[{"input":-3,"expected":-3},{"input":8,"expected":8}]}}
  ]
}`), 0600); err != nil {
		t.Fatal(err)
	}
	modelPath := writeSyntheticTinyGoModel(t, "add")
	var stdout, stderr bytes.Buffer
	code := run([]string{"package", "execute", "--json", "--cases", casesPath, "--body-plans", plansPath,
		"--tiny-model", modelPath, manifestPath}, &stdout, &stderr)
	if code != exitOK || stderr.Len() != 0 {
		t.Fatalf("tiny-model workspace execution failed: code=%d stderr=%q stdout=%q", code, stderr.String(), stdout.String())
	}
	var receipt packageExecutionReceipt
	if err := json.Unmarshal(stdout.Bytes(), &receipt); err != nil {
		t.Fatalf("decode tiny-model workspace receipt: %v", err)
	}
	if receipt.Decision != "PASS" || receipt.Result == nil || len(receipt.Result.BodyFills) != 2 ||
		receipt.Result.Runtime.FinitePassed != 2 || receipt.Result.Runtime.FiniteTotal != 2 || !receipt.Result.Runtime.RuntimeReplayed {
		t.Fatalf("tiny-model workspace receipt is incomplete: %#v", receipt)
	}
	for index, want := range []string{"increment", "add_zero"} {
		fill := receipt.Result.BodyFills[index].Generation.Report.BodyFill
		if fill == nil || fill.SelectedCandidateID != want || fill.Decision.Provider != decisionroute.ProviderTinyGo ||
			fill.LocalModelPredictions != 1 || fill.ExternalProviderCalls != 0 {
			t.Fatalf("activity %d did not use the one loaded local model: %#v", index, fill)
		}
		loadTime := receipt.Result.BodyFills[index].Generation.Report.BodyFill.Timing.TinyModelLoadMS
		if (index == 0 && loadTime == nil) || (index > 0 && loadTime != nil) {
			t.Fatalf("model-load timing must be attributed once: activity=%d timing=%#v", index, loadTime)
		}
	}
}
