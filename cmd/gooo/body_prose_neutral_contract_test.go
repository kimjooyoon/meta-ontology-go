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

// A new synthetic contract and zero weights test compiler binding, not the
// frozen trained model or any historical performance/native study.
const proseNeutralContractSource = `package neutralrate
namespace neutralrate
entity Integer id "neutral-rate-regression://integer"
activity Charge(Integer) -> Integer id "neutral-rate-regression://charge" computes ` + "`" + `let usage = input
let result = usage
if 137 < usage { result = usage * 5 } else { result = usage * 7 }
return result` + "`" + ` assembling {
 choice "comparison" operand_order at "0" intent "사용량이137 미만인지 확인한다. Check the usage threshold."
 choice "branches" branch_layout at "0" intent "137 미만이면 단가7, 그 외에는 단가5로 계산한다."
 case "-31" -> "-217"
 case "0" -> "0"
 case "136" -> "952"
 case "137" -> "685"
 case "138" -> "690"
 case "251" -> "1255"
 condition_case "comparison" input "-1" -> "true"
 condition_case "comparison" input "136" -> "true"
 condition_case "comparison" input "137" -> "false"
 attempts "4"
}
activity Main(Integer) -> Integer id "neutral-rate-regression://main" computes "return Charge(input)"
`

func TestProseNeutralCompilerContextAndOneCallFiniteConstruction(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "neutral.gooo")
	if err := os.WriteFile(source, []byte(proseNeutralContractSource), 0600); err != nil {
		t.Fatal(err)
	}
	full, err := contractdecision.NewSourceRelativeV3([contractdecision.InteractionRequirementParameterCount]float32{})
	if err != nil {
		t.Fatal(err)
	}
	raw, err := full.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	neutral, err := contractdecision.RepackageSourceRelativeV3ProseNeutral(raw)
	if err != nil {
		t.Fatal(err)
	}
	paths := [2]string{filepath.Join(root, "zero-full-v3.json"), filepath.Join(root, "zero-neutral-v3.json")}
	var contexts [2]bodycodegen.TypedPathContextExport
	for i, artifact := range [][]byte{raw, neutral} {
		if err := os.WriteFile(paths[i], artifact, 0600); err != nil {
			t.Fatal(err)
		}
		var out, diagnostics bytes.Buffer
		if code := run([]string{"body-context", "--model", paths[i], "--activity", "Charge", source}, &out, &diagnostics); code != exitOK {
			t.Fatal(code, diagnostics.String())
		}
		if err := json.Unmarshal(out.Bytes(), &contexts[i]); err != nil {
			t.Fatal(err)
		}
		if contexts[i].Context == nil || contexts[i].ContractCases == nil || contexts[i].ContractConditions == nil ||
			contexts[i].ModelPredictions != 0 || contexts[i].CandidateTests != 0 || contexts[i].SelectedEmission ||
			contexts[i].ContractCases.Count != 6 || contexts[i].ContractConditions.Count != 3 {
			t.Fatal("context ran work or lost original requirements")
		}
	}
	a, b := contexts[0], contexts[1]
	if b.Context.Schema != "gooo/compiler-source-relative-v3-prose-neutral-context/v1" || b.Context.FeatureVersion != contractdecision.SourceRelativeV3ProseNeutralFeatureVersion || b.Context.IntentPolicy != contractdecision.SourceRelativeV3ProseNeutralIntentPolicy || b.Context.ClearedIntentCells == nil || *b.Context.ClearedIntentCells != ([2]int{64, 192}) {
		t.Fatal("explicit input policy missing")
	}
	if a.Context.IntentPolicy != "" || a.Context.ClearedIntentCells != nil || len(a.Inputs) != len(b.Inputs) {
		t.Fatal("ordinary V3 context changed")
	}
	for i, row := range b.Inputs {
		old := a.Inputs[i]
		if row.Text != old.Text || row.Text == "" || row.OriginalIntentSHA != old.OriginalIntentSHA ||
			row.SourceRelativeFeatures == nil || old.SourceRelativeFeatures == nil || row.SourceRelativeOutputs == nil ||
			row.SourceRelativeOutputs.CaseCount != 6 {
			t.Fatal("original source/output view lost")
		}
		for slot, value := range *row.SourceRelativeFeatures {
			if slot >= 64 && slot < 192 {
				if value != 0 {
					t.Fatal("intent channel remained populated")
				}
			} else if value != old.SourceRelativeFeatures[slot] {
				t.Fatal("other source channel changed", slot)
			}
		}
	}
	var out, diagnostics bytes.Buffer
	if code := run([]string{"body-codegen", "--json", "--activity", "Charge", "--path-model", paths[1], "--path-step-attempts", "1", source}, &out, &diagnostics); code != exitOK {
		t.Fatal(code, diagnostics.String())
	}
	var result bodycodegen.Result
	if err := json.Unmarshal(out.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	p := result.Report.BodyPaths
	if p == nil || p.ContractRanking == nil || p.ContractRanking.Calls != 1 || p.FunctionalCompleteness != 100 || p.Conditions.Passed != 3 || p.ContractRanking.Schema != "gooo/source-relative-v3-prose-neutral-path-ranking/v1" {
		t.Fatal("source-owned finite construction after one prediction")
	}
	for i, row := range b.Inputs {
		if row.InputSHA != "sha256:"+p.ContractRanking.FeatureSHA[i] {
			t.Fatal("context digest differs from actual model input", i, row.InputSHA)
		}
	}
	if p.ModelContext == nil || p.ModelContext.IntentPolicy != b.Context.IntentPolicy {
		t.Fatal("construction receipt lost input policy")
	}
	for _, args := range [][]string{{"body-codegen", "--json", "--activity", "Charge", source}, {"body-context", "--model", paths[1], "--feature-version", contractdecision.SourceRelativeV3FeatureVersion, "--activity", "Charge", source}} {
		out.Reset()
		diagnostics.Reset()
		code := run(args, &out, &diagnostics)
		if len(args) > 6 {
			if code == exitOK {
				t.Fatal("explicit feature ABI mismatch accepted")
			}
			continue
		}
		if code != exitOK {
			t.Fatal("deterministic omission", code, diagnostics.String())
		}
		result = bodycodegen.Result{}
		if err := json.Unmarshal(out.Bytes(), &result); err != nil {
			t.Fatal(err)
		}
		if result.Report.BodyPaths.ContractRanking != nil || result.Report.BodyPaths.FunctionalCompleteness != 100 {
			t.Fatal("deterministic path changed")
		}
	}
}
