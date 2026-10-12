package workspaceexecution

import (
	"context"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime"
	"testing"
)

func TestRecordObserveWorkspaceRetainsAndRenamesOracle(t *testing.T) {
	m := observedWorkspace()
	m.Packages[0].Sources = []packageruntime.Source{
		{Filename: "records.gooo", Content: `package observed
namespace observed
entity Integer id "workspace-record://integer"
entity Box id "workspace-record://box" fields { field value id "workspace-record://value" type integer required one }
activity Fee(Integer) -> Box computes "return Box{value: input}" assembling {
 choice "value" field_value at "0" alternative "0 - input" intent "Negate."
 value_case "[0]" -> "{\"value\":0}"
 observe { value_input "[9007199254740993]" oracle "Reference" max_candidates "2" max_rounds "1" resolve_unique "true" }
 attempts "2"
}`},
		{Filename: "reference.gooo", Content: `package observed
namespace observed
activity Reference(Integer) -> Box computes "return Box{value: 0 - input}"`},
	}
	p, err := Prepare(m)
	if err != nil {
		t.Fatal(err)
	}
	source := []byte(p.Source)
	ctx := context.Background()
	spec, err := bodycodegen.SourceAssembly(ctx, "workspace.gooo", source, p.Entry.LoweredName)
	if err != nil || spec.Observation == nil || spec.Observation.OracleActivity == "Reference" || spec.Observation.ValueInputs[0] != "[9007199254740993]" {
		t.Fatal(err, spec)
	}
	g, _ := bodycodegen.NewTypedPathGenerator("")
	result, err := g.GenerateSourceAssembly(ctx, "workspace.gooo", source, p.Entry.LoweredName)
	if err != nil {
		t.Fatal(err)
	}
	r := result.Report.RecordAssembly
	if r == nil || r.Passed != 2 || r.Total != 2 || r.ModelCalls != 0 || r.Observation.ResolvedMask == nil {
		t.Fatal(r)
	}
	if _, err = bodycodegen.RealizeSourceAssembly(ctx, "workspace.gooo", source, result); err != nil {
		t.Fatal(err)
	}
}
