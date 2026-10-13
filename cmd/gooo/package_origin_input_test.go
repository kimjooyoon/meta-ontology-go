package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime/workspaceexecution"
)

func recordedOrigin9721(t *testing.T) bodyexecution.CompositionRuntime {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(originReaderRoot(t), "construction9533-runtime.json"))
	if err != nil {
		t.Fatal(err)
	}
	var saved bodyexecution.CompositionRuntime
	if err := bodyexecution.DecodeExecutionReceipt(raw, &saved); err != nil {
		t.Fatal(err)
	}
	return saved
}

func TestOriginEnvelope9721(t *testing.T) {
	r := recordedOrigin9721(t)
	envelopes := []any{r, bodyCompositionOutput{Runtime: r}, bodyConstructOutput{Evaluation: bodyexecution.JointEvaluation{Runtime: r}}, bodyexecution.JointEvaluation{Runtime: r}, packageExecutionReceipt{Schema: "gooo/workspace-body-execution-receipt/v1", Result: &workspaceexecution.Result{Runtime: r}}}
	for i, envelope := range envelopes {
		raw, err := json.Marshal(envelope)
		if err != nil {
			t.Fatal(err)
		}
		decoded, err := decodePackageOriginRuntime(raw)
		if err != nil || decoded.SelectedSourceSHA256 != r.SelectedSourceSHA256 {
			t.Fatal(i, err)
		}
		rows, origin, err := bodyexecution.SelectRecordedNativeOrigins(decoded, bodyexecution.NativeOriginQuery{Kind: "elements", Offset: 23, Limit: 128})
		if err != nil || len(rows) != 1 || origin.TotalRows != 24 || origin.Truncated {
			t.Fatal(i, origin, err)
		}
	}
	for _, raw := range []string{`{"runtime":{},"evaluation":{}}`, `{"runtime":{},"unknown":true}`, `{"schema":"a","schema":"b"}`, `{} {}`} {
		if _, err := decodePackageOriginRuntime([]byte(raw)); err == nil {
			t.Fatal("invalid envelope accepted", raw)
		}
	}
}

func TestOriginQuery9721(t *testing.T) {
	for _, args := range [][]string{
		{"--inputs", "i.json", "--origin-kind", "reads", "w.json"},
		{"--origin-receipt", "r.json", "--cases", "i.json", "w.json"},
		{"--origin-receipt", "r.json", "--origin-kind", "any", "w.json"},
		{"--origin-receipt", "r.json", "--origin-limit", "0", "w.json"},
		{"--origin-receipt", "r.json", "--origin-limit", "129", "w.json"},
		{"--origin-receipt", "r.json", "--origin-offset", "65537", "w.json"},
		{"--origin-receipt", "r.json", "--origin-offset", "x", "w.json"},
		{"--origin-receipt", "r.json", "--origin-limit", "2", "--origin-limit", "3", "w.json"},
	} {
		if _, _, err := parsePackageExecuteArgs(args); err == nil {
			t.Fatal("invalid origin flags accepted", args)
		}
	}
	flags, _, err := parsePackageExecuteArgs([]string{"--origin-receipt", "r.json", "w.json"})
	if err != nil {
		t.Fatal(err)
	}
	q, err := packageOriginQuery(flags)
	if err != nil || q.Kind != "elements" || q.Offset != 0 || q.Limit != 32 {
		t.Fatal(q, err)
	}
	r := recordedOrigin9721(t)
	if _, _, err := bodyexecution.SelectRecordedNativeOrigins(r, bodyexecution.NativeOriginQuery{Kind: "elements", Offset: 24, Limit: 1}); err == nil {
		t.Fatal("empty page silently accepted")
	}
	raw, _ := json.Marshal(r)
	encoded, origin, err := packageOriginInputs(raw, packageruntime.EntrySpec{PackagePath: "example.org/trace", Activity: "Explain"}, q)
	if err != nil {
		t.Fatal(err)
	}
	suite, err := bodyexecution.DecodeCompositionInputs(encoded)
	if err != nil || len(suite.Cases) != 24 || suite.Cases[0].Inputs["example.org/trace:Explain"] == nil {
		t.Fatal("input adapter", err)
	}
	var output bytes.Buffer
	writePackageOriginInput(&output, origin)
	if !strings.Contains(output.String(), "returned=24 total=24") || !strings.Contains(output.String(), "original observation: COMPLETE") {
		t.Fatal(output.String())
	}
	r.ConstructionSites[0].Elements[0].Source.Expression = strings.Repeat("x", 33000)
	raw, _ = json.Marshal(r)
	if _, origin, err := packageOriginInputs(raw, packageruntime.EntrySpec{PackagePath: "x", Activity: "Explain"}, q); err == nil || origin == nil || !strings.Contains(err.Error(), "lower --origin-limit") {
		t.Fatal("oversized page lost limit guidance", err)
	}
}

func TestOriginReferences9721(t *testing.T) {
	changes := []func(*bodyexecution.CompositionRuntime){
		func(r *bodyexecution.CompositionRuntime) { r.Traces[0].Reads[0].SiteIndex = -1 },
		func(r *bodyexecution.CompositionRuntime) { r.Traces[0].Constructions[0].SiteIndex = 999 },
		func(r *bodyexecution.CompositionRuntime) { r.Traces[0].Constructions[0].ReadEnd = 999 },
		func(r *bodyexecution.CompositionRuntime) { r.Traces[0].Constructions[0].Elements[0].ReadStart++ },
		func(r *bodyexecution.CompositionRuntime) { r.Traces[0].Reads[2].Construction.Instance = 999 },
		func(r *bodyexecution.CompositionRuntime) {
			r.Traces[0].Reads[2].Origin = &bodyexecution.CompositionReadOrigin{}
		},
		func(r *bodyexecution.CompositionRuntime) { r.Traces[0].ReadsObserved = false },
		func(r *bodyexecution.CompositionRuntime) { r.Traces[1].CaseIndex = 0 },
		func(r *bodyexecution.CompositionRuntime) { r.ReadObservation.RecordedReads++ },
		func(r *bodyexecution.CompositionRuntime) { r.ReadObservation.Status = "PARTIAL" },
		func(r *bodyexecution.CompositionRuntime) { r.Traces[0].ReadsDropped = -1 },
		func(r *bodyexecution.CompositionRuntime) { r.ReadObservation.Schema = "gooo/native-read-summary/v1" },
	}
	for i, change := range changes {
		r := recordedOrigin9721(t)
		change(&r)
		if _, _, err := bodyexecution.SelectRecordedNativeOrigins(r, bodyexecution.NativeOriginQuery{Kind: "elements", Limit: 1}); err == nil {
			t.Fatal("invalid recorded reference accepted", i)
		}
	}
}
