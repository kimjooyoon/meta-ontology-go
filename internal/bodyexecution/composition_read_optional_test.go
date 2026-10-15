package bodyexecution

import (
	"context"
	"encoding/json"
	"testing"
)

func TestOptionalReadInputs9311(t *testing.T) {
	const source = `package optionalread
namespace optional9311
entity Integer id "urn:gooo:type:integer"
entity Bundle id "urn:optional9311:bundle" fields {
 field values id "urn:optional9311:values" type integer required many
 field notes id "urn:optional9311:notes" type string optional many
}
activity Read(Bundle) -> Integer id "urn:optional9311:read" computes "return input.values[0]"
`
	suite := CompositionCases{Schema: "gooo/body-composition-cases/v1"}
	for _, probe := range []struct{ input, expected string }{
		{`{"values":[379,383]}`, `379`},
		{`{"values":[389,397],"notes":[]}`, `389`},
		{`{"values":[401,409],"notes":["메모"]}`, `401`},
	} {
		suite.Cases = append(suite.Cases, CompositionCase{Inputs: map[string]json.RawMessage{"Read": json.RawMessage(probe.input)}, Expected: map[string]json.RawMessage{"Read": json.RawMessage(probe.expected)}})
	}
	program, err := GenerateComposition(context.Background(), "선택목록9311.gooo", []byte(source), suite, "")
	observed := map[string]any{"source": source, "cases": suite, "composition": program}
	defer saveReadObservation(t, "optional9311.json", observed)
	if err != nil {
		t.Fatal(err)
	}
	runtime, err := ExecuteComposition(context.Background(), "선택목록9311.gooo", []byte(source), program, suite, "")
	observed["runtime"] = runtime
	if err != nil || runtime.FinitePassed != 3 || runtime.ReadObservation == nil || runtime.ReadObservation.ResolvedReads != 3 || runtime.ReadObservation.Status != "COMPLETE" {
		t.Fatal("optional registration", err, runtime.ReadObservation)
	}
	if len(runtime.ReadInputs) != 2 || runtime.ReadInputs[1].Presence != "optional" {
		t.Fatal("optional input identity", runtime.ReadInputs)
	}
	t.Log("optional9311: absent/empty/present optional collection with required reads3/3, resolved3/3; model0")
}
