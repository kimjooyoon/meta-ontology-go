package bodycodegen

import (
	"context"
	"os"
	"reflect"
	"strings"
	"sync"
	"testing"

	decision "github.com/kimjooyoon/gooo-decision-runtime"
	"github.com/kimjooyoon/gooo-decision-runtime/flowdecision"
	"github.com/kimjooyoon/gooo-decision-runtime/pathplan"
)

// Zero weights test versioned wiring; the published fresh weights are measured
// separately. These tests do not establish a learned accuracy result.
func semanticContractModel(t *testing.T, version string) string {
	t.Helper()
	m, err := flowdecision.NewForFeatures([flowdecision.ParameterCount]float32{}, version)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := m.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	return writeCandidateArtifact(t, raw)
}

func semanticSource(t *testing.T, body string) ([]byte, pathplan.Document) {
	t.Helper()
	source := []byte(strings.Replace(string(executionModelSource()),
		"if input < 10 { return input } else { return 10 }", body, 1))
	doc, err := DecodeSourcePathDocument(context.Background(), "semantic.gooo", source, "Choose", nil)
	if err != nil {
		t.Fatal(err)
	}
	return source, doc
}

func TestSemanticFlowCompilerEquivalentFormsAndRetainedConcurrency(t *testing.T) {
	name := semanticContractModel(t, decision.SemanticFlowFeatureVersion)
	forms := []string{
		"if input < 10 { return input } else { return 10 }",
		"let value = input; let limit = 10; if value < limit { return value } else { return limit }",
		"let result = input; if input < 10 { result = input } else { result = 10 }; return result",
		"let value = input; let copy = value; value = 0; if copy < 10 { return copy } else { return 10 }",
		"let result = input; if input < 10 { result = input } else { result = 10 }; let copy = result; result = 0; return copy",
	}
	var first [][384]float32
	for i, body := range forms {
		source, doc := semanticSource(t, body)
		before, err := ExportTypedPathModelContext(context.Background(), "semantic.gooo", source, "Choose", doc, name, decision.SemanticFlowFeatureVersion)
		if err != nil {
			t.Fatal(err)
		}
		var inputs [][384]float32
		for _, input := range before.Inputs {
			if input.SemanticFlow == nil || !input.SemanticFlow.Normalized || input.FlowFeatures == nil ||
				input.SourceFeatureSHA != digest(input.SemanticFlow.Source[:]) {
				t.Fatal("source normalization absent")
			}
			inputs = append(inputs, *input.FlowFeatures)
		}
		if i == 0 {
			first = inputs
		} else if !reflect.DeepEqual(first, inputs) {
			t.Fatal("source form changes decision input", i)
		}
		g, err := NewTypedPathGenerator(name)
		if err != nil {
			t.Fatal(err)
		}
		if i == len(forms)-1 {
			if err := os.Remove(name); err != nil {
				t.Fatal(err)
			}
			var wg sync.WaitGroup
			for range 3 {
				wg.Go(func() { checkSemanticGeneration(t, g, source, doc, before) })
			}
			wg.Wait()
		} else {
			checkSemanticGeneration(t, g, source, doc, before)
		}
	}
}

func checkSemanticGeneration(t *testing.T, g *TypedPathGenerator, source []byte, doc pathplan.Document, before TypedPathContextExport) {
	t.Helper()
	if before.ModelPredictions != 0 || before.CandidateTests != 0 {
		t.Error("preflight performed inference")
	}
	result, err := g.Generate(context.Background(), "semantic.gooo", source, "Choose", doc, TypedPathOptions{StepAttempts: 1, FeedbackRounds: 3})
	if err != nil {
		t.Error(err)
		return
	}
	p := result.Report.BodyPaths
	if p.FunctionalCompleteness != 100 || p.Conditions.Passed != 3 || !reflect.DeepEqual(p.ModelContext, before.Context) {
		t.Error("versioned assembly and preflight differ")
	}
	for i, input := range before.Inputs {
		if input.InputSHA != "sha256:"+p.ConditionProgress[0].Ranking.FeatureSHA[i] {
			t.Error("actual v5 array differs", i)
		}
	}
	for _, f := range p.ConditionFeedback {
		if f.Calls > 0 && f.FeatureVersion != decision.SemanticFlowFeatureVersion {
			t.Error("feedback changed feature version")
		}
	}
	if err := VerifyTypedPathProjection(context.Background(), "semantic.gooo", source, doc, result); err != nil {
		t.Error(err)
	}
}

func TestSemanticFlowCompilerKeepsUnsupportedShapeAndRejectsExplicitMismatch(t *testing.T) {
	source, doc := semanticSource(t, "if input < 10 { if input < 0 { return input } else { return 0 } } else { return 10 }")
	name := semanticContractModel(t, decision.SemanticFlowFeatureVersion)
	if _, err := ExportTypedPathModelContext(context.Background(), "semantic.gooo", source, "Choose", doc, name, decision.ExecutionFlowFeatureVersion); err == nil {
		t.Fatal("explicit feature mismatch accepted")
	}
	newInput, err := ExportTypedPathModelContext(context.Background(), "semantic.gooo", source, "Choose", doc, name, decision.SemanticFlowFeatureVersion)
	if err != nil {
		t.Fatal(err)
	}
	oldInput, err := ExportTypedPathModelContext(context.Background(), "semantic.gooo", source, "Choose", doc, semanticContractModel(t, decision.ExecutionFlowFeatureVersion), decision.ExecutionFlowFeatureVersion)
	if err != nil {
		t.Fatal(err)
	}
	for i, input := range newInput.Inputs {
		if input.SemanticFlow == nil || input.SemanticFlow.Normalized || *input.FlowFeatures != *oldInput.Inputs[i].FlowFeatures || oldInput.Inputs[i].SemanticFlow != nil {
			t.Fatal("unsupported shape or old ABI changed")
		}
	}
}
