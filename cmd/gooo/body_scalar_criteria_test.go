package main

import (
	"bytes"
	"encoding/json"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestScalarCriteriaCLIReportsSourceChoicesAndScalarHoldouts(t *testing.T) {
	source := `package feecriteria
namespace feecriteria
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
 case "12" -> "1"
 holdout_case "21" -> "10"
 attempts "2"
}
activity Difference(Integer,Integer) -> Boolean computes "return input1 == input0 - 11"
`
	path := filepath.Join(t.TempDir(), "scalar.gooo")
	if err := os.WriteFile(path, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	var out, diagnostics bytes.Buffer
	if code := run([]string{"body-codegen", "--json", "--activity", "Fee", path}, &out, &diagnostics); code != exitOK {
		t.Fatal(code, out.String(), diagnostics.String())
	}
	var generated bodycodegen.Result
	if err := json.Unmarshal(out.Bytes(), &generated); err != nil {
		t.Fatal(err)
	}
	if generated.Report.OutputType != "int64" || generated.Report.RecordAssembly.Criteria.Passed != 2 {
		t.Fatal(generated.Report)
	}
	c := bodyexecution.JointConstruction{Initial: bodyexecution.Composition{Steps: []bodyexecution.CompositionStep{{Generation: generated}}}}
	rows := constructionRecordHoldoutRows(c)
	if len(rows) != 2 || rows[0][0] != "Initial scalar body holdout" || !strings.Contains(rows[0][1], "Integer cases") || strings.Contains(rows[0][1], "fields") {
		t.Fatal(rows)
	}
	out.Reset()
	diagnostics.Reset()
	if code := run([]string{"body-plan", "--source", path, "--entry", "Fee"}, &out, &diagnostics); code != exitOK || !strings.Contains(out.String(), "scalar_criteria_choices") {
		t.Fatal(code, out.String(), diagnostics.String())
	}
}
