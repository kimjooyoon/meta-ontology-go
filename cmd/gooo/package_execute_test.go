package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
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

func TestRunPackageExecuteTransportsOptionalV4RecordFieldsAcrossPackageBind(t *testing.T) {
	t.Setenv("GOOO_LAYA_URL", "")
	t.Setenv("GOOO_LAYA_API_KEY", "")
	root := filepath.Join("..", "..", "examples", "package-optional-record-flow")
	var stdout, stderr bytes.Buffer
	code := run([]string{"package", "execute", "--json", "--cases", filepath.Join(root, "cases.json"),
		filepath.Join(root, "gooo.workspace.json")}, &stdout, &stderr)
	if code != exitOK || stderr.Len() != 0 {
		t.Fatalf("optional V4 package flow failed: code=%d stderr=%q stdout=%q", code, stderr.String(), stdout.String())
	}
	var receipt packageExecutionReceipt
	if err := json.Unmarshal(stdout.Bytes(), &receipt); err != nil {
		t.Fatalf("decode optional package execution receipt: %v", err)
	}
	if receipt.Decision != "PASS" || receipt.Result == nil || receipt.Result.Runtime.FinitePassed != 6 ||
		receipt.Result.Runtime.FiniteTotal != 6 || !receipt.Result.Runtime.RuntimeReplayed ||
		len(receipt.Result.Runtime.Traces) != 3 {
		t.Fatalf("optional V4 package evidence is incomplete: %#v", receipt)
	}
	wantPresent := [][]bool{{false, false, false}, {true, true, true}, {true, false, true}}
	wantValues := [][]string{{"", "", ""}, {`""`, "false", "0"}, {`"queued"`, "", "9007199254740995"}}
	fieldNames := []string{"note", "complete", "count"}
	for caseIndex, trace := range receipt.Result.Runtime.Traces {
		if len(trace.Deliveries) != 2 || !bytes.Equal(trace.Deliveries[0].Actual, trace.Deliveries[1].Input) ||
			!bytes.Equal(trace.Deliveries[0].Actual, trace.Deliveries[1].Actual) {
			t.Fatalf("case %d did not preserve the record across its package bind: %+v", caseIndex, trace.Deliveries)
		}
		for deliveryIndex, delivery := range trace.Deliveries {
			if len(delivery.ActualFields) != len(fieldNames) {
				t.Fatalf("case %d delivery %d field observations=%+v", caseIndex, deliveryIndex, delivery.ActualFields)
			}
			for fieldIndex, field := range delivery.ActualFields {
				if field.Name != fieldNames[fieldIndex] || field.Presence != "optional" || field.Present == nil ||
					*field.Present != wantPresent[caseIndex][fieldIndex] {
					t.Fatalf("case %d delivery %d field %q lost V4 presence: %+v", caseIndex, deliveryIndex, field.Name, field)
				}
				if !wantPresent[caseIndex][fieldIndex] {
					if field.Value != nil && string(field.Value) != "null" {
						t.Fatalf("case %d absent field %q has value %s", caseIndex, field.Name, field.Value)
					}
					continue
				}
				if string(field.Value) != wantValues[caseIndex][fieldIndex] {
					t.Fatalf("case %d field %q value=%s, want %s", caseIndex, field.Name, field.Value, wantValues[caseIndex][fieldIndex])
				}
			}
		}
	}
}
