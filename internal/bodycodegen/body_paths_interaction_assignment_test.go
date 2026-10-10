package bodycodegen

import (
	"context"
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"

	decision "github.com/kimjooyoon/gooo-decision-runtime"
	"github.com/kimjooyoon/gooo-decision-runtime/contractdecision"
)

func interactionAssignmentSource(t *testing.T) []byte {
	t.Helper()
	source, err := os.ReadFile("../../examples/body-codegen/source-interaction-assignment-cases.gooo.fixture")
	if err != nil {
		t.Fatal(err)
	}
	return source
}

func TestInteractionAssignmentContextExplainsActualArray(t *testing.T) {
	source := interactionAssignmentSource(t)
	doc := declaredContractDocument(t, source)
	name := writeInteractionContractModel(t, contractdecision.MeanPooling)
	before, err := ExportTypedPathModelContext(context.Background(), "assignment.gooo", source, "Choose", doc, name, "")
	if err != nil || len(before.Inputs) != 2 {
		t.Fatal("assignment context unavailable", err)
	}
	want := [3]decision.OrderedExpression{
		{Operation: "less_than", Operands: [2]decision.FlowValue{{Present: true, Kind: "input"}, {Present: true, Kind: "int"}}},
		{Operation: "value", Operands: [2]decision.FlowValue{{Present: true, Kind: "input"}}},
		{Operation: "subtract", Operands: [2]decision.FlowValue{{Present: true, Kind: "int"}, {Present: true, Kind: "input"}}},
	}
	for _, row := range before.Inputs {
		if row.OrderedBranch == nil || !row.OrderedBranch.Available || row.OrderedBranch.Expressions != want {
			t.Fatal("reaching writes, branch assignment or final return lost", row.OrderedBranch)
		}
		var encoded [decision.OrderedExpressionFeatureDim]float32
		if err := decision.OrderedExpressionFeaturesInto(row.OrderedBranch.Expressions, &encoded); err != nil ||
			!reflect.DeepEqual(encoded[:], row.OrderedFeatures[contractdecision.FeatureDim:]) {
			t.Fatal("explanation differs from actual model input", err)
		}
	}
	checkInteractionContractExport(t, before, doc)
	g, err := NewTypedPathGenerator(name)
	if err != nil {
		t.Fatal(err)
	}
	checkInteractionContractGeneration(t, g, source, doc, before)
}

func TestInteractionAssignmentContextKeepsExactLargeLiteral(t *testing.T) {
	source := []byte(strings.Replace(string(interactionAssignmentSource(t)), "limit = 0", "limit = 9007199254740993", 1))
	doc := declaredContractDocument(t, source)
	name := writeInteractionContractModel(t, contractdecision.ExtremePooling)
	before, err := ExportTypedPathModelContext(context.Background(), "large.gooo", source, "Choose", doc, name, "")
	if err != nil || len(before.Inputs) != 2 || before.ModelPredictions != 0 || before.CandidateTests != 0 {
		t.Fatal("source inspection performed work or declined", err)
	}
	for _, row := range before.Inputs {
		if row.OrderedBranch == nil || row.OrderedBranch.Expressions[0].Operands[1].Int != 9007199254740993 {
			t.Fatal("large integer changed in explanation")
		}
		encoded, err := json.Marshal(row.OrderedBranch)
		if err != nil || !strings.Contains(string(encoded), "9007199254740993") {
			t.Fatal("large integer changed in JSON", err)
		}
	}
}
