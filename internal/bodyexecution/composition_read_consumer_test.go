package bodyexecution

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestReadFactsGoooConsumer9127(t *testing.T) {
	path := os.Getenv("GOOO_READ_PRIOR")
	if path == "" {
		path = filepath.Join("..", "..", "docs", "observations", "native-read-origins", "native9059.json")
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var saved struct {
		Runtime CompositionRuntime `json:"runtime"`
	}
	if err = json.Unmarshal(raw, &saved); err != nil {
		t.Fatal(err)
	}
	const source = `package originview
namespace originview9127
entity Text id "urn:gooo:type:string"
entity Read id "urn:read9127:fact" fields {
 field case_index id "urn:read9127:case" type integer required one
 field root_activity_id id "urn:read9127:root" type string required one
 field activity_id id "urn:read9127:activity" type string required one
 field kind id "urn:read9127:kind" type string required one
 field index id "urn:read9127:index" type integer required one
 field length id "urn:read9127:length" type integer required one
 field status id "urn:read9127:status" type string required one
 field input_key id "urn:read9127:input" type string required one
 field field_id id "urn:read9127:field" type string required one
 field origin_index id "urn:read9127:origin" type integer required one
 field source_sha256 id "urn:read9127:sha" type string required one
 field source_start id "urn:read9127:start" type integer required one
 field source_end id "urn:read9127:end" type integer required one
 field expression id "urn:read9127:expression" type string required one
}
activity Explain(Read) -> Text id "urn:read9127:explain" computes "if input.status != \"RESOLVED\" { return \"입력 기원 미연결\" }; if input.index != input.origin_index { return \"슬라이스 위치 이동\" }; return \"입력 위치 유지\""
`
	facts := CompositionReadFacts(saved.Runtime)
	if len(facts) != 20 {
		t.Fatal("saved observation must contain20 reads", len(facts))
	}
	suite := CompositionCases{Schema: "gooo/body-composition-cases/v1"}
	counts := map[string]int{}
	for _, fact := range facts {
		label := "입력 위치 유지"
		if fact.Status != "RESOLVED" {
			label = "입력 기원 미연결"
		} else if fact.Index != fact.OriginIndex {
			label = "슬라이스 위치 이동"
		}
		counts[label]++
		input, _ := json.Marshal(fact)
		expected, _ := json.Marshal(label)
		suite.Cases = append(suite.Cases, CompositionCase{Inputs: map[string]json.RawMessage{"Explain": input}, Expected: map[string]json.RawMessage{"Explain": expected}})
	}
	if counts["입력 기원 미연결"] != 1 || counts["슬라이스 위치 이동"] != 10 || counts["입력 위치 유지"] != 9 {
		t.Fatal("saved positional distinctions", counts)
	}
	program, err := GenerateComposition(context.Background(), "기원표시9127.gooo", []byte(source), suite, "")
	if err != nil {
		t.Fatal(err)
	}
	runtime, err := ExecuteComposition(context.Background(), "기원표시9127.gooo", []byte(source), program, suite, "")
	saveReadObservation(t, "consumer9127.json", map[string]any{"source": source, "cases": suite, "composition": program, "runtime": runtime, "counts": counts, "original_program_executions": 0})
	if err != nil || runtime.FinitePassed != 20 || runtime.FiniteTotal != 20 {
		t.Fatal("Gooo consumption", err, runtime.FinitePassed)
	}
	after, err := os.ReadFile(path)
	if err != nil || string(raw) != string(after) {
		t.Fatal("saved original changed", err)
	}
	t.Log("consumer9127: Gooo classified20/20 saved reads:10 shifted positions,9 unchanged,1 unresolved; original execution0; model0")
}
