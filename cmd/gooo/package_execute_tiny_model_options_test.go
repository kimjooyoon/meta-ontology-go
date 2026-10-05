package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

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
	var receipt tinyModelTestReceipt
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
