package bodycodegen

import (
	"context"
	"encoding/json"
	"testing"
)

func TestRetainedBranchLocalUse11053(t *testing.T) {
	source := []byte(`package use11053
namespace use11053
entity Integer id "urn:gooo:type:integer"
activity Fold(Integer) -> Integer id "urn:use11053:fold" computes "let carry=__GOOO_BODY_HOLE_carry__;let previous=carry;if input<0 {carry=carry+input}else{carry=carry-input};return previous+carry" assembling {
 source_fill intent "분기별 대입과 사용을 구분하고 생략된 과거 좌표는 그대로 둔다" {
  hole "carry"
  candidate "carry11053" {fill "carry" "367"}
  candidate "other11053" {fill "carry" "369"}
 }
 case "373" -> "361"
 case "-379" -> "355"
}
`)
	prior := generateOriginFill3371(t, "use11053.gooo", source, "Fold")
	if len(prior.Report.BodyLocals) != 4 || len(prior.Report.BodyLocals[0].Uses) != 4 {
		t.Fatal("fresh branch write/use population", prior.Report.BodyLocals)
	}
	encoded, err := json.Marshal(prior)
	if err != nil {
		t.Fatal(err)
	}
	var absent Result
	if err := json.Unmarshal(encoded, &absent); err != nil {
		t.Fatal(err)
	}
	for i := range absent.Report.BodyLocals {
		absent.Report.BodyLocals[i].Uses = nil
	}
	realized, err := RealizeSourceAssembly(context.Background(), "use11053.gooo", source, absent)
	if err != nil || realized.ModelCalls != 0 {
		t.Fatal("absent optional coordinates stopped branch realization", err)
	}
	for _, origin := range absent.Report.BodyLocals {
		if origin.Uses != nil {
			t.Fatal("historical use claim fabricated")
		}
	}
	checkRetainedUseContradictions11017(t, encoded, prior)
	after, err := json.Marshal(prior)
	if err != nil || string(encoded) != string(after) {
		t.Fatal("comparison mutated current metadata", err)
	}
	saveLocalOrigin9811(t, "use11053.json", map[string]any{"source": string(source), "generation": prior,
		"optional_use_absence": "preserved", "contradictions_rejected": 4, "native_executions": 0, "model_calls": 0})
	t.Log("fresh branch source_fill two cases; four writes; four coordinate contradictions rejected; omitted uses stay absent; native0/model0")
}
