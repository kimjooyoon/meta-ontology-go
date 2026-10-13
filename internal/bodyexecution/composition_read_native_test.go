package bodyexecution

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestNativeReadOrigins9059(t *testing.T) {
	const source = `package reading
namespace reading9059
entity Integer id "urn:gooo:type:integer"
entity Request id "urn:read9059:request" fields {
 field readings id "urn:read9059:request-values" type integer required many
 field labels id "urn:read9059:request-labels" type string required many
 field flags id "urn:read9059:request-flags" type boolean required many
 field lower id "urn:read9059:lower" type integer required one
 field upper id "urn:read9059:upper" type integer required one
 field pick id "urn:read9059:pick" type integer required one
}
entity Window id "urn:read9059:window" fields {
 field readings id "urn:read9059:window-values" type integer required many
 field labels id "urn:read9059:window-labels" type string required many
 field flags id "urn:read9059:window-flags" type boolean required many
 field pick id "urn:read9059:window-pick" type integer required one
}
activity Clip(Request) -> Window id "urn:read9059:clip" computes "let 시작 = input.lower; let 끝 = input.upper; return Window{readings: input.readings[시작:끝], labels: input.labels[시작:끝], flags: input.flags[시작:끝], pick: input.pick}"
activity At(Window) -> Integer id "urn:read9059:at" computes "if input.pick < 0 || input.pick >= int64(len(input.readings)) { return -9 }; let total = 0; for 값 in input.readings { total = total + 값 }; if input.flags[input.pick] && input.labels[input.pick] != \"skip\" { if input.labels[input.pick] == \"fresh\" { return []int64{37,37}[input.pick % 2] }; return total / (input.pick + 1) + input.readings[__GOOO_BODY_HOLE_position__] }; return -7" assembling {
 source_fill intent "원본 순서를 유지한 구간에서 요청한 위치와 합계를 사용한다" {
  hole "position"
  candidate "first9059" { fill "position" "0" }
  candidate "requested9059" { fill "position" "input.pick" }
 }
 value_case "[{\"readings\":[31,47],\"labels\":[\"yes\",\"yes\"],\"flags\":[true,true],\"pick\":1}]" -> "86"
 value_case "[{\"readings\":[61,71,89],\"labels\":[\"yes\",\"yes\",\"yes\"],\"flags\":[true,true,true],\"pick\":2}]" -> "162"
 holdout_value_case "[{\"readings\":[101,107],\"labels\":[\"yes\",\"yes\"],\"flags\":[true,true],\"pick\":1}]" -> "211"
}
activity Run(Window) -> Integer id "urn:read9059:run" computes "return At(input)"
bind Clip.result -> Run.input
`
	const cases = `{"schema":"gooo/body-composition-cases/v1","cases":[
{"inputs":{"Clip":{"readings":[17,37,37,43],"labels":["yes","yes","yes","yes"],"flags":[true,true,true,true],"lower":1,"upper":4,"pick":1}},"expected":{"Run":95}},
{"inputs":{"Clip":{"readings":[181,191],"labels":["yes","yes"],"flags":[true,true],"lower":0,"upper":2,"pick":4}},"expected":{"Run":-9}},
{"inputs":{"Clip":{"readings":[127,131,137],"labels":["yes","yes","yes"],"flags":[true,false,true],"lower":1,"upper":3,"pick":0}},"expected":{"Run":-7}},
{"inputs":{"Clip":{"readings":[193,197],"labels":["yes","yes"],"flags":[true,true],"lower":1,"upper":1,"pick":0}},"expected":{"Run":-9}},
{"inputs":{"Clip":{"readings":[139,149],"labels":["yes","skip"],"flags":[true,true],"lower":0,"upper":2,"pick":1}},"expected":{"Run":-7}},
{"inputs":{"Clip":{"readings":[199,211],"labels":["yes","yes"],"flags":[true,true],"lower":0,"upper":2,"pick":-1}},"expected":{"Run":-9}},
{"inputs":{"Clip":{"readings":[37,37],"labels":["fresh","fresh"],"flags":[true,true],"lower":0,"upper":2,"pick":1}},"expected":{"Run":37}}
]}`
	suite, err := DecodeCompositionCases([]byte(cases))
	if err != nil {
		t.Fatal(err)
	}
	program, err := GenerateCompositionWithOptions(context.Background(), "읽은위치9059.gooo", []byte(source), suite, CompositionOptions{EntryActivity: "Run"})
	observed := map[string]any{"source": source, "cases": json.RawMessage(cases), "composition": program}
	defer saveReadObservation(t, "native9059.json", observed)
	if err != nil {
		t.Fatal("new source construction", err)
	}
	runtime, err := ExecuteComposition(context.Background(), "읽은위치9059.gooo", []byte(source), program, suite, "")
	observed["runtime"] = runtime
	if err != nil {
		t.Fatal("new native source", err)
	}
	if runtime.FinitePassed != 7 || runtime.FiniteTotal != 7 || !runtime.RuntimeReplayed || runtime.ModelCalls != 0 {
		t.Fatal("native outcomes", runtime.FinitePassed, runtime.FiniteTotal)
	}
	want := []int{7, 0, 3, 0, 4, 0, 6}
	for c, trace := range runtime.Traces {
		if !trace.ReadsObserved || len(trace.Reads) != want[c] {
			t.Fatalf("case%d reads=%d want%d", c, len(trace.Reads), want[c])
		}
	}
	summary := runtime.ReadObservation
	if summary == nil || summary.RecordedReads != 20 || summary.ResolvedReads != 19 || summary.UnresolvedReads != 1 || summary.Status != "PARTIAL" || summary.DroppedReads != 0 {
		t.Fatal("read denominators", summary)
	}
	for i, event := range runtime.Traces[0].Reads[:3] {
		if event.Index != i || event.Origin == nil || event.Origin.Index != i+1 || runtime.ReadInputs[event.Origin.Input].FieldID != "urn:read9059:request-values" {
			t.Fatal("range preserves sliced input position", event)
		}
	}
	last := runtime.Traces[0].Reads[6]
	if last.Index != 1 || last.Origin == nil || last.Origin.Index != 2 {
		t.Fatal("equal values must retain actual second position", last)
	}
	unknown := runtime.Traces[6].Reads[5]
	if unknown.Origin != nil || unknown.Reason != "NO_REGISTERED_INPUT_ELEMENT" {
		t.Fatal("literal must not inherit equal input value", unknown)
	}
	for _, site := range runtime.ReadSites {
		if site.RootActivityID != "urn:read9059:run" || site.ActivityID != "urn:read9059:at" || site.Source == nil || site.Source.Status != "MAPPED" || site.Source.SourceSHA256 != program.SelectedSourceSHA256 || site.Source.End <= site.Source.Start {
			t.Fatal("source-owned helper read", site)
		}
	}
	facts := CompositionReadFacts(runtime)
	observed["facts"] = facts
	if len(facts) != 20 {
		t.Fatal("Gooo record facts", len(facts))
	}
	originalExpression := facts[0].Expression
	facts[0].Expression = "detached"
	if runtime.ReadSites[runtime.Traces[0].Reads[0].SiteIndex].Source.Expression == "detached" {
		t.Fatal("fact mutation changed runtime")
	}
	facts[0].Expression = originalExpression
	t.Logf("native9059: seven fresh cases7/7;20 successful reads;19 exact input origins;1 constructed sequence unresolved; Integer/Text/Boolean, range/slice/bind/helper/source_fill/short circuit/arithmetic; model0")
}

func saveReadObservation(t *testing.T, name string, observation any) {
	t.Helper()
	if folder := os.Getenv("GOOO_READ_OUTPUT"); folder != "" {
		raw, err := json.MarshalIndent(observation, "", "  ")
		if err != nil {
			t.Error(err)
			return
		}
		if err := os.WriteFile(filepath.Join(folder, name), raw, 0600); err != nil {
			t.Error(err)
		}
	}
}
