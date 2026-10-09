package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
)

func TestPackageConstructionReplayRetainsCurrentFailure(t *testing.T) {
	root := "../../examples/package-caller-construction/"
	original, raw := runPackageConstructionFixture(t, "--construction-cases", root+"construction-cases.json",
		"--inputs", root+"inputs.json", "--attempts", "6", root+"gooo.workspace.json")
	saved := filepath.Join(t.TempDir(), "saved.json")
	if err := os.WriteFile(saved, raw, 0600); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	code := run([]string{"package", "construct", "--json", "--receipt", saved,
		"--cases", root + "evaluation-cases.json", "--go", filepath.Join(t.TempDir(), "missing-go"),
		root + "gooo.workspace.json"}, &stdout, &stderr)
	var got packageConstructionReceipt
	if err := bodyexecution.DecodeExecutionReceipt(stdout.Bytes(), &got); err != nil || got.Result == nil {
		t.Fatal("missing failed receipt", err)
	}
	var observation struct {
		Result struct {
			Evaluation struct {
				ReplayFailure *struct {
					AttemptIndex int `json:"attempt_index"`
					Stage        string
					Runtime      struct{ Stage, Failure string }
				} `json:"replay_failure"`
			}
		}
	}
	if err := json.Unmarshal(stdout.Bytes(), &observation); err != nil {
		t.Fatal(err)
	}
	failure := observation.Result.Evaluation.ReplayFailure
	if code != exitFailure || got.Decision != "FAIL_CLOSED" || got.Error == "" ||
		failure == nil || failure.AttemptIndex != 0 || failure.Stage != "CALLER_EXECUTION" ||
		failure.Runtime.Stage == "" || !strings.Contains(failure.Runtime.Failure, "missing-go") {
		t.Fatalf("current failed attempt was discarded: code=%d failure=%+v error=%s", code, failure, got.Error)
	}
	if got.Result.Evaluation.ConstructionReplayed || got.Result.Evaluation.Runtime.Stage != "" ||
		got.Result.Evaluation.NewModelCalls != 0 {
		t.Fatal("failed historical replay claimed a new evaluation")
	}
	before, _ := json.Marshal(original.Result.Construction)
	after, _ := json.Marshal(got.Result.Construction)
	if !bytes.Equal(before, after) {
		t.Fatal("current failure changed the saved construction history")
	}
	if err := os.WriteFile(saved, stdout.Bytes(), 0600); err != nil {
		t.Fatal(err)
	}
	stdout.Reset()
	stderr.Reset()
	code = run([]string{"package", "construct", "--json", "--receipt", saved,
		"--inputs", root + "inputs.json", root + "gooo.workspace.json"}, &stdout, &stderr)
	if code != exitFailure || !strings.Contains(stdout.String(), "saved package construction envelope") {
		t.Fatal("failed history was accepted for another replay", code)
	}
}
