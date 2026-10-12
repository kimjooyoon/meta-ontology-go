package main

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

func TestSourceObserveCLIUsesGoooContractWithoutJSONOptions(t *testing.T) {
	source := `package observed
namespace observed
entity Integer id "observed://integer"
activity Fee(Integer) -> Integer computes "return input - 13" assembling {
 choice "direction" operand_order at "0" intent "13에서 입력을 뺀다."
 case "13" -> "0"
 observe { input "21" oracle "Reference" max_candidates "2" max_rounds "1" resolve_unique "true" }
 attempts "2"
}
activity Reference(Integer) -> Integer computes "return 13 - input"`
	reader := mapSourceReader{"observed.gooo": []byte(source)}
	var out, diagnostics bytes.Buffer
	if code := runBodyCodegen([]string{"--json", "--activity", "Fee", "observed.gooo"}, reader, &out, &diagnostics); code != exitOK {
		t.Fatal(code, diagnostics.String(), out.String())
	}
	var result bodycodegen.Result
	if err := json.Unmarshal(out.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if result.Report.BodyPaths == nil || result.Report.BodyPaths.Observation == nil ||
		result.Report.BodyPaths.Observation.Status != "ONE_SURVIVING_CANDIDATE" ||
		!strings.Contains(result.Source, "13 - input") || result.Report.BodyPaths.LocalModelRequested {
		t.Fatal("CLI dropped the source-declared question and reference")
	}
}
