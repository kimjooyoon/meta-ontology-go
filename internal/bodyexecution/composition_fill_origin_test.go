package bodyexecution

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestCompositionFillOriginAndGoooConsumer3607(t *testing.T) {
	source := []byte(`package origin3607
namespace origin3607
entity Integer id "urn:gooo:type:integer"
activity Quote(Integer) -> Integer id "urn:origin3607:quote" computes "let base = __GOOO_BODY_HOLE_seed__; if __GOOO_BODY_HOLE_open__ { return base }; return input - 17" assembling {
 source_fill intent "금액 규칙의 선택 경로를 남긴다" {
  hole "seed"
  hole "open"
  candidate "bounded3607" { fill "seed" "input + 29" fill "open" "input < 61" }
  candidate "other3607" { fill "seed" "input + 3" fill "open" "input < 9" }
 }
 case "19" -> "48"
 case "71" -> "54"
}
`)
	suite := calledCompositionCases(t, `{"schema":"gooo/body-composition-cases/v1","cases":[{"inputs":{"Quote":23},"expected":{"Quote":52}},{"inputs":{"Quote":83},"expected":{"Quote":66}}]}`)
	prior, err := GenerateComposition(context.Background(), "기원3607.gooo", source, suite, "")
	if err != nil {
		t.Fatal(err)
	}
	runtime, err := ExecuteComposition(context.Background(), "기원3607.gooo", source, prior, suite, nativeTool())
	if err != nil || runtime.FinitePassed != 2 || len(runtime.ConditionSites) != 1 {
		t.Fatal("native fill origin", err, runtime.FinitePassed)
	}
	facts := ConditionFillOriginFacts(runtime)
	if len(facts) != 1 || facts[0].CandidateID != "bounded3607" || facts[0].HoleID != "open" ||
		facts[0].Category != "declared_candidate" || facts[0].Filename != "기원3607.gooo" ||
		string(source[facts[0].SourceStart:facts[0].SourceEnd]) != "input < 61" {
		t.Fatal("native condition lost its authored candidate expression", facts)
	}
	consumer := consumeFillOrigins3607(t, facts)
	if output := os.Getenv("GOOO_FILL_ORIGIN_OUTPUT"); output != "" {
		raw, err := json.MarshalIndent(map[string]any{"source": string(source), "composition": prior,
			"runtime": runtime, "facts": facts, "consumer_source": originConsumer3607,
			"actual_fact_inputs": len(facts), "interpretation_fixtures": 3, "consumer_runtime": consumer}, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(output, "native-consumer3607.json"), append(raw, '\n'), 0600); err != nil {
			t.Fatal(err)
		}
	}
	if runtime.ModelCalls != 0 || consumer.ModelCalls != 0 {
		t.Fatal("unexpected model inference")
	}
	t.Log("native3607: producer2/2; condition -> chosen candidate/open hole -> exact input expression; Gooo consumer4/4 (one actual origin + three explicit interpretation fixtures); model0")
}

const originConsumer3607 = `package originconsumer3607
namespace originconsumer3607
entity Text id "urn:gooo:type:string"
entity Fact id "urn:originconsumer3607:fact" fields {
 field site_id id "urn:originconsumer3607:site" type string required one
 field activity_id id "urn:originconsumer3607:activity" type string required one
 field candidate_id id "urn:originconsumer3607:candidate" type string required one
 field hole_id id "urn:originconsumer3607:hole" type string required one
 field category id "urn:originconsumer3607:category" type string required one
 field filename id "urn:originconsumer3607:filename" type string required one
 field source_sha id "urn:originconsumer3607:source" type string required one
 field source_start id "urn:originconsumer3607:start" type integer required one
 field source_end id "urn:originconsumer3607:end" type integer required one
}
activity Describe(Fact) -> Text id "urn:originconsumer3607:describe" computes "if input.category == \"declared_candidate\" { return \"declared\" }; if input.category == \"derived_grammar\" { return \"derived\" }; if input.category == \"retained_template\" { return \"retained\" }; return \"missing\""
`

func consumeFillOrigins3607(t *testing.T, actual []ConditionFillOriginFact) CompositionRuntime {
	t.Helper()
	inputs := append([]ConditionFillOriginFact(nil), actual...)
	want := []string{"declared", "derived", "retained", "missing"}
	for _, category := range []string{"derived_grammar", "retained_template", "unknown"} {
		inputs = append(inputs, ConditionFillOriginFact{Category: category})
	}
	var rows []map[string]any
	for i, fact := range inputs {
		rows = append(rows, map[string]any{"inputs": map[string]any{"Describe": fact}, "expected": map[string]any{"Describe": want[i]}})
	}
	raw, err := json.Marshal(map[string]any{"schema": "gooo/body-composition-cases/v1", "cases": rows})
	if err != nil {
		t.Fatal(err)
	}
	suite := calledCompositionCases(t, string(raw))
	prior, err := GenerateComposition(context.Background(), "originconsumer3607.gooo", []byte(originConsumer3607), suite, "")
	if err != nil {
		t.Fatal(err)
	}
	runtime, err := ExecuteComposition(context.Background(), "originconsumer3607.gooo", []byte(originConsumer3607), prior, suite, nativeTool())
	if err != nil || runtime.FinitePassed != 4 {
		t.Fatal("Gooo origin consumer", err, runtime.FinitePassed)
	}
	return runtime
}
