package bodyexecution

import (
	"context"
	"encoding/json"
	"testing"
)

func TestLocalFillFragmentConsumer11213(t *testing.T) {
	source := []byte(`package fill11213
namespace fill11213
entity Integer id "urn:gooo:type:integer"
activity Build(Integer) -> Integer id "urn:fill11213:build" computes "let seed=__GOOO_BODY_HOLE_seed__;let doubled=seed*2;seed=doubled+input;return seed" assembling {
 source_fill intent "변수의 대입을 그 표현을 조립한 원래 후보와 연결한다" {
  hole "seed"
  candidate "plus11213" {fill "seed" "input+383"}
  candidate "minus11213" {fill "seed" "input-389"}
 }
 case "397" -> "1957"
 case "-401" -> "-437"
}
`)
	suite := calledCompositionCases(t, `{"schema":"gooo/body-composition-cases/v1","cases":[{"inputs":{"Build":409},"expected":{"Build":1993}},{"inputs":{"Build":-419},"expected":{"Build":-491}}]}`)
	prior, err := GenerateComposition(context.Background(), "지역대입11213.gooo", source, suite, "")
	if err != nil {
		t.Fatal(err)
	}
	runtime, err := ExecuteComposition(context.Background(), "지역대입11213.gooo", source, prior, suite, "")
	if err != nil || runtime.FinitePassed != 2 || !runtime.RuntimeReplayed {
		t.Fatal("fresh local fill runtime", err, runtime.FinitePassed)
	}
	rows, page, err := SelectRecordedNativeOrigins(runtime, NativeOriginQuery{Kind: "local-fill-fragments", Limit: 128})
	if err != nil || len(rows) != 6 || page.TotalRows != 6 || page.ProducerExecutions != 0 || page.ProducerModelCalls != 0 {
		t.Fatal("recorded local fill rows", err, len(rows), page)
	}
	consumer := consumeLocalFill11213(t, source, rows)
	for _, site := range runtime.LocalSites {
		if site.Source.Construction == nil || site.Source.Construction.SourceSHA256 != digest(source) {
			t.Fatal("local assembly input origin absent", site)
		}
	}
	saveReadObservation(t, "local-fill11213.json", map[string]any{"source": string(source), "composition": prior,
		"runtime": runtime, "origin_input": page, "rows": rows, "consumer_source": localFillReader11213,
		"consumer_runtime": consumer, "original_program_executions_during_selection": 0})
	t.Log("fresh source_fill producer2/2; six actual write fragments -> original candidate/template; Gooo consumer6/6; selection producer0/models0")
}

func consumeLocalFill11213(t *testing.T, source []byte, rows []json.RawMessage) CompositionRuntime {
	t.Helper()
	suite := CompositionCases{Schema: "gooo/body-composition-cases/v1"}
	for _, row := range rows {
		var fact LocalFillFragmentFact
		if err := json.Unmarshal(row, &fact); err != nil {
			t.Fatal(err)
		}
		want := "template"
		if fact.AssignmentIndex == 0 {
			want = "candidate"
			if fact.Category != "declared_candidate" || fact.HoleID != "seed" ||
				string(source[fact.SourceStart:fact.SourceEnd]) != "input+383" {
				t.Fatal("write did not locate original candidate expression", fact)
			}
		} else if fact.Category != "retained_template" {
			t.Fatal("template write got a candidate claim", fact)
		}
		if fact.CandidateID != "plus11213" || fact.SourceSHA != digest(source) || !fact.Completed {
			t.Fatal("assembly input identity", fact)
		}
		expected, _ := json.Marshal(want)
		suite.Cases = append(suite.Cases, CompositionCase{Inputs: map[string]json.RawMessage{"Describe": row}, Expected: map[string]json.RawMessage{"Describe": expected}})
	}
	prior, err := GenerateComposition(context.Background(), "읽기11213.gooo", []byte(localFillReader11213), suite, "")
	if err != nil {
		t.Fatal(err)
	}
	runtime, err := ExecuteComposition(context.Background(), "읽기11213.gooo", []byte(localFillReader11213), prior, suite, "")
	if err != nil || runtime.FinitePassed != 6 || runtime.ModelCalls != 0 {
		t.Fatal("Gooo fragment interpretation", err, runtime.FinitePassed)
	}
	return runtime
}

const localFillReader11213 = `package reader11213
namespace reader11213
entity Text id "urn:gooo:type:string"
entity Fragment id "urn:reader11213:fragment" fields {
 field case_index id "urn:reader11213:case" type integer required one
 field assignment_index id "urn:reader11213:write" type integer required one
 field frame id "urn:reader11213:frame" type integer required one
 field completed id "urn:reader11213:completed" type boolean required one
 field site_id id "urn:reader11213:site" type string required one
 field variable_name id "urn:reader11213:variable" type string required one
 field activity_id id "urn:reader11213:activity" type string required one
 field candidate_id id "urn:reader11213:candidate" type string required one
 field hole_id id "urn:reader11213:hole" type string required one
 field category id "urn:reader11213:category" type string required one
 field filename id "urn:reader11213:filename" type string required one
 field source_sha id "urn:reader11213:sha" type string required one
 field source_start id "urn:reader11213:start" type integer required one
 field source_end id "urn:reader11213:end" type integer required one
 field selected_decoded_start id "urn:reader11213:decoded-start" type integer required one
 field selected_decoded_end id "urn:reader11213:decoded-end" type integer required one
}
activity Describe(Fragment) -> Text id "urn:reader11213:describe" computes "if input.category==\"declared_candidate\" {return \"candidate\"};if input.category==\"derived_grammar\" {return \"grammar\"};if input.category==\"retained_template\" {return \"template\"};return \"unknown\""
`
