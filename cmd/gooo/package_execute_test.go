package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/decisionroute"
)

func TestRunPackageExecuteRunsImportedActivityBodies(t *testing.T) {
	root := t.TempDir()
	writeWorkspaceSource(t, root, "core.gooo.fixture", `package core
namespace core
entity Text id "workspace://core/text"
activity Normalize(Text) -> Text computes "return input"
`)
	writeWorkspaceSource(t, root, "app.gooo.fixture", `package app
namespace app
import core "example/core"
activity Main(Text) -> Text computes "return input"
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
    "inputs": {"example/core:Normalize": "hello"},
    "expected": {"example/core:Normalize": "hello", "example/app:Main": "hello"}
  }]
}`), 0600); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	code := run([]string{"package", "execute", "--json", "--cases", casesPath, manifestPath}, &stdout, &stderr)
	if code != exitOK || stderr.Len() != 0 {
		t.Fatalf("workspace body execution failed: code=%d stderr=%q stdout=%q", code, stderr.String(), stdout.String())
	}
	var receipt packageExecutionReceipt
	if err := json.Unmarshal(stdout.Bytes(), &receipt); err != nil {
		t.Fatalf("decode package execution receipt: %v", err)
	}
	if receipt.Decision != "PASS" || receipt.Result == nil || receipt.Result.Runtime.FinitePassed != 2 ||
		receipt.Result.Runtime.FiniteTotal != 2 || !receipt.Result.Runtime.RuntimeReplayed ||
		receipt.Result.Program.Entry.Activity != "Main" {
		t.Fatalf("workspace execution receipt is incomplete: %#v", receipt)
	}
}

func TestRunPackageExecuteLoadsActivityBodyPlans(t *testing.T) {
	t.Setenv("GOOO_LAYA_URL", "")
	t.Setenv("GOOO_LAYA_API_KEY", "")
	root := filepath.Join("..", "..", "examples", "package-imports")
	var stdout, stderr bytes.Buffer
	code := run([]string{"package", "execute", "--json", "--cases", filepath.Join(root, "cases.json"),
		"--body-plans", filepath.Join(root, "body-plans.json"), filepath.Join(root, "gooo.workspace.json")}, &stdout, &stderr)
	if code != exitOK || stderr.Len() != 0 {
		t.Fatalf("workspace body plans failed: code=%d stderr=%q stdout=%q", code, stderr.String(), stdout.String())
	}
	var receipt packageExecutionReceipt
	if err := json.Unmarshal(stdout.Bytes(), &receipt); err != nil {
		t.Fatalf("decode package execution receipt: %v", err)
	}
	if receipt.Decision != "PASS" || receipt.Result == nil || len(receipt.Result.BodyFills) != 2 ||
		receipt.Result.BodyFills[0].Generation.Report.BodyFill.SelectedCandidateID != "increment" ||
		receipt.Result.Runtime.FinitePassed != 2 || receipt.Result.Runtime.FiniteTotal != 2 {
		t.Fatalf("body-fill plans were not scored and applied before execution: %#v", receipt)
	}
}

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

func TestRunPackageExecuteRejectsTinyModelProviderConflict(t *testing.T) {
	t.Setenv("GOOO_LAYA_URL", "http://laya.example.invalid/v1/systemone")
	root := t.TempDir()
	writeWorkspaceSource(t, root, "main.gooo.fixture", `package app
namespace app
entity Integer id "app://integer"
activity Main(Integer) -> Integer computes "return __GOOO_BODY_HOLE_value__"
`)
	manifestPath := writeWorkspaceManifest(t, root, `{"schema":"gooo/package-workspace-manifest/v1","entry":{"package_path":"app","activity":"Main"},"packages":[{"path":"app","name":"app","imports":[],"sources":["main.gooo.fixture"]}]}`)
	casesPath := filepath.Join(root, "cases.json")
	if err := os.WriteFile(casesPath, []byte(`{"schema":"gooo/body-composition-cases/v1","cases":[{"inputs":{"app:Main":0},"expected":{"app:Main":1}}]}`), 0600); err != nil {
		t.Fatal(err)
	}
	plansPath := filepath.Join(root, "plans.json")
	if err := os.WriteFile(plansPath, []byte(`{"schema":"gooo/workspace-body-fill-plans/v1","activities":[{"package_path":"app","activity":"Main","plan":{"schema":"gooo/body-codegen-ir-fill-plan/v1","intent":"add","hole_id":"value","candidates":[{"id":"add","expression":"input + 1"},{"id":"multiply","expression":"input * 1"}],"test_cases":[{"input":0,"expected":1}]}}]}`), 0600); err != nil {
		t.Fatal(err)
	}
	modelPath := writeSyntheticTinyGoModel(t, "add")
	var stdout, stderr bytes.Buffer
	code := run([]string{"package", "execute", "--json", "--cases", casesPath, "--body-plans", plansPath,
		"--tiny-model", modelPath, manifestPath}, &stdout, &stderr)
	if code != exitFailure || !strings.Contains(stdout.String(), "FAIL_CLOSED") ||
		strings.Contains(stdout.String()+stderr.String(), modelPath) || strings.Contains(stdout.String()+stderr.String(), "laya.example.invalid") {
		t.Fatalf("tiny-model provider conflict was not a sanitized closed failure: code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
	}
}

func TestLibraryStarterRunsWithTheLocalTinyModel(t *testing.T) {
	t.Setenv("GOOO_LAYA_URL", "")
	t.Setenv("GOOO_LAYA_API_KEY", "")
	root := t.TempDir()
	workspace := filepath.Join(root, "boundedint")
	var initOutput, initError bytes.Buffer
	if code := runInit([]string{"--template", "library", workspace}, &initOutput, &initError); code != exitOK {
		t.Fatalf("library starter init failed: code=%d stdout=%q stderr=%q", code, initOutput.String(), initError.String())
	}
	modelPath := writeSyntheticTinyGoModel(t, "add")
	var stdout, stderr bytes.Buffer
	code := run([]string{"package", "execute", "--json", "--cases", filepath.Join(workspace, "cases.json"),
		"--body-plans", filepath.Join(workspace, "body-fill-plans.json"), "--tiny-model", modelPath,
		filepath.Join(workspace, "gooo.workspace.json")}, &stdout, &stderr)
	if code != exitOK || stderr.Len() != 0 {
		t.Fatalf("library starter local model execution failed: code=%d stderr=%q stdout=%q", code, stderr.String(), stdout.String())
	}
	var receipt packageExecutionReceipt
	if err := json.Unmarshal(stdout.Bytes(), &receipt); err != nil {
		t.Fatalf("decode library starter execution receipt: %v", err)
	}
	if receipt.Decision != "PASS" || receipt.Result == nil || len(receipt.Result.BodyFills) != 2 ||
		receipt.Result.BodyFills[0].Generation.Report.BodyFill.SelectedCandidateID != "increment" ||
		receipt.Result.BodyFills[1].Generation.Report.BodyFill.SelectedCandidateID != "add_zero" ||
		receipt.Result.Runtime.FinitePassed != 2 || receipt.Result.Runtime.FiniteTotal != 2 || !receipt.Result.Runtime.RuntimeReplayed {
		t.Fatalf("library starter did not prove both model-filled package activities: %#v", receipt)
	}
}
