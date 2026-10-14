package bodyexecution

import (
	"context"
	"encoding/json"
	"testing"
)

func TestLocalOriginRefactor10943(t *testing.T) {
	source := `package local10943
namespace local10943
entity Integer id "urn:gooo:type:integer"
entity Batch id "urn:local10943:batch" fields {field values id "urn:local10943:values" type integer required many}
activity Twice(Integer) -> Integer id "urn:local10943:twice" computes "let changed=input+307;return changed*2"
activity Build(Batch) -> Batch id "urn:local10943:build" computes "let total=311;for current in input.values {let copy=current;total=total+copy};let result=Twice(total);return Batch{values:[]int64{total,result}}"
`
	suite := CompositionCases{Schema: "gooo/body-composition-cases/v1"}
	for _, row := range [][2]string{{`{"values":[313,317]}`, `{"values":[941,2496]}`}, {`{"values":[]}`, `{"values":[311,1236]}`}, {`{"values":[-331]}`, `{"values":[-20,574]}`}} {
		suite.Cases = append(suite.Cases, CompositionCase{Inputs: map[string]json.RawMessage{"Build": json.RawMessage(row[0])}, Expected: map[string]json.RawMessage{"Build": json.RawMessage(row[1])}})
	}
	observation := map[string]any{"source": source, "cases": suite}
	defer saveReadObservation(t, "local-refactor10943.json", observation)
	composition, err := GenerateCompositionWithOptions(context.Background(), "local10943.gooo", []byte(source), suite, CompositionOptions{EntryActivity: "Build"})
	if err != nil {
		t.Fatal(err)
	}
	observation["composition"] = composition
	runtime, err := ExecuteComposition(context.Background(), "local10943.gooo", []byte(source), composition, suite, "")
	observation["runtime"] = runtime
	if err != nil || runtime.FinitePassed != 3 || !runtime.RuntimeReplayed || runtime.LocalObservation.Status != "COMPLETE" {
		t.Fatal("native local origin behavior", err, runtime.FinitePassed)
	}
	for _, site := range runtime.LocalUseSites {
		if site.Source == nil || site.Source.Status != "MAPPED" || site.Source.SourceSHA256 != runtime.SelectedSourceSHA256 || site.Source.Expression != site.Name {
			t.Fatal("source use", site)
		}
	}
	_, page, err := SelectRecordedNativeOrigins(runtime, NativeOriginQuery{Kind: "local-uses", Limit: 1})
	if err != nil || page.LocalObservation == nil || page.ProducerExecutions != 0 {
		t.Fatal("recorded page", err, page)
	}
	t.Log("fresh range/local-copy/helper/result-array source3/3;selected Gooo uses mapped;observation COMPLETE;source recorder refactoring preserves fresh finite results")
}
