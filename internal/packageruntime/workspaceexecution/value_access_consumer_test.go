package workspaceexecution

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
)

func TestGoooExplainsSavedAccess8933(t *testing.T) {
	input := os.Getenv("GOOO_ACCESS_SAVED_INPUT")
	if input == "" {
		input = "../../../docs/observations/collection-access/access8887.json"
	}
	raw, err := os.ReadFile(input)
	if err != nil {
		t.Fatal(err)
	}
	var saved struct {
		Flow ProgramValueOrigins `json:"flow"`
	}
	if err := json.Unmarshal(raw, &saved); err != nil {
		t.Fatal(err)
	}
	facts := ProgramAccessFacts(saved.Flow)
	if len(facts) != 16 || saved.Flow.Truncated || saved.Flow.AccessesTruncated {
		t.Fatal("saved actual access occurrences", len(facts), saved.Flow)
	}
	const source = `package accessexplanation
namespace accessexplanation8933
entity Text id "urn:gooo:type:string"
entity Access id "urn:accessexplanation8933:access" fields {
 field root_id id "urn:accessexplanation8933:root" type string required one
 field activity_id id "urn:accessexplanation8933:activity" type string required one
 field kind id "urn:accessexplanation8933:kind" type string required one
 field expression id "urn:accessexplanation8933:expression" type string required one
 field body_sha256 id "urn:accessexplanation8933:body" type string required one
 field start_byte id "urn:accessexplanation8933:start" type integer required one
 field end_byte id "urn:accessexplanation8933:end" type integer required one
 field snippet_truncated id "urn:accessexplanation8933:truncated" type boolean required one
}
activity Explain(Access) -> Text id "urn:accessexplanation8933:explain" computes "if input.kind == \"index_operand\" { return \"위치를 정하는 식\" }; if input.kind == \"slice_low\" { return \"구간 시작\" }; if input.kind == \"slice_high\" { return \"구간 끝\" }; if input.kind == \"slice_max\" { return \"구간 용량 경계\" }; if input.kind == \"index_access\" { return \"원소 읽기\" }; if input.kind == \"slice_access\" { return \"구간 읽기\" }; if input.kind == \"access_base\" { return \"읽을 목록\" }; return \"필드 읽기\""
`
	want := map[string]string{"field_access": "필드 읽기", "index_operand": "위치를 정하는 식", "index_access": "원소 읽기",
		"slice_low": "구간 시작", "slice_high": "구간 끝", "slice_max": "구간 용량 경계", "slice_access": "구간 읽기", "access_base": "읽을 목록"}
	var rows []map[string]any
	for _, fact := range facts {
		expected, ok := want[fact.Kind]
		if !ok {
			t.Fatal("unknown saved access kind", fact)
		}
		rows = append(rows, map[string]any{"inputs": map[string]any{"Explain": fact}, "expected": map[string]any{"Explain": expected}})
	}
	cases, _ := json.Marshal(map[string]any{"schema": "gooo/body-composition-cases/v1", "cases": rows})
	suite, err := bodyexecution.DecodeCompositionCases(cases)
	if err != nil {
		t.Fatal(err)
	}
	generation, err := bodyexecution.GenerateComposition(context.Background(), "접근설명.gooo", []byte(source), suite, "")
	if err != nil {
		t.Fatal(err)
	}
	runtime, err := bodyexecution.ExecuteComposition(context.Background(), "접근설명.gooo", []byte(source), generation, suite, "")
	if err != nil || runtime.FinitePassed != len(facts) || runtime.FiniteTotal != len(facts) || runtime.ModelCalls != 0 {
		t.Fatal(err, runtime)
	}
	facts[0].Expression = "changed copy"
	if saved.Flow.Accesses[0].Expression == "changed copy" {
		t.Fatal("access facts alias the saved report")
	}
	if output := os.Getenv("GOOO_ACCESS_OUTPUT"); output != "" {
		raw, _ := json.MarshalIndent(map[string]any{"source": source, "cases": json.RawMessage(cases),
			"runtime": runtime, "generation": generation, "saved_input_sha256": sourceSHA256(raw),
			"original_program_executions": 0, "synthetic_facts": 0}, "", "  ")
		if err := os.WriteFile(filepath.Join(output, "consumer8933.json"), raw, 0600); err != nil {
			t.Fatal(err)
		}
	}
	t.Log("8933: Gooo describes16 actual saved source access occurrences; original program executions0, model0; detached facts preserve saved report")
}
