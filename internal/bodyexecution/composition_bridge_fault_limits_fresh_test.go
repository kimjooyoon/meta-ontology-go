package bodyexecution

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFieldLocalBridgeFault15617(t *testing.T) {
	const source = `package bridgefault15617
namespace bridgefault15617
entity Integer id "urn:gooo:type:integer"
entity Input id "urn:bridgefault15617:input" fields {
field value id "urn:bridgefault15617:value" type integer required one
field divisor id "urn:bridgefault15617:divisor" type integer required one
field local id "urn:bridgefault15617:local" type boolean required one
}
activity Build(Input)->Integer id "urn:bridgefault15617:build" computes ` + "`let current=input;let saved=current.value;if input.local {let failed=saved/input.divisor;current.value=failed}else{current.value=current.value/input.divisor};return current.value`" + `
`
	suite := CompositionCases{Schema: "gooo/body-composition-cases/v1", Cases: []CompositionCase{
		{Inputs: map[string]json.RawMessage{"Build": json.RawMessage(`{"value":77,"divisor":7,"local":false}`)}, Expected: map[string]json.RawMessage{"Build": json.RawMessage(`11`)}},
		{Inputs: map[string]json.RawMessage{"Build": json.RawMessage(`{"value":83,"divisor":0,"local":true}`)}, Expected: map[string]json.RawMessage{"Build": json.RawMessage(`0`)}},
		{Inputs: map[string]json.RawMessage{"Build": json.RawMessage(`{"value":89,"divisor":0,"local":false}`)}, Expected: map[string]json.RawMessage{"Build": json.RawMessage(`0`)}},
	}}
	program, err := GenerateCompositionWithOptions(context.Background(), "bridgefault15617.gooo", []byte(source), suite, CompositionOptions{EntryActivity: "Build"})
	if err != nil {
		t.Fatal(err)
	}
	r, err := ExecuteComposition(context.Background(), "bridgefault15617.gooo", []byte(source), program, suite, "")
	if err != nil {
		t.Fatal(err)
	}
	if r.FinitePassed != 1 || !r.RuntimeReplayed || r.ModelCalls != 0 || !hasCompositionFault(r) || r.BridgeObservation == nil || r.BridgeObservation.Status != "PARTIAL" || r.BridgeObservation.Interrupted != 2 || r.FieldObservation.Interrupted != 1 || r.LocalObservation.Interrupted != 1 {
		t.Fatal("one normal value and two distinct interrupted scopes", r.FinitePassed, r.BridgeObservation)
	}
	if !bridgeCaseRecordingComplete(r.Traces[0]) || bridgeCaseRecordingComplete(r.Traces[1]) || bridgeCaseRecordingComplete(r.Traces[2]) {
		t.Fatal("recording completeness leaked between distinct cases")
	}
	offset := 0
	for c, trace := range r.Traces {
		columns, report, err := SelectRecordedBridgeLineage(r, NativeOriginQuery{Kind: "field-uses", Offset: offset, Limit: 1})
		if err != nil || columns.RecordingComplete != (c == 0) || report.ProducerExecutions != 0 || report.ProducerModelCalls != 0 {
			t.Fatal("per-case recorded input scope", c, err, report)
		}
		offset += len(trace.FieldUses)
	}
	saveBridgeRuntime15617(t, "fault15617", source, suite, r)
	t.Log("fresh3 cases:one actual11 andtwo intended division faults;one scalar write andone field write remain unfinished;shared clock replay equal;per-case complete/partial input labels;models0")
}

func TestFieldLocalBridgeLimits15641(t *testing.T) {
	source := `package bridgelimits15641
namespace bridgelimits15641
entity State id "urn:bridgelimits15641:state" fields {
field value id "urn:bridgelimits15641:value" type integer required one
}
activity Build(State)->State id "urn:bridgelimits15641:build" computes ` + "`let current=input;let outer=0;let carried=0;for outer=0;outer<1024;outer=outer+1 {" + strings.Repeat("carried=current.value;current.value=carried+carried-carried-carried+carried+1;", 16) + "};return current`" + `
`
	suite := CompositionCases{Schema: "gooo/body-composition-cases/v1", Cases: []CompositionCase{
		{Inputs: map[string]json.RawMessage{"Build": json.RawMessage(`{"value":101}`)}, Expected: map[string]json.RawMessage{"Build": json.RawMessage(`{"value":16485}`)}},
		{Inputs: map[string]json.RawMessage{"Build": json.RawMessage(`{"value":211}`)}, Expected: map[string]json.RawMessage{"Build": json.RawMessage(`{"value":16595}`)}},
	}}
	program, err := GenerateCompositionWithOptions(context.Background(), "bridgelimits15641.gooo", []byte(source), suite, CompositionOptions{EntryActivity: "Build"})
	if err != nil {
		t.Fatal(err)
	}
	r, err := ExecuteComposition(context.Background(), "bridgelimits15641.gooo", []byte(source), program, suite, "")
	if err != nil {
		t.Fatal(err)
	}
	b := r.BridgeObservation
	if r.FinitePassed != 2 || !r.RuntimeReplayed || r.ModelCalls != 0 || b == nil || b.Status != "PARTIAL" || b.FieldWindows != 4096 || b.LocalWindows != 4096 || b.Completions != 8192 || b.LocalUses != 65536 || r.FieldObservation.WritesDropped < 1 || r.LocalObservation.WritesDropped < 1 || r.LocalObservation.UsesDropped < 1 {
		t.Fatal("bounded retained journals and uninterrupted actual arithmetic", r.FinitePassed, b, r.FieldObservation, r.LocalObservation)
	}
	if err := validateRecordedBridgeObservations(r); err != nil {
		t.Fatal(err)
	}
	saveBridgeRuntime15617(t, "limits15641", source, suite, r)
	t.Log("fresh2 cases and32768 field/local handoffs;actual values2/2;field/local4096+4096 writes and8192 shared completions retained,local65536 uses;loss stays partial;business computation continues;2-run equality;models0")
}

func saveBridgeRuntime15617(t *testing.T, name, source string, suite CompositionCases, r CompositionRuntime) {
	t.Helper()
	dir := os.Getenv("GOOO_FRESH_CLI_OUTPUT")
	if dir == "" {
		return
	}
	raw, err := json.Marshal(map[string]any{"source": source, "cases": suite, "runtime": r})
	if err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	z := gzip.NewWriter(&out)
	if _, err := z.Write(raw); err != nil {
		t.Fatal(err)
	}
	if err := z.Close(); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, name+".json.gz"), out.Bytes(), 0600); err != nil {
		t.Fatal(err)
	}
}
