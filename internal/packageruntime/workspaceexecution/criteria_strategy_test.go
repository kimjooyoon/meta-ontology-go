package workspaceexecution

import (
	"context"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime"
)

func TestCriteriaStrategyWorkspaceRetainsChallengeAndPredicate(t *testing.T) {
	m := observedWorkspace()
	m.Packages[0].Sources = []packageruntime.Source{
		{Filename: "fee.gooo", Content: `package observed
namespace observed
entity Integer id "challenge://integer"
entity Boolean id "challenge://boolean"
entity Box id "challenge://box" fields { field value id "challenge://value" type integer required one }
activity Fee(Integer) -> Box computes "return Box{value:input}" assembling {
 choice "value" field_value at "0" alternative "input+0" intent "Negate."
 criteria { predicate "negative" activity "Negative" input "[0]" }
 observe { criteria_input "[7]" criteria_strategy "distinguish_then_check" max_candidates "2" max_rounds "1" }
 attempts "2"
}`},
		{Filename: "predicate.gooo", Content: `package observed
namespace observed
activity Negative(Integer,Box) -> Boolean computes "return input1.value == 0-input0"`},
	}
	p, err := Prepare(m)
	if err != nil {
		t.Fatal(err)
	}
	g, err := bodycodegen.NewTypedPathGenerator("")
	if err != nil {
		t.Fatal(err)
	}
	r, err := g.GenerateSourceAssembly(context.Background(), "workspace.gooo", []byte(p.Source), p.Entry.LoweredName)
	if err != nil {
		t.Fatal(err)
	}
	a := r.Report.RecordAssembly
	if a.CriteriaQuestions.Options.CriteriaStrategy != "distinguish_then_check" || a.CriteriaQuestions.Status != "NO_SURVIVING_CANDIDATE" ||
		a.Criteria.Failed != 1 || a.Criteria.Total != 2 {
		t.Fatal(a)
	}
}
