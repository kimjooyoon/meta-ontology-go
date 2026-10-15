package bodyexecution

import (
	"context"
	"encoding/json"
	"testing"
)

func TestRecordCompoundNestedFaults(t *testing.T) {
	const source = `package nesteddivision
namespace nesteddivision
entity State id "urn:nesteddivision:state" fields {
field count id "urn:nesteddivision:count" type integer required one
field divisor id "urn:nesteddivision:divisor" type integer required one
field step id "urn:nesteddivision:step" type integer required one
}
activity Divide(원본 State) -> State id "urn:nesteddivision:divide" computes ` + "`" + `
let copy = 원본
copy.count /= (원본.divisor / 원본.step)
copy.count /= 2
return copy
` + "`" + `
activity Main(기록 State) -> State id "urn:nesteddivision:main" computes "return Divide(기록)"
`
	suite := CompositionCases{Schema: "gooo/body-composition-cases/v1"}
	for i, value := range []string{`{"count":43,"divisor":8,"step":2}`, `{"count":43,"divisor":1,"step":2}`, `{"count":47,"divisor":5,"step":0}`} {
		expected := value
		if i == 0 {
			expected = `{"count":5,"divisor":8,"step":2}`
		}
		suite.Cases = append(suite.Cases, CompositionCase{Inputs: map[string]json.RawMessage{"Main": json.RawMessage(value)},
			Expected: map[string]json.RawMessage{"Main": json.RawMessage(expected)}})
	}
	observation := map[string]any{"source": source, "cases": suite}
	defer saveReadObservation(t, "nested-field-faults.json", observation)
	program, err := GenerateCompositionWithOptions(context.Background(), "nested-field-faults.gooo", []byte(source), suite, CompositionOptions{EntryActivity: "Main"})
	if err != nil {
		t.Fatal(err)
	}
	observation["composition"] = program
	r, err := ExecuteComposition(context.Background(), "nested-field-faults.gooo", []byte(source), program, suite, "")
	observation["runtime"] = r
	if err != nil || r.FinitePassed != 1 || r.Outcomes == nil || r.Outcomes.Faulted != 2 || r.ModelCalls != 0 {
		t.Fatal("nested division outcomes", err, r.Outcomes)
	}
	for i, left := range []int64{43, 5} {
		fault := r.Traces[i+1].Deliveries[0].Fault
		if fault == nil || fault.Left != left || fault.Right != 0 || fault.Site.RootActivityID != "urn:nesteddivision:main" || fault.Site.ActivityID != "urn:nesteddivision:divide" {
			t.Fatal("nested source fault identity", i, fault)
		}
		if (fault.Site.Assignment == "/=") != (i == 0) {
			t.Fatal("inner quotient and outer update were conflated", i, fault)
		}
	}
	if len(r.FaultSites) != 3 {
		t.Fatal("fault site population", len(r.FaultSites))
	}
}
