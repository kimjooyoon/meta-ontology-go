package workspaceexecution

import (
	"context"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime"
)

func TestCriteriaQuestionsWorkspaceExecutesSourcePredicates(t *testing.T) {
	m := observedWorkspace()
	m.Packages[0].Sources = []packageruntime.Source{
		{Filename: "records.gooo", Content: `package observed
namespace observed
entity Integer id "workspace-question://integer"
entity Boolean id "workspace-question://boolean"
entity Box id "workspace-question://box" fields { field value id "workspace-question://value" type integer required one }
activity Fee(Integer) -> Box computes "return Box{value:input}" assembling {
 choice "value" field_value at "0" alternative "0-input" intent "Negate."
 criteria { predicate "negative" activity "Negative" input "[0]" holdout_input "[9]" }
 observe { criteria_input "[9007199254740993]" max_candidates "2" max_rounds "1" resolve_unique "true" }
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
	generated, err := g.GenerateSourceAssembly(context.Background(), "workspace.gooo", []byte(p.Source), p.Entry.LoweredName)
	if err != nil {
		t.Fatal(err)
	}
	r := generated.Report.RecordAssembly
	if r.CriteriaQuestions.EffectiveInputs != 2 || r.CriteriaQuestions.AcquiredInputs[0] != "[9007199254740993]" ||
		r.Criteria.Passed != 2 || r.Criteria.Inputs[1].ActualJSON != `{"value":-9007199254740993}` || r.SelectedMask != 1 {
		t.Fatal("workspace lost typed criteria question", r)
	}
}
