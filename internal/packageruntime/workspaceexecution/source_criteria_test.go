package workspaceexecution

import (
	"context"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime"
)

func TestCriteriaWorkspaceRetainsPredicateFromSeparateSource(t *testing.T) {
	m := observedWorkspace()
	m.Packages[0].Sources = []packageruntime.Source{
		{Filename: "records.gooo", Content: `package observed
namespace observed
entity Integer id "workspace-criteria://integer"
entity Boolean id "workspace-criteria://boolean"
entity Box id "workspace-criteria://box" fields { field value id "workspace-criteria://value" type integer required one }
activity Fee(Integer) -> Box computes "return Box{value:input}" assembling {
 choice "value" field_value at "0" alternative "0-input" intent "Negate."
 criteria { predicate "negative" activity "Negative" input "[9007199254740993]" holdout_input "[9]" }
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
	if p.PureCalls == nil || len(p.PureCalls.Sites) != 1 || p.PureCalls.Sites[0].Surface != "criteria:negative" ||
		len(p.PureCalls.Activities) != 2 {
		t.Fatal("workspace lost criterion dependency", p.PureCalls)
	}
	spec, err := bodycodegen.SourceAssembly(context.Background(), "workspace.gooo", []byte(p.Source), p.Entry.LoweredName)
	if err != nil || spec.Criteria == nil || spec.Criteria.Predicates[0].Activity == "Negative" ||
		spec.Criteria.Inputs[0] != "[9007199254740993]" {
		t.Fatal("criterion was not retained with its package", err, spec)
	}
	g, err := bodycodegen.NewTypedPathGenerator("")
	if err != nil {
		t.Fatal(err)
	}
	generated, err := g.GenerateSourceAssembly(context.Background(), "workspace.gooo", []byte(p.Source), p.Entry.LoweredName)
	if err != nil {
		t.Fatal(err)
	}
	if generated.Report.RecordAssembly.Criteria.Passed != 1 || generated.Report.RecordAssembly.SelectedMask != 1 ||
		generated.Report.RecordAssembly.Criteria.Inputs[0].ActualJSON != `{"value":-9007199254740993}` {
		t.Fatal("workspace criterion did not reach selection", generated.Report.RecordAssembly)
	}
	for _, name := range []string{"Missing", "Fee"} {
		invalid := m
		invalid.Packages = append([]packageruntime.PackageSpec(nil), m.Packages...)
		invalid.Packages[0].Sources = append([]packageruntime.Source(nil), m.Packages[0].Sources...)
		invalid.Packages[0].Sources[0].Content = strings.Replace(m.Packages[0].Sources[0].Content,
			`activity "Negative"`, `activity "`+name+`"`, 1)
		if _, err := Prepare(invalid); err == nil {
			t.Fatal("invalid criterion accepted", name)
		}
	}
}
