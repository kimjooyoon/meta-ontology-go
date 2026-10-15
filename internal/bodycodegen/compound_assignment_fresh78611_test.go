package bodycodegen

import (
	"context"
	"fmt"
	"strings"
	"testing"
)

const compoundBody78611 = "let value=input;value+=7;value*=3;value-=5;value/=2;return value"

func compoundSource78611(body string) []byte {
	return []byte("package compound78611\nnamespace compound78611\nentity Integer id \"urn:gooo:type:integer\"\n" +
		fmt.Sprintf("activity Adjust78611(Integer)->Integer id \"urn:compound78611:adjust\" computes %q\n", body))
}

func TestCompoundGenerationAndEvaluation78611(t *testing.T) {
	source := compoundSource78611(compoundBody78611)
	result, err := Generate("갱신78611.gooo", source, "Adjust78611")
	if err != nil || !result.Report.TypecheckPassed || !result.Report.DeterministicReplay {
		t.Fatal("source update generation", err, result.Report)
	}
	for _, operation := range []string{"+=", "*=", "-=", "/="} {
		if !strings.Contains(result.Source, operation) {
			t.Fatal("authored update token lost", operation)
		}
	}
	cases := []IRBodyFillTestCase{{Input: -9, Expected: -5}, {Input: 0, Expected: 8}, {Input: 10, Expected: 23}}
	if _, passed, err := evaluateIntegerCases([]byte(result.Source), "Adjust78611", cases); err != nil || passed != len(cases) {
		t.Fatal("four updates differ from deterministic outcomes", passed, err)
	}
	if len(result.Report.BodyLocals) != 5 || len(result.Report.BodyLocals[0].Uses) != 5 {
		t.Fatal("declaration and four prior-value reads", result.Report.BodyLocals)
	}
	t.Log("new source78611:four authored operators,three evaluator rows,five local reads;model/native0")
}

func TestCompoundTypedOrigins78611(t *testing.T) {
	source := compoundSource78611(compoundBody78611)
	report, err := InspectTypedValueOrigins("기원78611.gooo", source, "Adjust78611")
	if err != nil || report.Status != "RESOLVED" || len(report.Inputs) != 1 || len(report.Outputs) != 1 {
		t.Fatal("typed origin scope", err, report.Reason)
	}
	seen := compoundAncestors78611(report.Flow.Nodes, report.Outputs[0].Root)
	if !seen[report.Inputs[0].Node] {
		t.Fatal("original value disappeared after compound assignment")
	}
	operations := map[string]bool{}
	for index, node := range report.Flow.Nodes {
		if node.Kind != "expression" || !strings.Contains("+-*/", node.Operator) || node.Operator == "" {
			continue
		}
		if node.Parents[0] == 0 || node.Parents[1] == 0 || report.Sources[index].Status != "MAPPED" {
			t.Fatal("update lacks two parents or authored source", node, report.Sources[index])
		}
		operations[node.Operator] = true
	}
	if len(operations) != 4 {
		t.Fatal("four operators lack source-backed nodes", operations)
	}
}

func compoundAncestors78611(nodes []RecordFlowNode, root uint16) map[uint16]bool {
	seen := map[uint16]bool{}
	pending := []uint16{root}
	for len(pending) > 0 {
		id := pending[len(pending)-1]
		pending = pending[:len(pending)-1]
		if id == 0 || seen[id] || int(id) > len(nodes) {
			continue
		}
		seen[id] = true
		pending = append(pending, nodes[id-1].Parents[:]...)
	}
	return seen
}

func TestCompoundRecipeDecision78611(t *testing.T) {
	source := compoundSource78611(compoundBody78611)
	raw := []byte(`{"schema":"gooo/source-typed-path-recipe/v1","choices":[{"id":"operand78611","kind":"operand_order","occurrence":0,"intent":"기존 값에 더한다. Add to the prior value."}],"test_cases":[{"input":-9,"expected":-5},{"input":0,"expected":8},{"input":10,"expected":23}],"max_attempts":2}`)
	document, err := DecodeSourcePathDocument(context.Background(), "경로78611.gooo", source, "Adjust78611", raw)
	if err != nil {
		t.Fatal("source update cannot enter the typed decision path", err)
	}
	result, err := GenerateWithTypedPaths(context.Background(), "경로78611.gooo", source, "Adjust78611", document, "")
	if err != nil || result.Report.BodyPaths == nil || result.Report.BodyPaths.FunctionalCompleteness != 100 ||
		result.Report.BodyPaths.Search.Selection.ModelCalls != 0 {
		t.Fatal("source-owned path fallback", err, result.Report.BodyPaths)
	}
	t.Log("new recipe78611:three finite rows;source-owned update expands to prior-local binary expression;model0")
}

func TestCompoundSourceBoundaries78611(t *testing.T) {
	for _, body := range []string{
		"input+=7;return input", "missing+=7;return input",
		"let value=input;value%=3;return value", "let value=true;value+=true;return input",
		"let value=input;value/=0;return value", "let value=input;value+=true;return value",
	} {
		if _, err := Generate("境界78611.gooo", compoundSource78611(body), "Adjust78611"); err == nil {
			t.Fatal("source boundary accepted", body)
		}
	}
}
