package bodycodegen

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kimjooyoon/gooo-decision-runtime/criteriadecision"
)

const scalarModelBridgeSource = `package sourcedecision
namespace sourcedecision
entity Integer id "urn:sourcedecision:integer"
entity Boolean id "urn:sourcedecision:boolean"
activity Offset(Integer) -> Integer id "urn:sourcedecision:offset" computes "return input - 37" assembling {
 choice "direction" operand_order at "0" intent "37에서 입력을 뺀다. Subtract the input from thirty-seven."
 criteria {
  predicate "sum" activity "SumIs37"
  input "[5]"
  input "[12]"
  holdout_input "[41]"
 }
 attempts "1"
}
activity SumIs37(Integer,Integer) -> Boolean computes "return input0 + input1 == 37"
`

func scalarBridgeSyntheticArtifact(t *testing.T) string {
	t.Helper()
	var parameters [criteriadecision.ParameterCount]float32
	// This architecture's second unary bias selects option 1. No fitting occurs.
	parameters[criteriadecision.ParameterCount-criteriadecision.HiddenDim-1] = 3
	metadata, weights, err := criteriadecision.EncodeArtifact(&parameters)
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	name := filepath.Join(dir, "metadata.json")
	for path, raw := range map[string][]byte{name: metadata, filepath.Join(dir, "weights.bin"): weights} {
		if err := os.WriteFile(path, raw, 0600); err != nil {
			t.Fatal(err)
		}
	}
	return name
}

func TestScalarSourceModelRanksActualGoooCriteria(t *testing.T) {
	name := scalarBridgeSyntheticArtifact(t)
	g, err := NewTypedPathGenerator(name)
	if err != nil {
		t.Fatal(err)
	}
	source := []byte(scalarModelBridgeSource)
	result, err := g.GenerateSourceAssembly(context.Background(), "source-model.gooo", source, "Offset")
	if err != nil {
		t.Fatal(err)
	}
	r := result.Report.RecordAssembly
	if r.ModelCalls != 1 || r.ScalarPrediction == nil || r.SelectedMask != 1 || len(r.Attempts) != 1 ||
		r.Criteria.Passed != 2 || r.Criteria.Total != 2 || r.CriteriaHoldout.Passed != 1 ||
		r.ScalarPrediction.CandidateCount != 2 || r.Context.Status != "ENCODED" {
		t.Fatalf("source ranking and actual obligations differ: %+v", r)
	}
	if strings.Contains(r.Context.Text, "[41]") || strings.Contains(r.Context.Text, "holdout") ||
		strings.Contains(r.Context.Text, "expected") || !strings.Contains(result.GoooSource, `id "urn:sourcedecision:offset"`) {
		t.Fatal("source input leaked finite cases or identity changed")
	}
	verifyScalarBridgeFreshReplay(t, name, source, result)
}

func verifyScalarBridgeFreshReplay(t *testing.T, name string, source []byte, result Result) {
	t.Helper()
	for _, path := range []string{name, filepath.Join(filepath.Dir(name), "weights.bin")} {
		if err := os.Remove(path); err != nil {
			t.Fatal(err)
		}
	}
	raw, err := json.Marshal(result)
	if err != nil {
		t.Fatal(err)
	}
	var decoded Result
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatal(err)
	}
	if _, err := RealizeSourceAssembly(context.Background(), "source-model.gooo", source, decoded); err != nil {
		t.Fatal("fresh receipt required a model reload or lost source bindings", err)
	}
	decoded.Report.RecordAssembly.ScalarPrediction.FeatureSHA256 = "sha256:" + strings.Repeat("0", 64)
	if _, err := RealizeSourceAssembly(context.Background(), "source-model.gooo", source, decoded); err == nil {
		t.Fatal("changed numerical feature identity accepted")
	}
}

func TestScalarSourceModelDisconnectedKeepsFiniteSearch(t *testing.T) {
	source := []byte(`package disconnectedsource
namespace disconnectedsource
entity Integer id "urn:disconnectedsource:integer"
entity Boolean id "urn:disconnectedsource:boolean"
activity Delta(Integer) -> Integer computes "return 5 - input" assembling {
 choice "direction" operand_order at "0" intent "입력에서 5를 빼기."
 criteria { predicate "delta" activity "Difference" input "[9]" input "[-3]" }
 attempts "2"
}
activity Difference(Integer,Integer) -> Boolean computes "return input1 == input0 - 5"
`)
	g, err := NewTypedPathGenerator("")
	if err != nil {
		t.Fatal(err)
	}
	result, err := g.GenerateSourceAssembly(context.Background(), "disconnected-model.gooo", source, "Delta")
	if err != nil {
		t.Fatal(err)
	}
	r := result.Report.RecordAssembly
	if r.ModelCalls != 0 || r.Model != nil || r.ScalarPrediction != nil || r.SelectedMask != 1 ||
		len(r.Attempts) != 2 || r.Criteria.Passed != 2 || r.Criteria.Total != 2 {
		t.Fatalf("disconnected finite search changed: %+v", r)
	}
}
