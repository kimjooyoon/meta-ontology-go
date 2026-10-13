package bodyexecution

import (
	"context"
	"encoding/json"
	"testing"
)

func TestLiteralDecisions9281(t *testing.T) {
	const source = `package literaldecisions
namespace literal9281
entity Integer id "urn:gooo:type:integer"
entity Boolean id "urn:gooo:type:boolean"
entity Text id "urn:gooo:type:string"
entity Request id "urn:literal9281:request" fields {
 field left id "urn:literal9281:left" type string required one
 field right id "urn:literal9281:right" type string required one
 field flag id "urn:literal9281:flag" type boolean required one
 field position id "urn:literal9281:position" type integer required one
}
activity Label(Request) -> Text id "urn:literal9281:label" computes "return []string{input.left,input.right}[__GOOO_BODY_HOLE_pick__]" assembling {
 source_fill intent "요청한 위치의 이름을 구성한 목록에서 읽는다" {
  hole "pick"
  candidate "first9281" { fill "pick" "0" }
  candidate "requested9281" { fill "pick" "input.position" }
 }
 value_case "[{\"left\":\"전\",\"right\":\"후\",\"flag\":true,\"position\":1}]" -> "\"후\""
 holdout_value_case "[{\"left\":\"위\",\"right\":\"아래\",\"flag\":false,\"position\":0}]" -> "\"위\""
}
activity Flag(Request) -> Boolean id "urn:literal9281:flag-result" computes "return []bool{input.flag,false}[__GOOO_BODY_HOLE_pick__]" assembling {
 source_fill intent "입력에서 받은 불리언을 선택한다" {
  hole "pick"
  candidate "false9281" { fill "pick" "1" }
  candidate "input9281" { fill "pick" "0" }
 }
 value_case "[{\"left\":\"자\",\"right\":\"컴퍼스\",\"flag\":true,\"position\":1}]" -> "true"
 holdout_value_case "[{\"left\":\"desk\",\"right\":\"chair\",\"flag\":false,\"position\":0}]" -> "false"
}
activity Seed(Request) -> Request id "urn:literal9281:seed" computes "return input"
activity Empty(Request) -> Integer id "urn:literal9281:empty" computes "let names = []string{}; return int64(len(names))"
bind Seed.result -> Label.input
bind Seed.result -> Flag.input
bind Seed.result -> Empty.input
`
	suite := CompositionCases{Schema: "gooo/body-composition-cases/v1"}
	for _, probe := range []struct{ input, text, flag string }{
		{`{"left":"모래","right":"돌","flag":true,"position":1}`, `"돌"`, `true`},
		{`{"left":"들판","right":"언덕","flag":false,"position":0}`, `"들판"`, `false`},
		{`{"left":"rock","right":"sand","flag":true,"position":1}`, `"sand"`, `true`},
	} {
		suite.Cases = append(suite.Cases, CompositionCase{Inputs: map[string]json.RawMessage{"Seed": json.RawMessage(probe.input)}, Expected: map[string]json.RawMessage{"Label": json.RawMessage(probe.text), "Flag": json.RawMessage(probe.flag), "Empty": json.RawMessage(`0`)}})
	}
	program, err := GenerateComposition(context.Background(), "목록결정9281.gooo", []byte(source), suite, "")
	observed := map[string]any{"source": source, "cases": suite, "composition": program}
	defer saveReadObservation(t, "literal9281.json", observed)
	if err != nil {
		t.Fatal("typed literal selection", err)
	}
	runtime, err := ExecuteComposition(context.Background(), "목록결정9281.gooo", []byte(source), program, suite, "")
	observed["runtime"] = runtime
	if err != nil || runtime.FinitePassed != 9 || runtime.FiniteTotal != 9 {
		t.Fatal("three fresh literal programs", err, runtime.FinitePassed)
	}
	if runtime.ReadObservation == nil || runtime.ReadObservation.RecordedReads != 6 || runtime.ReadObservation.UnresolvedReads != 0 || runtime.ReadObservation.ConstructedReads != 6 || len(runtime.ReadInputs) != 0 {
		t.Fatal("constructed values carry separate construction origins", runtime.ReadObservation)
	}
	for _, site := range runtime.ReadSites {
		if site.Source == nil || site.Source.Status != "MAPPED" {
			t.Fatal("literal source range", site)
		}
	}
	t.Log("literal9281: Text/Boolean selected bodies, empty Text collection, native9/9; six constructed reads linked to creation sites; model0")
}
