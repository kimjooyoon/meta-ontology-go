package workspaceexecution

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime"
)

func TestSelectedClosureCallFramesAndBind6373(t *testing.T) {
	const core = `package core
namespace transfer6373
entity Integer id "urn:gooo:type:integer"
activity Offset(Integer) -> Integer id "urn:transfer6373:bump" computes "return input + __GOOO_BODY_HOLE_delta__" assembling {
 source_fill intent "수량에 고정 보정치를 더한다" {
  hole "delta"
  candidate "plus6373" { fill "delta" "19" }
  candidate "zero6373" { fill "delta" "0" }
 }
 case "5" -> "24"
}
activity SelectLeft(Integer,Integer) -> Integer id "urn:transfer6373:first" computes "return input0"
`
	const app = `package app
namespace transferapp6373
import core "transfer6373/core"
activity Start(Integer) -> Integer id "urn:transfer6373:start" computes "return core.Offset(input)"
activity Main(Integer,Integer) -> Integer id "urn:transfer6373:main" computes "let first = core.SelectLeft(input0, input1); let second = core.SelectLeft(input1, input0); return first + second"
bind Start.result -> Main.input0
`
	manifest := packageruntime.Manifest{Schema: packageruntime.ManifestSchema, Entry: packageruntime.EntrySpec{PackagePath: "transfer6373/app", Activity: "Main"}, Packages: []packageruntime.PackageSpec{
		{Path: "transfer6373/core", Name: "core", Sources: []packageruntime.Source{{Filename: "보정.gooo", Content: core}}},
		{Path: "transfer6373/app", Name: "app", Imports: []string{"transfer6373/core"}, Sources: []packageruntime.Source{{Filename: "계산.gooo", Content: app}}},
	}}
	suite, err := bodyexecution.DecodeCompositionCases([]byte(`{"schema":"gooo/body-composition-cases/v1","cases":[{"inputs":{"transfer6373/app:Start":55,"transfer6373/app:Main.input1":6},"expected":{"transfer6373/app:Start":74,"transfer6373/app:Main":80}},{"inputs":{"transfer6373/app:Start":-31,"transfer6373/app:Main.input1":3},"expected":{"transfer6373/app:Start":-12,"transfer6373/app:Main":-9}}]}`))
	if err != nil {
		t.Fatal(err)
	}
	result, err := ExecuteWorkspace(context.Background(), manifest, suite, "", "")
	if err != nil {
		t.Fatal(err)
	}
	if result.Runtime.FinitePassed != 4 || result.Runtime.FiniteTotal != 4 || result.Runtime.ModelCalls != 0 {
		t.Fatal("fresh native outputs", result.Runtime.FinitePassed, result.Runtime.FiniteTotal)
	}
	before, _ := json.Marshal(result)
	report, err := InspectRecordedOrigins(context.Background(), result, OriginQuery{Follow: true, Limit: 128})
	if err != nil {
		t.Fatal(err)
	}
	flow := report.ValueFlow
	if flow == nil || flow.Analysis != "MAPPED" || flow.Truncated {
		t.Fatal("flow status", flow)
	}
	external := map[string]bool{}
	candidates := 0
	for _, row := range flow.Rows {
		if row.Kind == "external_input" {
			external[row.ActivityID] = true
		}
		if row.Kind != "candidate" {
			continue
		}
		candidates++
		if row.ActivityID != "urn:transfer6373:bump" || row.CandidateID != "plus6373" || row.HoleID != "delta" {
			t.Fatal("candidate source", row)
		}
		vias := map[string]bool{}
		for _, step := range row.Path {
			vias[step.Via] = true
		}
		if !vias["bind"] || !vias["call_return"] || !vias["call_argument"] {
			t.Fatal("cross-activity links absent", row)
		}
		original := false
		for _, location := range row.Locations {
			if location.View == "workspace_input_file" {
				original = true
				if location.Filename != "보정.gooo" || location.SourceSHA256 != sourceSHA256([]byte(core)) || !strings.Contains(core[location.Start:location.End], "19") {
					t.Fatal("return-only original source", location)
				}
			}
		}
		if !original {
			t.Fatal("return-only original location absent", row)
		}
	}
	if candidates != 1 || len(external) != 2 || !external["urn:transfer6373:start"] || !external["urn:transfer6373:main"] {
		t.Fatal("distinct argument/bind origins", candidates, external)
	}
	after, _ := json.Marshal(result)
	if string(before) != string(after) {
		t.Fatal("query changed receipt")
	}
	var saved Result
	if err := json.Unmarshal(before, &saved); err != nil {
		t.Fatal(err)
	}
	filtered, err := InspectSelectedValueOrigins(context.Background(), saved, OriginQuery{CandidateID: "plus6373", Limit: 1})
	if err != nil || len(filtered.Rows) != 1 || filtered.Truncated {
		t.Fatal("saved selected candidate", err, filtered)
	}
	if output := os.Getenv("GOOO_TRANSFER_OUTPUT"); output != "" {
		raw, _ := json.MarshalIndent(map[string]any{"manifest": manifest, "cases": suite, "result": result, "report": report}, "", "  ")
		if err := os.WriteFile(filepath.Join(output, "transfer6373.json"), raw, 0600); err != nil {
			t.Fatal(err)
		}
	}
	t.Logf("6373: native4/4; candidate origins=%d; distinct external activity origins=%d; original return-only source retained; query read-only; model0", candidates, len(external))
}

func TestUnusedTupleCandidate6389(t *testing.T) {
	const source = `package ignored
namespace ignored6389
entity Integer id "urn:gooo:type:integer"
activity Keep(Integer,Integer) -> Integer id "urn:ignored6389:keep" computes "return input0"
activity Main(Integer,Integer) -> Integer id "urn:ignored6389:main" computes "return Keep(input0, __GOOO_BODY_HOLE_unused__)" assembling {
 source_fill intent "첫 인자를 반환하며 둘째 인자는 계산 결과에 사용하지 않는다" {
  hole "unused"
  candidate "extra6389" { fill "unused" "input1 * 23" }
  candidate "zero6389" { fill "unused" "0" }
 }
 value_case "[2,3]" -> "2"
}
`
	manifest := packageruntime.Manifest{Schema: packageruntime.ManifestSchema, Entry: packageruntime.EntrySpec{PackagePath: "ignored6389", Activity: "Main"}, Packages: []packageruntime.PackageSpec{{Path: "ignored6389", Name: "ignored", Sources: []packageruntime.Source{{Filename: "unused.gooo", Content: source}}}}}
	suite, err := bodyexecution.DecodeCompositionCases([]byte(`{"schema":"gooo/body-composition-cases/v1","cases":[{"inputs":{"ignored6389:Main.input0":97,"ignored6389:Main.input1":23},"expected":{"ignored6389:Main":97}},{"inputs":{"ignored6389:Main.input0":-37,"ignored6389:Main.input1":109},"expected":{"ignored6389:Main":-37}}]}`))
	if err != nil {
		t.Fatal(err)
	}
	result, err := ExecuteWorkspace(context.Background(), manifest, suite, "", "")
	if err != nil {
		t.Fatal(err)
	}
	flow, err := InspectSelectedValueOrigins(context.Background(), result, OriginQuery{Limit: 128})
	if err != nil || result.Runtime.FinitePassed != 2 {
		t.Fatal(err, flow)
	}
	if flow.Analysis != "MAPPED" || len(flow.Rows) != 1 || flow.Rows[0].Kind != "external_input" {
		t.Fatal("unused argument polluted result", flow)
	}
	last := flow.Rows[0].Path[len(flow.Rows[0].Path)-1]
	if last.Expression != "input0" || last.ActivityID != "urn:ignored6389:main" {
		t.Fatal("wrong caller argument", last)
	}
	filtered, err := InspectSelectedValueOrigins(context.Background(), result, OriginQuery{CandidateID: "extra6389", Limit: 32})
	if err != nil || len(filtered.Rows) != 0 {
		t.Fatal("unused candidate claimed as value dependency", err, filtered)
	}
	if output := os.Getenv("GOOO_TRANSFER_OUTPUT"); output != "" {
		raw, _ := json.MarshalIndent(map[string]any{"result": result, "flow": flow, "filtered": filtered}, "", "  ")
		if err := os.WriteFile(filepath.Join(output, "unused6389.json"), raw, 0600); err != nil {
			t.Fatal(err)
		}
	}
	t.Log("6389: native2/2; unused second argument and its candidate do not become returned-value origins")
}
