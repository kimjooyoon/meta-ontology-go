package workspaceexecution

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime"
)

func TestRecordedCandidateQueryAndGoooConsumer5581(t *testing.T) {
	const source = `package amounts
namespace amounts5581
entity Integer id "urn:gooo:type:integer"
activity Amount(Integer) -> Integer id "urn:amounts5581:amount" computes "let passed = input >= __GOOO_BODY_HOLE_gate__; let output = input + __GOOO_BODY_HOLE_adjust__; if passed { output = output * 2 }; return output" assembling {
 source_fill intent "기준을 통과하면 보정 금액의 두 배를 반환한다" {
  hole "gate"
  hole "adjust"
  candidate "amount5581" { fill "gate" "61" fill "adjust" "5" }
  candidate "other5581" { fill "gate" "161" fill "adjust" "0" }
 }
 case "67" -> "144"
}
`
	manifest := packageruntime.Manifest{Schema: packageruntime.ManifestSchema, Entry: packageruntime.EntrySpec{PackagePath: "amounts5581", Activity: "Amount"},
		Packages: []packageruntime.PackageSpec{{Path: "amounts5581", Name: "amounts", Sources: []packageruntime.Source{{Filename: "보정.gooo", Content: source}}}}}
	suite, err := bodyexecution.DecodeCompositionCases([]byte(`{"schema":"gooo/body-composition-cases/v1","cases":[{"inputs":{"amounts5581:Amount":73},"expected":{"amounts5581:Amount":156}},{"inputs":{"amounts5581:Amount":47},"expected":{"amounts5581:Amount":52}}]}`))
	if err != nil {
		t.Fatal(err)
	}
	result, err := ExecuteWorkspace(context.Background(), manifest, suite, "", "")
	if err != nil {
		t.Fatal(err)
	}
	before, _ := json.Marshal(result)
	report, err := InspectRecordedOrigins(context.Background(), result, OriginQuery{CandidateID: "amount5581", Limit: 32})
	if err != nil {
		t.Fatal(err)
	}
	if result.Runtime.FinitePassed != 2 || report.Truncated || len(report.Rows) != 3 || report.OriginsMatched != 1 {
		t.Fatal("actual recorded uses", report)
	}
	facts := OriginUseFacts(report)
	if len(facts) != 5 {
		t.Fatal("root/location scope", facts)
	}
	consumerSource, consumer := consumeOriginUses5581(t, facts)
	after, _ := json.Marshal(result)
	if string(before) != string(after) {
		t.Fatal("read query changed saved construction")
	}
	filtered, err := InspectRecordedOrigins(context.Background(), result, OriginQuery{CandidateID: "absent5581", Limit: 32})
	if err != nil || len(filtered.Rows) != 0 || filtered.OriginsMatched != 0 {
		t.Fatal("candidate filter", err, filtered)
	}
	limited, err := InspectRecordedOrigins(context.Background(), result, OriginQuery{CandidateID: "amount5581", Limit: 1})
	if err != nil || !limited.Truncated || len(limited.Rows) != 1 {
		t.Fatal("bounded report", err, limited)
	}
	if output := os.Getenv("GOOO_ORIGIN_QUERY_OUTPUT"); output != "" {
		raw, _ := json.MarshalIndent(map[string]any{"manifest": manifest, "result": result, "report": report, "facts": facts, "consumer_source": consumerSource, "consumer_runtime": consumer, "synthetic_fixtures": 0}, "", "  ")
		if err := os.WriteFile(filepath.Join(output, "query5581.json"), raw, 0600); err != nil {
			t.Fatal(err)
		}
	}
	t.Log("5581: native2/2; three candidate uses across condition and return; five actual root/location facts consumed by Gooo5/5; filter and truncation explicit; query leaves receipt unchanged")
}

func consumeOriginUses5581(t *testing.T, facts []OriginUseFact) (string, bodyexecution.CompositionRuntime) {
	t.Helper()
	const source = `package usage
namespace usage5581
entity Text id "urn:gooo:type:string"
entity Use id "urn:usage5581:use" fields {
 field activity_id id "urn:usage5581:activity" type string required one
 field candidate_id id "urn:usage5581:candidate" type string required one
 field hole_id id "urn:usage5581:hole" type string required one
 field root_id id "urn:usage5581:root" type string required one
 field kind id "urn:usage5581:kind" type string required one
 field relation id "urn:usage5581:relation" type string required one
 field analysis_status id "urn:usage5581:status" type string required one
 field expression id "urn:usage5581:expression" type string required one
 field source_view id "urn:usage5581:view" type string required one
 field filename id "urn:usage5581:file" type string required one
 field source_start id "urn:usage5581:start" type integer required one
 field source_end id "urn:usage5581:end" type integer required one
 field source_sha id "urn:usage5581:sha" type string required one
}
activity Explain(Use) -> Text id "urn:usage5581:explain" computes "if input.kind == \"return\" { return input.hole_id + \"은 반환값으로 이어짐\" }; return input.hole_id + \"은 조건으로 이어짐\""
`
	var rows []map[string]any
	for _, fact := range facts {
		want := fact.HoleID + "은 조건으로 이어짐"
		if fact.Kind == "return" {
			want = fact.HoleID + "은 반환값으로 이어짐"
		}
		rows = append(rows, map[string]any{"inputs": map[string]any{"Explain": fact}, "expected": map[string]any{"Explain": want}})
	}
	raw, _ := json.Marshal(map[string]any{"schema": "gooo/body-composition-cases/v1", "cases": rows})
	suite, err := bodyexecution.DecodeCompositionCases(raw)
	if err != nil {
		t.Fatal(err)
	}
	prior, err := bodyexecution.GenerateComposition(context.Background(), "사용처.gooo", []byte(source), suite, "")
	if err != nil {
		t.Fatal(err)
	}
	runtime, err := bodyexecution.ExecuteComposition(context.Background(), "사용처.gooo", []byte(source), prior, suite, "")
	if err != nil || runtime.FinitePassed != len(facts) || runtime.ModelCalls != 0 {
		t.Fatal("Gooo origin use interpretation", err, runtime)
	}
	return source, runtime
}
