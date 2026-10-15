package bodyexecution

import (
	"context"
	"encoding/json"
	"testing"
)

const mixedDivisionSource = `package mixeddivision
namespace mixeddivision
entity Integer id "urn:gooo:type:integer"
entity Boolean id "urn:gooo:type:boolean"
entity State id "urn:mixeddivision:state" fields {
field count id "urn:mixeddivision:count" type integer required one
field divisor id "urn:mixeddivision:divisor" type integer required one
}
activity Main(기록 State, 먼저 Boolean) -> Integer id "urn:mixeddivision:main" computes ` + "`" + `
let value = 기록.count
let copy = 기록
if 먼저 {
value /= 기록.divisor
copy.count /= value
} else {
copy.count /= 기록.divisor
value /= copy.count
}
return copy.count + value
` + "`\n"

func TestMixedLocalFieldDivisionsRetainSeparateOrigins(t *testing.T) {
	suite := CompositionCases{Schema: "gooo/body-composition-cases/v1"}
	for _, row := range [][3]string{
		{`{"count":10,"divisor":2}`, "true", "7"}, {`{"count":10,"divisor":2}`, "false", "7"},
		{`{"count":0,"divisor":2}`, "true", "0"}, {`{"count":1,"divisor":2}`, "false", "1"},
	} {
		suite.Cases = append(suite.Cases, CompositionCase{Inputs: map[string]json.RawMessage{
			"Main.input0": json.RawMessage(row[0]), "Main.input1": json.RawMessage(row[1])},
			Expected: map[string]json.RawMessage{"Main": json.RawMessage(row[2])}})
	}
	observation := map[string]any{"source": mixedDivisionSource, "cases": suite}
	defer saveReadObservation(t, "mixed-division-faults.json", observation)
	program, err := GenerateCompositionWithOptions(context.Background(), "mixed-division.gooo", []byte(mixedDivisionSource), suite, CompositionOptions{EntryActivity: "Main"})
	if err != nil {
		t.Fatal(err)
	}
	observation["composition"] = program
	r, err := ExecuteComposition(context.Background(), "mixed-division.gooo", []byte(mixedDivisionSource), program, suite, "")
	observation["runtime"] = r
	if err != nil || r.FinitePassed != 2 || r.Outcomes == nil || r.Outcomes.Faulted != 2 || len(r.FaultSites) != 4 || r.ModelCalls != 0 {
		t.Fatal("mixed division results", err, r.Outcomes)
	}
	for i, variable := range []string{"", "value"} {
		fault := r.Traces[i+2].Deliveries[0].Fault
		if fault == nil || fault.Site.Assignment != "/=" || fault.Site.Variable != variable || fault.Left != int64(i) || fault.Right != 0 {
			t.Fatal("local and field source sites conflated", i, fault)
		}
	}
}
