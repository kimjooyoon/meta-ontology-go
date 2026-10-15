package main

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

func TestBodyContextCLIExportsGoooCriteriaWithoutEvaluation(t *testing.T) {
	source := []byte(`package flow
namespace flow
entity Integer id "flow://integer"
entity Boolean id "flow://boolean"
activity Fee(Integer) -> Integer computes "return 11 - input" assembling {
 choice "direction" operand_order at "0" intent "Subtract eleven. 입력에서 11을 뺀다."
 criteria { predicate "difference" activity "Difference" input "[11]" holdout_input "[20]" }
 attempts "2"
}
activity Difference(Integer,Integer) -> Boolean computes "return input1 == input0 - 11"
`)
	for _, includePlan := range []bool{false, true} {
		var out, stderr bytes.Buffer
		args := []string{"--value-flow", "--activity", "Fee", "fee.gooo"}
		if includePlan {
			args = append([]string{"--include-plan"}, args...)
		}
		if runBodyContext(args, mapSourceReader{"fee.gooo": source}, &out, &stderr) != exitOK {
			t.Fatal("criteria source command", out.String(), stderr.String())
		}
		var result bodycodegen.RecordAssemblyContextExport
		if err := json.Unmarshal(out.Bytes(), &result); err != nil {
			t.Fatal(err)
		}
		flow := result.CriteriaFlow
		if result.Schema != "gooo/scalar-criteria-input-export/v1" || flow == nil || flow.Status != "RESOLVED" ||
			len(flow.Predicates) != 1 || flow.CandidatePlan == nil || result.Context.Status != "DECLINED_TO_DETERMINISTIC" ||
			result.ModelPredictions != 0 || result.CandidateTests != 0 || (result.ExpandedPlan != nil) != includePlan {
			t.Fatal("source export lost optionality or misreported model execution", out.String())
		}
		if includePlan && (result.ExpandedPlan.Criteria == nil || len(result.ExpandedPlan.Criteria.HoldoutInputs) != 1) {
			t.Fatal("explicit finite plan was lost")
		}
	}
}
