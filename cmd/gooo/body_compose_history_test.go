package main

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
)

func TestBodyCompositionCaseSeriesExecutesFreshRecordInputsWithOneBuild(t *testing.T) {
	root := t.TempDir()
	source := "../../examples/body-codegen/record-field-assembly.gooo.fixture"
	suite := bodyexecution.CompositionCases{Schema: "gooo/body-composition-cases/v1", Cases: []bodyexecution.CompositionCase{{
		Inputs:   map[string]json.RawMessage{"Select.input0": json.RawMessage(`{"title":"first","state":"queued","reason":"검토"}`), "Select.input1": json.RawMessage(`true`)},
		Expected: map[string]json.RawMessage{"Select": json.RawMessage(`{"title":"first","state":"ready","reason":"검토:accepted"}`), "Label": json.RawMessage(`"first:ready:검토:accepted"`)},
	}}}
	second := bodyexecution.CompositionCases{Schema: suite.Schema, Cases: []bodyexecution.CompositionCase{{
		Inputs:   map[string]json.RawMessage{"Select.input0": json.RawMessage(`{"title":"new","state":"queued","reason":"later"}`), "Select.input1": json.RawMessage(`false`)},
		Expected: map[string]json.RawMessage{"Select": json.RawMessage(`{"title":"new","state":"queued","reason":"later"}`), "Label": json.RawMessage(`"wrong expectation"`)},
	}}}
	series := bodyexecution.CompositionCaseSeries{Schema: "gooo/body-composition-case-series/v1", Suites: []bodyexecution.CompositionCases{suite, second}}
	raw, err := json.Marshal(series)
	if err != nil {
		t.Fatal(err)
	}
	seriesPath := filepath.Join(root, "series.json")
	if err := os.WriteFile(seriesPath, raw, 0600); err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(root, "out")
	var stdout, stderr bytes.Buffer
	args := []string{"--source", source, "--case-series", seriesPath, "--repeat", "2", "--out", out}
	if code := runBodyComposeContext(context.Background(), args, &stdout, &stderr); code != exitOK {
		t.Fatal(code, stderr.String())
	}
	var output bodyCompositionOutput
	if err := json.Unmarshal(stdout.Bytes(), &output); err != nil {
		t.Fatal(err)
	}
	if output.CaseSeries == nil || len(output.RuntimeHistory) != 4 || output.Runtime.FinitePassed != 1 ||
		output.Runtime.FiniteTotal != 2 || !output.Runtime.Artifact.Reused {
		t.Fatal(output.Runtime)
	}
	for i, observation := range output.RuntimeHistory {
		passed := 2
		if i%2 == 1 {
			passed = 1
		}
		if observation.FinitePassed != passed || observation.FiniteTotal != 2 || len(observation.Runs) != 2 ||
			observation.ModelCalls != 0 || observation.Artifact.Reused != (i > 0) || observation.Build.Started != (i == 0) {
			t.Fatal(i, observation)
		}
	}
	for _, name := range []string{"runtime-history.json", "case-series.json", "cases.json"} {
		if info, err := os.Stat(filepath.Join(out, name)); err != nil || info.Size() == 0 {
			t.Fatal(name, err)
		}
	}
	stdout.Reset()
	stderr.Reset()
	args = []string{"--source", source, "--case-series", seriesPath, "--composition", filepath.Join(out, "composition.json")}
	if code := runBodyComposeContext(context.Background(), args, &stdout, &stderr); code != exitOK {
		t.Fatal(code, stderr.String())
	}
	if err := json.Unmarshal(stdout.Bytes(), &output); err != nil {
		t.Fatal(err)
	}
	if output.GeneratedNow || len(output.RuntimeHistory) != 2 || !output.Runtime.Artifact.Reused || output.Runtime.ModelCalls != 0 {
		t.Fatal(output)
	}
}

func TestBodyCompositionSeriesAndRepeatLimitsBeforeGeneration(t *testing.T) {
	for _, extra := range [][]string{{"--repeat", "0"}, {"--repeat", "17"}, {"--repeat", "bad"}, {"--case-series", "also.json"}} {
		args := append([]string{"--source", "missing.gooo", "--cases", "missing.json"}, extra...)
		var stdout, stderr bytes.Buffer
		if code := runBodyComposeContext(context.Background(), args, &stdout, &stderr); code != exitUsage || stdout.Len() != 0 {
			t.Fatal(args, code, stderr.String())
		}
	}
	series := `{"schema":"gooo/body-composition-case-series/v1","suites":[{"schema":"gooo/body-composition-cases/v1","cases":[{"inputs":{"unknown":1},"expected":{"Label":"text"}}]}]}`
	name := filepath.Join(t.TempDir(), "bad-series.json")
	if err := os.WriteFile(name, []byte(series), 0600); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	args := []string{"--source", "../../examples/body-codegen/record-field-assembly.gooo.fixture", "--case-series", name, "--model", "missing-model.json"}
	if code := runBodyComposeContext(context.Background(), args, &stdout, &stderr); code != exitFailure || stdout.Len() != 0 ||
		!bytes.Contains(stderr.Bytes(), []byte("suite 0")) {
		t.Fatal(code, stdout.String(), stderr.String())
	}
}
