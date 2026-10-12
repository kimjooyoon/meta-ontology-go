package bodycodegen

import (
	"context"
	"reflect"
	"strings"
	"testing"

	"github.com/kimjooyoon/gooo-decision-runtime/pathplan"
)

func TestActivityRenameKeepsCriterionGraphIdentity(t *testing.T) {
	const source = `package refactorgraph
namespace refactorgraph
entity Integer id "urn:gooo:type:integer"
entity Boolean id "urn:gooo:type:boolean"
activity Fee(Integer) -> Integer computes "return 11 - input" assembling {
 choice "direction" operand_order at "0" intent "입력에서 11을 뺀다. Subtract eleven from the input."
 criteria {
  predicate "difference" activity "Difference"
  input "[11]"
  holdout_input "[20]"
 }
 attempts "2"
}
activity Difference(Integer,Integer) -> Boolean computes "let expected = Delta(input0); if input0 < 0 { return input1 == expected }; return input1 == expected"
activity Delta(Integer) -> Integer computes "return input - 11"
`
	export := func(source, activity string) RecordAssemblyContextExport {
		t.Helper()
		r, err := ExportRecordAssemblyContextWithFlow(context.Background(), "rename-graph.gooo", []byte(source), activity, false, "")
		if err != nil || r.CriteriaFlow == nil || r.CriteriaFlow.Status != "RESOLVED" || len(r.CriteriaFlow.Predicates) != 1 {
			t.Fatal("unresolved criteria graph", err, r.CriteriaFlow)
		}
		if r.ModelPredictions != 0 || r.CandidateTests != 0 {
			t.Fatal("source-only graph acquired runtime outcomes", r)
		}
		return r
	}
	original := export(source, "Fee")
	current, owner := source, "Fee"
	for _, rename := range [][2]string{{"Delta", "차액"}, {"Difference", "일치조건"}, {"Fee", "요금"}} {
		result, err := RenameActivitySource(context.Background(), "rename-graph.gooo", []byte(current), rename[0], rename[1])
		if err != nil {
			t.Fatal(rename, err)
		}
		current = result.Source
		if rename[0] == owner {
			owner = rename[1]
		}
		after := export(current, owner)
		a, b := original.CriteriaFlow.Predicates[0], after.CriteriaFlow.Predicates[0]
		if a.ActivityID != b.ActivityID || a.GraphSHA256 != b.GraphSHA256 ||
			!reflect.DeepEqual(a.Nodes, b.Nodes) || !reflect.DeepEqual(a.Inputs, b.Inputs) {
			t.Fatal("display rename changed the typed requirement", rename, a.GraphSHA256, b.GraphSHA256)
		}
		if original.OriginalSourceSHA256 == after.OriginalSourceSHA256 {
			t.Fatal("changed source lost its new identity")
		}
		if len(b.ValueFlow.Helpers) != 1 || b.ValueFlow.Helpers[0].ActivityID != "refactorgraph://activity/delta" {
			t.Fatal("helper identity was lost", b.ValueFlow.Helpers)
		}
		beforePlan, err := pathplan.Prepare(*original.CriteriaFlow.CandidatePlan)
		if err != nil {
			t.Fatal(err)
		}
		afterPlan, err := pathplan.Prepare(*after.CriteriaFlow.CandidatePlan)
		if err != nil {
			t.Fatal(err)
		}
		beforeFeatures, err := beforePlan.SourceFeatures("direction")
		if err != nil {
			t.Fatal(err)
		}
		afterFeatures, err := afterPlan.SourceFeatures("direction")
		if err != nil || beforeFeatures != afterFeatures {
			t.Fatal("display rename changed candidate structural features", rename, err)
		}
	}
	if !strings.Contains(current, `predicate "difference" activity "일치조건"`) {
		t.Fatal("requirement reference stayed on the old display name")
	}
}
