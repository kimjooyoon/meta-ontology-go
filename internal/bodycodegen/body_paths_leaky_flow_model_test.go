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

// Controlled negative weights isolate computation routing, not learned accuracy.
func leakyContractModel(t *testing.T) string {
	t.Helper()
	var weights [flowdecision.ParameterCount]float32
	bias := flowdecision.FeatureDim * flowdecision.HiddenDim
	weights[bias] = -1
	weights[bias+flowdecision.HiddenDim] = .5
	weights[bias+2*flowdecision.HiddenDim] = -.5
	m, err := flowdecision.NewForActivation(weights, decision.RelationalFlowFeatureVersion, flowdecision.LeakyReLUActivation)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := m.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	return writeCandidateArtifact(t, raw)
}

func TestLeakyCompilerPreservesActualRelationalInputAndComputation(t *testing.T) {
	name := leakyContractModel(t)
	forms := []string{
		"if input < 10 { return input } else { return 10 }",
		"let result = input; if input < 10 { result = input } else { result = 10 }; let copy = result; result = 0; return copy",
	}
	var first [][384]float32
	for i, body := range forms {
		source, doc := semanticSource(t, body)
		before, err := ExportTypedPathModelContext(context.Background(), "leaky.gooo", source, "Choose", doc, name, decision.RelationalFlowFeatureVersion)
		if err != nil {
			t.Fatal(err)
		}
		var arrays [][384]float32
		for _, input := range before.Inputs {
			if input.SemanticFlow == nil || !input.SemanticFlow.Normalized || input.FlowFeatures == nil || input.FlowFeatures[255] != 3 {
				t.Fatal("relational source projection absent")
			}
			arrays = append(arrays, *input.FlowFeatures)
		}
		if i == 0 {
			first = arrays
		} else if !reflect.DeepEqual(first, arrays) {
			t.Fatal("equivalent source forms changed relational input")
		}
		g, err := NewTypedPathGenerator(name)
		if err != nil {
			t.Fatal(err)
		}
		if i == len(forms)-1 {
			if err := os.Remove(name); err != nil {
				t.Fatal(err)
			}
		}
		var wg sync.WaitGroup
		for range 3 {
			wg.Go(func() { checkLeakyCompilerGeneration(t, g, source, doc, before) })
		}
		wg.Wait()
	}
}

func checkLeakyCompilerGeneration(t *testing.T, g *TypedPathGenerator, source []byte, doc pathplan.Document, before TypedPathContextExport) {
	t.Helper()
	if g.Info().ModelSchema != flowdecision.ActivationSchema || before.ModelCompatibility.Model.ModelSchema != flowdecision.ActivationSchema || before.ModelPredictions != 0 || before.CandidateTests != 0 {
		t.Error("computation identity or preflight changed")
	}
	result, err := g.Generate(context.Background(), "leaky.gooo", source, "Choose", doc, TypedPathOptions{StepAttempts: 1, FeedbackRounds: 4})
	if err != nil {
		t.Error(err)
		return
	}
	p := result.Report.BodyPaths
	if p.FunctionalCompleteness != 100 || p.Conditions.Passed != 3 || p.ConditionProgress[0].Ranking.Proposed != 3 || !reflect.DeepEqual(before.Context, p.ModelContext) {
		t.Error("signed computation did not reach checked assembly")
	}
	for i, input := range before.Inputs {
		if input.InputSHA != "sha256:"+p.ConditionProgress[0].Ranking.FeatureSHA[i] {
			t.Error("actual model array differs from preflight")
		}
	}
	for _, feedback := range p.ConditionFeedback {
		if feedback.Calls > 0 && feedback.FeatureVersion != decision.RelationalFlowFeatureVersion {
			t.Error("feedback changed the input contract")
		}
	}
	if err := VerifyTypedPathProjection(context.Background(), "leaky.gooo", source, doc, result); err != nil {
		t.Error(err)
	}
}

func TestLeakyCompilerRejectsMismatchedComputationArtifact(t *testing.T) {
	name := leakyContractModel(t)
	raw, err := os.ReadFile(name)
	if err != nil {
		t.Fatal(err)
	}
	for _, bad := range []string{
		strings.Replace(string(raw), flowdecision.ActivationSchema, flowdecision.Schema, 1),
		strings.Replace(string(raw), flowdecision.LeakyReLUActivation, "unknown", 1),
		strings.Replace(string(raw), `,"activation":"`+flowdecision.LeakyReLUActivation+`"`, "", 1),
	} {
		if _, err := NewTypedPathGenerator(writeCandidateArtifact(t, []byte(bad))); err == nil {
			t.Fatal("invalid computation artifact accepted")
		}
	}
}

func TestRelationalCompilerKeepsUnsupportedShapeAndRejectsFeatureMismatch(t *testing.T) {
	source, doc := semanticSource(t, "if input < 10 { if input < 0 { return input } else { return 0 } } else { return 10 }")
	name := leakyContractModel(t)
	if _, err := ExportTypedPathModelContext(context.Background(), "leaky.gooo", source, "Choose", doc, name, decision.SemanticFlowFeatureVersion); err == nil {
		t.Fatal("explicit feature mismatch accepted")
	}
	v6, err := ExportTypedPathModelContext(context.Background(), "leaky.gooo", source, "Choose", doc, name, decision.RelationalFlowFeatureVersion)
	if err != nil {
		t.Fatal(err)
	}
	v5, err := ExportTypedPathModelContext(context.Background(), "leaky.gooo", source, "Choose", doc, semanticContractModel(t, decision.SemanticFlowFeatureVersion), decision.SemanticFlowFeatureVersion)
	if err != nil {
		t.Fatal(err)
	}
	for i, input := range v6.Inputs {
		if input.SemanticFlow == nil || input.SemanticFlow.Normalized || *input.FlowFeatures != *v5.Inputs[i].FlowFeatures {
			t.Fatal("unsupported shape input changed")
		}
	}
}
