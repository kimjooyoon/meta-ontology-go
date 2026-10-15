package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/kimjooyoon/gooo-decision-runtime/contractdecision"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

const v3ContractSource = `package v3route
namespace v3route
entity Integer id "v3-regression://integer"
activity Choose(Integer) -> Integer id "v3-regression://choose" computes ` + "`" + `let load = input
let result = load
if 67 < load { result = load + 13 } else { result = load / 4 }
return result` + "`" + ` assembling {
 choice "comparison" operand_order at "0" intent "입력이67 미만인지 확인한다. Check whether input is below67."
 choice "branches" branch_layout at "0" intent "67 미만이면4로 나눈 정수 몫, 나머지는13을 더한다."
 case "-17" -> "-4"
 case "-1" -> "0"
 case "66" -> "16"
 case "67" -> "80"
 case "68" -> "81"
 case "99" -> "112"
 condition_case "comparison" input "-1" -> "true"
 condition_case "comparison" input "66" -> "true"
 condition_case "comparison" input "67" -> "false"
 attempts "4"
}
activity Main(Integer) -> Integer id "v3-regression://main" computes "return Choose(input)"
`

func freshV3Fixture(t *testing.T) (string, string) {
	t.Helper()
	root := t.TempDir()
	source, model := filepath.Join(root, "source.gooo"), filepath.Join(root, "zero-v3.json")
	m, err := contractdecision.NewSourceRelativeV3([contractdecision.InteractionRequirementParameterCount]float32{})
	if err != nil {
		t.Fatal(err)
	}
	raw, err := m.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(source, []byte(v3ContractSource), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(model, raw, 0600); err != nil {
		t.Fatal(err)
	}
	return source, model
}

func TestSourceRelativeV3CLIInspectsThenCallsOneModelBeforeFiniteChecks(t *testing.T) {
	source, model := freshV3Fixture(t)
	var out, diagnostics bytes.Buffer
	if code := run([]string{"body-context", "--model", model, "--activity", "Choose", source}, &out, &diagnostics); code != exitOK {
		t.Fatal(code, diagnostics.String())
	}
	var before bodycodegen.TypedPathContextExport
	if err := json.Unmarshal(out.Bytes(), &before); err != nil {
		t.Fatal(err)
	}
	if before.Context == nil || before.Context.Schema != "gooo/compiler-source-relative-contract-model-context/v3" ||
		before.Context.FeatureVersion != contractdecision.SourceRelativeV3FeatureVersion || before.ContractCases == nil ||
		before.ContractCases.Count != 6 || before.ModelPredictions != 0 || before.CandidateTests != 0 || before.SelectedEmission {
		t.Fatal("V3 source binding or zero-call inspection differs")
	}
	for _, row := range before.Inputs {
		if row.SourceRelativeFeatures == nil || row.SourceRelativeOutputs == nil || !row.SourceRelativeOutputs.Available ||
			row.SourceRelativeOutputs.CaseCount != 6 || row.SourceRelativeOutputs.CaseSHA256 == "" {
			t.Fatal("V3 output relationships not source/requirement bound")
		}
	}
	out.Reset()
	diagnostics.Reset()
	if code := run([]string{"body-codegen", "--json", "--activity", "Choose", "--path-model", model, "--path-step-attempts", "1", source}, &out, &diagnostics); code != exitOK {
		t.Fatal(code, diagnostics.String())
	}
	var result bodycodegen.Result
	if err := json.Unmarshal(out.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	p := result.Report.BodyPaths
	if p == nil || p.ContractRanking == nil || p.ContractRanking.Calls != 1 || p.Conditions.Passed != 3 ||
		p.FunctionalCompleteness != 100 || p.ContractRanking.Schema != "gooo/source-relative-requirement-path-ranking/v3" {
		t.Fatal("V3 source-owned finite search did not finish after one prediction")
	}
}
