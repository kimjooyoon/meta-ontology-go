package workspaceexecution

import (
	"context"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime"
	"testing"
)

func TestScalarCriteriaWorkspaceBindsSeparatePredicateSource(t *testing.T) {
	m := observedWorkspace()
	m.Packages[0].Sources = []packageruntime.Source{
		{Filename: "fee.gooo", Content: `package observed
namespace observed
entity Integer id "fee-criteria://integer"
entity Boolean id "fee-criteria://boolean"
activity Fee(Integer) -> Integer id "fee-criteria://fee" computes "return 11 - input" assembling {
 choice "direction" operand_order at "0" intent "입력에서 11을 뺀다. Subtract eleven from the input."
 criteria {
  predicate "difference" activity "Difference"
  input "[11]"
  holdout_input "[20]"
 }
 observe {
  criteria_strategy "distinguish_then_check"
  criteria_input "[17]"
  max_candidates "2"
  max_rounds "1"
 }
 attempts "2"
}
`},
		{Filename: "predicate.gooo", Content: `package observed
namespace observed
activity Difference(Integer,Integer)->Boolean computes "return input1 == input0-11"`},
	}
	p, err := Prepare(m)
	if err != nil {
		t.Fatal(err)
	}
	g, err := bodycodegen.NewTypedPathGenerator("")
	if err != nil {
		t.Fatal(err)
	}
	result, err := g.GenerateSourceAssembly(context.Background(), "workspace.gooo", []byte(p.Source), p.Entry.LoweredName)
	if err != nil {
		t.Fatal(err)
	}
	a := result.Report.RecordAssembly
	if result.Report.OutputType != "int64" || a.SelectedMask != 1 || a.Criteria.Passed != 2 || a.CriteriaHoldout.Passed != 1 {
		t.Fatal(a)
	}
}
