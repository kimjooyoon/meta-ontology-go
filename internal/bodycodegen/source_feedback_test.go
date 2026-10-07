package bodycodegen

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"testing"
)

func TestSourceFeedbackPromotesHoldoutAndPreservesExactIntegers(t *testing.T) {
	source := []byte(`package offset
namespace offset
entity Integer id "offset://integer"
activity Add(Integer) -> Integer computes "return __GOOO_BODY_HOLE_value__" assembling {
    search hole "value" grammar "integer-offset-constant/v1" intent "Add one." max_candidates "16"
    case "2" -> "3"
    holdout_case "9007199254740993" -> "9007199254740994"
    attempts "8"
}`)
	row := AssemblyFeedback{Inputs: []json.RawMessage{json.RawMessage(`9007199254740993`)}, Expected: json.RawMessage(`9007199254740994`)}
	ctx := context.Background()
	next, added, err := ExtendAssemblyCases(ctx, "source.gooo", source, "Add", []AssemblyFeedback{row, row})
	if err != nil || added != 1 {
		t.Fatal("feedback promotion failed", err, added)
	}
	spec, err := SourceAssembly(ctx, "source.gooo", next, "Add")
	if err != nil || len(spec.Cases) != 2 || len(spec.HoldoutCases) != 0 || spec.Cases[1].Input != 9007199254740993 {
		t.Fatal("promoted case lost its identity or remained a holdout", err)
	}
	same, added, err := ExtendAssemblyCases(ctx, "source.gooo", next, "Add", []AssemblyFeedback{row})
	if err != nil || added != 0 || !bytes.Equal(next, same) {
		t.Fatal("duplicate feedback changed source", err)
	}
	row.Expected = json.RawMessage(`0`)
	if _, _, err := ExtendAssemblyCases(ctx, "source.gooo", source, "Add", []AssemblyFeedback{row}); err == nil {
		t.Fatal("feedback silently replaced an existing expected answer")
	}
}

func TestSourceFeedbackRecordCasesDeduplicateJSONKeyOrder(t *testing.T) {
	source, err := os.ReadFile("../../examples/body-codegen/record-candidate-continuation.gooo.fixture")
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	row := AssemblyFeedback{Inputs: []json.RawMessage{json.RawMessage(`{"reason":"review","state":"queued","title":"English"}`), json.RawMessage(`true`)},
		Expected: json.RawMessage(`{"reason":"English:ready","state":"ready","title":"English"}`)}
	next, added, err := ExtendAssemblyCases(ctx, "source.gooo", source, "Select", []AssemblyFeedback{row})
	if err != nil || added != 0 || !bytes.Equal(next, source) {
		t.Fatal("object order became a new source case", err)
	}
	row.Expected = json.RawMessage(`{"reason":"bad","state":"ready","title":"English"}`)
	if _, _, err := ExtendAssemblyCases(ctx, "source.gooo", source, "Select", []AssemblyFeedback{row}); err == nil {
		t.Fatal("contradictory record feedback replaced source intent")
	}
	if got, err := canonicalFeedbackJSON([]byte(`{"n":9007199254740993}`)); err != nil || got != `{"n":9007199254740993}` {
		t.Fatal("canonical feedback lost integer precision", got, err)
	}
}
