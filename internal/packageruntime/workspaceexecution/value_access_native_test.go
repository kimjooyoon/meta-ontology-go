package workspaceexecution

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime"
)

func TestWorkspaceCollectionAccess8887(t *testing.T) {
	const source = `package readings
namespace access8887
entity Integer id "urn:gooo:type:integer"
entity Request id "urn:access8887:request" fields {
 field readings id "urn:access8887:request-readings" type integer required many
 field lower id "urn:access8887:lower" type integer required one
 field upper id "urn:access8887:upper" type integer required one
 field pick id "urn:access8887:request-pick" type integer required one
}
entity Window id "urn:access8887:window" fields {
 field readings id "urn:access8887:window-readings" type integer required many
 field pick id "urn:access8887:window-pick" type integer required one
}
activity TakeWindow(Request) -> Window id "urn:access8887:take" computes "let 시작 = input.lower; let 끝 = input.upper; return Window{readings: input.readings[시작:끝], pick: input.pick}"
activity At(Window) -> Integer id "urn:access8887:at" computes "if input.pick < 0 || input.pick >= int64(len(input.readings)) { return -1 }; return input.readings[__GOOO_BODY_HOLE_position__]" assembling {
 source_fill intent "선택한 구간에서 요청한 위치의 측정값을 읽는다" {
  hole "position"
  candidate "first8887" { fill "position" "0" }
  candidate "requested8887" { fill "position" "input.pick" }
 }
 value_case "[{\"readings\":[31,47,73],\"pick\":1}]" -> "47"
 value_case "[{\"readings\":[97,41],\"pick\":1}]" -> "41"
 holdout_value_case "[{\"readings\":[67,103,53],\"pick\":2}]" -> "53"
}
activity Read(Window) -> Integer id "urn:access8887:read" computes "return At(input)"
bind TakeWindow.result -> Read.input
`
	const cases = `{"schema":"gooo/body-composition-cases/v1","cases":[
 {"inputs":{"access8887:TakeWindow":{"readings":[193,197,199,211],"lower":1,"upper":4,"pick":1}},"expected":{"access8887:TakeWindow":{"readings":[197,199,211],"pick":1},"access8887:Read":199}},
 {"inputs":{"access8887:TakeWindow":{"readings":[223,227,229,233],"lower":0,"upper":3,"pick":0}},"expected":{"access8887:TakeWindow":{"readings":[223,227,229],"pick":0},"access8887:Read":223}},
 {"inputs":{"access8887:TakeWindow":{"readings":[-41,23,31,61,71],"lower":1,"upper":5,"pick":3}},"expected":{"access8887:TakeWindow":{"readings":[23,31,61,71],"pick":3},"access8887:Read":71}},
 {"inputs":{"access8887:TakeWindow":{"readings":[79,79,79],"lower":0,"upper":3,"pick":2}},"expected":{"access8887:TakeWindow":{"readings":[79,79,79],"pick":2},"access8887:Read":79}},
 {"inputs":{"access8887:TakeWindow":{"readings":[101,103,107],"lower":2,"upper":2,"pick":0}},"expected":{"access8887:TakeWindow":{"readings":[],"pick":0},"access8887:Read":-1}}
 ]}`
	suite, err := bodyexecution.DecodeCompositionCases([]byte(cases))
	if err != nil {
		t.Fatal(err)
	}
	manifest := packageruntime.Manifest{Schema: packageruntime.ManifestSchema,
		Entry: packageruntime.EntrySpec{PackagePath: "access8887", Activity: "Read"},
		Packages: []packageruntime.PackageSpec{{Path: "access8887", Name: "readings",
			Sources: []packageruntime.Source{{Filename: "측정구간.gooo", Content: source}}}},
	}
	result, err := ExecuteWorkspace(context.Background(), manifest, suite, "", "")
	if err != nil {
		t.Fatal("fresh collection source", err)
	}
	observed := map[string]any{"source": source, "manifest": manifest, "cases": json.RawMessage(cases), "result": result}
	defer func() {
		if output := os.Getenv("GOOO_ACCESS_OUTPUT"); output != "" {
			raw, _ := json.MarshalIndent(observed, "", "  ")
			if err := os.WriteFile(filepath.Join(output, "access8887.json"), raw, 0600); err != nil {
				t.Error(err)
			}
		}
	}()
	if result.Runtime.FinitePassed != 10 || result.Runtime.FiniteTotal != 10 || result.Runtime.ModelCalls != 0 {
		t.Fatal("five requests and ten activity outputs", result.Runtime)
	}
	before, _ := json.Marshal(result)
	flow, err := InspectSelectedValueOrigins(context.Background(), result, OriginQuery{Limit: 128})
	if err != nil || flow.Truncated {
		t.Fatal("source access inspection", err, flow)
	}
	observed["flow"] = flow
	facts := ProgramAccessFacts(flow)
	observed["access_facts"] = facts
	want := map[string]bool{"index_operand:input.pick": false, "slice_low:시작": false, "slice_high:끝": false,
		"slice_access:input.readings[시작:끝]": false, "index_access:input.readings[input.pick]": false}
	for _, fact := range facts {
		key := fact.Kind + ":" + fact.Expression
		if _, present := want[key]; present {
			want[key] = true
		}
		if fact.BodySHA256 == "" || fact.Start < 0 || fact.End <= fact.Start || fact.Truncated {
			t.Fatal("source access fact identity", fact)
		}
	}
	for key, found := range want {
		if !found {
			t.Fatal("missing access fact", key, facts)
		}
	}
	linked, original := false, false
	for _, row := range flow.Rows {
		vias := map[string]bool{}
		for _, step := range row.Path {
			vias[step.Via] = true
		}
		linked = linked || vias["call_return"] && vias["call_argument"] && vias["bind"]
		if row.CandidateID == "requested8887" {
			for _, location := range row.Locations {
				original = original || location.View == "workspace_input_file" && location.Filename == "측정구간.gooo" &&
					location.SourceSHA256 == sourceSHA256([]byte(source)) && strings.Contains(source[location.Start:location.End], "input.pick")
			}
		}
	}
	if !linked || !original {
		t.Fatal("call/bind path and selected index source", linked, original)
	}
	after, _ := json.Marshal(result)
	if !bytes.Equal(before, after) {
		t.Fatal("access query changed saved receipt")
	}
	if output := os.Getenv("GOOO_ACCESS_OUTPUT"); output != "" {
		raw, _ := json.MarshalIndent(map[string]any{"source": source, "manifest": manifest, "cases": json.RawMessage(cases),
			"result": result, "flow": flow, "access_facts": facts, "model_calls": 0}, "", "  ")
		if err := os.WriteFile(filepath.Join(output, "access8887.json"), raw, 0600); err != nil {
			t.Fatal(err)
		}
	}
	t.Logf("8887: native10/10 across5 requests, identical values and empty window included; access facts=%d; call/bind links and original selected index retained; receipt unchanged; model0", len(facts))
}
