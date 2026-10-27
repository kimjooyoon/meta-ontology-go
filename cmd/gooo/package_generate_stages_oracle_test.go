package main

const libraryStageSource = `package stages
namespace stages
entity Record id "urn:stages:record" fields {
 field value id "urn:stages:value" type integer required one
}
activity Helper(Record) -> Record id "urn:stages:helper" computes "return Record{value:input.value+1}"
activity Seed(Record) -> Record id "urn:stages:seed" computes ` + "`" + `
let Plan = Helper(input)
return Record{value:Plan.value+9}
` + "`" + `
activity Finish(Record) -> Record id "urn:stages:finish" computes "return Record{value:input.value*2}"
bind Seed.result -> Finish.input
`

const libraryStageOracle = `package stages
import "testing"
func TestIndependentStageAndGraphExpectations(t *testing.T) {
 for _, row := range []struct{ input, plan, decision, graph int64 }{
  {-5,5,-10,10}, {0,10,0,20}, {3,13,6,26},
 } {
  input := R_Record{F_value:row.input}
  if got:=Plan(input);got.F_value!=row.plan {t.Fatal("plan",got,row.plan)}
  if got:=Decide(input);got.F_value!=row.decision {t.Fatal("body-only decision",got,row.decision)}
  if got:=Evaluate(input);got.F_value!=row.graph {t.Fatal("bound graph",got,row.graph)}
  if got:=Decide(Plan(input));got.F_value!=row.graph {t.Fatal("shared record types",got,row.graph)}
 }
}
`
