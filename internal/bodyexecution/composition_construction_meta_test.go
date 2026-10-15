package bodyexecution

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestConstructionSourceFill9661(t *testing.T) {
	const source = `package filledcollection
namespace filled9661
entity Integer id "urn:gooo:type:integer"
entity Request id "urn:filled9661:request" fields {
 field values id "urn:filled9661:values" type integer required many
 field pick id "urn:filled9661:pick" type integer required one
}
activity Delta(Request) -> Integer id "urn:filled9661:delta" computes "let choices=[]int64{input.values[0],__GOOO_BODY_HOLE_delta__};return choices[input.pick]" assembling {
 source_fill intent "두 번째 원소는 원본 두 값의 합을 계산한다" {
  hole "delta"
  candidate "reverse9661" { fill "delta" "input.values[0]-input.values[1]" }
  candidate "forward9661" { fill "delta" "input.values[1]+input.values[0]" }
 }
 value_case "[{\"values\":[13,17],\"pick\":1}]" -> "30"
 holdout_value_case "[{\"values\":[19,23],\"pick\":1}]" -> "42"
}
`
	suite := CompositionCases{Schema: "gooo/body-composition-cases/v1", Cases: []CompositionCase{
		{Inputs: map[string]json.RawMessage{"Delta": json.RawMessage(`{"values":[647,653],"pick":1}`)}, Expected: map[string]json.RawMessage{"Delta": json.RawMessage(`1300`)}},
		{Inputs: map[string]json.RawMessage{"Delta": json.RawMessage(`{"values":[659,661],"pick":0}`)}, Expected: map[string]json.RawMessage{"Delta": json.RawMessage(`659`)}},
	}}
	program, err := GenerateComposition(context.Background(), "선택된원소9661.gooo", []byte(source), suite, "")
	observed := map[string]any{"source": source, "cases": suite, "composition": program}
	defer saveReadObservation(t, "construction9661.json", observed)
	if err != nil {
		t.Fatal(err)
	}
	runtime, err := ExecuteComposition(context.Background(), "선택된원소9661.gooo", []byte(source), program, suite, "")
	observed["runtime"] = runtime
	if err != nil || runtime.FinitePassed != 2 || len(runtime.ConstructionSites) != 1 {
		t.Fatal("selected source body", err, runtime.ReadObservation)
	}
	location := runtime.ConstructionSites[0].Elements[1].Source
	if location.Status != "MAPPED" || strings.ReplaceAll(location.Expression, " ", "") != "(input.values[1]+input.values[0])" || location.SourceSHA256 != program.SelectedSourceSHA256 {
		t.Fatal("chosen expression location", location)
	}
	for _, trace := range runtime.Traces {
		part := trace.Constructions[0].Elements[1]
		if part.ReadEnd-part.ReadStart != 2 {
			t.Fatal("two actual input reads", part)
		}
	}
}

func TestConstructionGoooConsumer9593(t *testing.T) {
	path := os.Getenv("GOOO_CONSTRUCTION_PRIOR")
	if path == "" {
		path = filepath.Join("..", "..", "docs", "observations", "construction-origins", "construction9533.json")
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var saved struct {
		Runtime CompositionRuntime `json:"runtime"`
	}
	if err := json.Unmarshal(raw, &saved); err != nil {
		t.Fatal(err)
	}
	const source = `package constructionview
namespace constructionview9593
entity Text id "urn:gooo:type:string"
entity Element id "urn:view9593:element" fields {
 field case_index id "urn:view9593:case" type integer required one
 field instance_index id "urn:view9593:instance" type integer required one
 field element_index id "urn:view9593:element-index" type integer required one
 field root_activity_id id "urn:view9593:root" type string required one
 field activity_id id "urn:view9593:activity" type string required one
 field completed id "urn:view9593:completed" type boolean required one
 field read_start id "urn:view9593:read-start" type integer required one
 field read_end id "urn:view9593:read-end" type integer required one
 field reads_dropped id "urn:view9593:lost" type integer required one
 field source_sha256 id "urn:view9593:sha" type string required one
 field source_start id "urn:view9593:start" type integer required one
 field source_end id "urn:view9593:end" type integer required one
 field expression id "urn:view9593:expression" type string required one
 field mapping id "urn:view9593:mapping" type string required one
}
activity Explain(Element) -> Text id "urn:view9593:explain" computes "if !input.completed {return \"계산 중단\"};if input.read_end>input.read_start {return \"계산 중 읽기 연결\"};return \"표현 위치 연결\""
`
	facts := CompositionConstructionFacts(saved.Runtime)
	suite := CompositionCases{Schema: "gooo/body-composition-cases/v1"}
	counts := map[string]int{}
	for _, fact := range facts {
		label := "표현 위치 연결"
		if !fact.Completed {
			label = "계산 중단"
		} else if fact.ReadEnd > fact.ReadStart {
			label = "계산 중 읽기 연결"
		}
		counts[label]++
		input, _ := json.Marshal(fact)
		want, _ := json.Marshal(label)
		suite.Cases = append(suite.Cases, CompositionCase{Inputs: map[string]json.RawMessage{"Explain": input}, Expected: map[string]json.RawMessage{"Explain": want}})
	}
	if len(facts) != 24 || counts["계산 중 읽기 연결"] != 9 || counts["표현 위치 연결"] != 15 {
		t.Fatal("source observations", counts)
	}
	program, err := GenerateComposition(context.Background(), "생성설명9593.gooo", []byte(source), suite, "")
	if err != nil {
		t.Fatal(err)
	}
	runtime, err := ExecuteComposition(context.Background(), "생성설명9593.gooo", []byte(source), program, suite, "")
	saveReadObservation(t, "consumer9593.json", map[string]any{"source": source, "cases": suite, "composition": program, "runtime": runtime, "counts": counts, "original_program_executions": 0})
	if err != nil || runtime.FinitePassed != 24 {
		t.Fatal("Gooo explains its construction facts", err, runtime.FinitePassed)
	}
	after, err := os.ReadFile(path)
	if err != nil || string(after) != string(raw) {
		t.Fatal("prior observation changed", err)
	}
}
