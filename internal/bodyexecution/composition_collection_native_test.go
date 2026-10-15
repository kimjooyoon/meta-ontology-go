package bodyexecution

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestOrderedIntegerAggregation7013(t *testing.T) {
	const source = `package basket7013
namespace basket7013
entity Integer id "urn:gooo:type:integer"
entity Basket id "urn:basket7013:basket" fields {
 field amounts id "urn:basket7013:amounts" type integer required many
}
activity Total(Basket) -> Integer id "urn:basket7013:total" computes "let total = 0; for _, amount := range input.amounts { total = total + amount }; return total"
`
	const cases = `{"schema":"gooo/body-composition-cases/v1","cases":[
{"inputs":{"Total":{"amounts":[3,-5,11]}},"expected":{"Total":9}},
{"inputs":{"Total":{"amounts":[]}},"expected":{"Total":0}},
{"inputs":{"Total":{"amounts":[9007199254740993,7]}},"expected":{"Total":9007199254741000}}
]}`
	suite, err := DecodeCompositionCases([]byte(cases))
	if err != nil {
		t.Fatal(err)
	}
	generated, err := GenerateComposition(context.Background(), "주문7013.gooo", []byte(source), suite, "")
	if err != nil {
		t.Fatal("generate collection aggregation", err)
	}
	if len(generated.Plan.Records) != 1 || generated.Plan.Records[0].Fields[0].Cardinality != "many" {
		t.Fatal("repeated field disappeared from typed plan")
	}
	runtime, err := ExecuteComposition(context.Background(), "주문7013.gooo", []byte(source), generated, suite, nativeTool())
	if err != nil || runtime.FinitePassed != 3 || runtime.FiniteTotal != 3 || runtime.ModelCalls != 0 {
		t.Fatal("native collection results", err, runtime.FinitePassed, runtime.FiniteTotal, runtime.ModelCalls)
	}
	if directory := os.Getenv("GOOO_COLLECTION_OBSERVATION"); directory != "" {
		raw, err := json.MarshalIndent(map[string]any{"source": source, "cases": json.RawMessage(cases),
			"generation": generated, "runtime": runtime, "historical_source_executions": 0}, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(directory, "aggregation7013.json"), append(raw, '\n'), 0600); err != nil {
			t.Fatal(err)
		}
	}
	t.Log("7013: ordered Integer field, empty array and exact large integer native results3/3; model0")
}
