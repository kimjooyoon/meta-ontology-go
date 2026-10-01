package main

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

func TestTypedPathDiagnosisCLIAndStrictRequest(t *testing.T) {
	read := func(name string) []byte {
		raw, err := os.ReadFile("../../examples/body-codegen/" + name)
		if err != nil {
			t.Fatal(err)
		}
		return raw
	}
	reader := mapSourceReader{"fixture.gooo": read("path-diagnosis.gooo.fixture"),
		"plan.json": read("path-diagnosis-plan.json"), "diagnosis.json": read("path-diagnosis-inputs.json")}
	args := []string{"--json", "--path-plan", "plan.json", "--path-diagnosis", "diagnosis.json",
		"--activity", "Probe", "fixture.gooo"}
	var stdout, stderr bytes.Buffer
	code := runBodyCodegen(args, reader, &stdout, &stderr)
	var result bodycodegen.Result
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if code != exitOK || stderr.Len() != 0 || result.Report.BodyPaths == nil || result.Report.BodyPaths.Diagnosis == nil ||
		result.Report.BodyPaths.Diagnosis.ProbeDistinguished != 1 || result.Report.BodyPaths.Search.Selection.ModelCalls != 0 {
		t.Fatal("CLI diagnosis failed", code, stderr.String())
	}
	for _, raw := range []string{
		`{"schema":"gooo/path-diagnosis-request/v1","inputs":[3],"max_candidates":2,"max_candidates":1}`,
		`{"schema":"gooo/path-diagnosis-request/v1","inputs":[3],"max_candidates":2,"expected":1}`,
		`{"schema":"gooo/path-diagnosis-request/v1","inputs":[3.5],"max_candidates":2}`,
		`{"schema":"gooo/path-diagnosis-request/v1","inputs":[],"max_candidates":2}`,
		`{"schema":"gooo/path-diagnosis-request/v1","inputs":[3],"max_candidates":65}`,
	} {
		if _, err := decodePathDiagnosis([]byte(raw)); err == nil {
			t.Fatal("invalid diagnosis JSON accepted", raw)
		}
	}
	for _, flags := range [][]string{{"--path-diagnosis", "diagnosis.json"},
		{"--path-plan", "plan.json", "--path-diagnosis", "diagnosis.json", "--path-diagnosis", "again.json"}} {
		stdout.Reset()
		stderr.Reset()
		if code := runBodyCodegen(append(flags, "--activity", "Probe", "fixture.gooo"), reader, &stdout, &stderr); code != exitUsage {
			t.Fatal("invalid diagnosis CLI mode accepted")
		}
	}
}
