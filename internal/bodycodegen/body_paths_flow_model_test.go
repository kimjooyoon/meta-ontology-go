package bodycodegen

import (
	"context"
	"os"
	"reflect"
	"slices"
	"strings"
	"sync"
	"testing"

	decision "github.com/kimjooyoon/gooo-decision-runtime"
	"github.com/kimjooyoon/gooo-decision-runtime/flowdecision"
)

// Controlled weights prove wiring; measured learned accuracy belongs to the
// published study, whose weights remain unchanged.
func writeFlowContractModel(t *testing.T, feedback bool) string {
	t.Helper()
	var weights [flowdecision.ParameterCount]float32
	if feedback {
		weights[256], weights[311] = 8, 128
		weights[flowdecision.FeatureDim*flowdecision.HiddenDim] = -1
	} else {
		weights[321] = 8
	}
	weights[flowdecision.FeatureDim*flowdecision.HiddenDim+2*flowdecision.HiddenDim] = 2
	m, err := flowdecision.New(weights)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := m.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	return writeCandidateArtifact(t, raw)
}

func TestFlowModelCompilerAssignmentAndActualInput(t *testing.T) {
	ctx := context.Background()
	source := []byte(strings.Replace(string(executionModelSource()),
		"if input < 10 { return input } else { return 10 }",
		"let value = input; let limit = 10; let result = input; if value < limit { result = value } else { result = limit }; return result", 1))
	doc, err := DecodeSourcePathDocument(ctx, "flow.gooo", source, "Choose", nil)
	if err != nil {
		t.Fatal(err)
	}
	name := writeFlowContractModel(t, false)
	before, err := ExportTypedPathModelContext(ctx, "flow.gooo", source, "Choose", doc, name, decision.ExecutionFlowFeatureVersion)
	if err != nil {
		t.Fatal(err)
	}
	info := before.ModelCompatibility.Model
	if info.ModelSchema != flowdecision.Schema || info.ResidentTensorBytes != 37160 || before.ModelPredictions != 0 || before.CandidateTests != 0 || before.Context.Schema != "gooo/compiler-flow-model-context/v1" {
		t.Fatal("incorrect flow artifact identity or preflight work", before)
	}
	legacy, err := ExportTypedPathModelContext(ctx, "flow.gooo", source, "Choose", doc, writeExecutionContractModel(t), decision.ExecutionFeatureVersion)
	if err != nil {
		t.Fatal(err)
	}
	g, err := NewTypedPathGenerator(name)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(name); err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	for range 3 {
		wg.Go(func() {
			result, err := g.Generate(ctx, "flow.gooo", source, "Choose", doc, TypedPathOptions{StepAttempts: 1})
			if err != nil {
				t.Error(err)
				return
			}
			p := result.Report.BodyPaths
			if p.Search.Status != "TRAINING_COMPLETE" || len(p.Search.Attempts) != 1 || p.Search.Attempts[0].Mask != 2 || p.Search.Selection.ModelCalls != 1 || p.Search.Selection.ModelVariant != "flow_fp32" || p.FunctionalCompleteness != 100 || p.Conditions.Passed != 3 {
				t.Error("static assignment facts did not select and check the body", p)
			}
			if !reflect.DeepEqual(before.Context, p.ModelContext) {
				t.Error("preflight and actual input contexts differ")
			}
			for i, input := range before.Inputs {
				if input.FlowFeatures == nil || input.Features != nil || input.ExecutionFeatures != nil || input.Bytes != 1536 || input.InputSHA != "sha256:"+p.ConditionProgress[0].Ranking.FeatureSHA[i] || input.InputSHA != candidateFeatureDigest(input.FlowFeatures[:]) {
					t.Error("flow array differs from actual prediction", i)
					continue
				}
				if !slices.Equal(input.FlowFeatures[:320], legacy.Inputs[i].ExecutionFeatures[:]) {
					t.Error("flow extension changed the old prefix")
				}
				for _, value := range input.FlowFeatures[256:320] {
					if value != 0 {
						t.Error("future output entered initial prediction")
					}
				}
			}
			if err := VerifyTypedPathProjection(ctx, "flow.gooo", source, doc, result); err != nil {
				t.Error("saved program requires model file", err)
			}
		})
	}
	wg.Wait()
}

func TestFlowModelCompilerOutputFeedback(t *testing.T) {
	ctx := context.Background()
	source := executionModelSource()
	doc, err := DecodeSourcePathDocument(ctx, "flow.gooo", source, "Choose", nil)
	if err != nil {
		t.Fatal(err)
	}
	name := writeFlowContractModel(t, true)
	if _, err := ExportTypedPathModelContext(ctx, "flow.gooo", source, "Choose", doc, name, decision.ExecutionFeatureVersion); err == nil {
		t.Fatal("mismatched explicit feature version accepted")
	}
	result, err := GenerateWithTypedPathFeedback(ctx, "flow.gooo", source, "Choose", doc, name, 1, 3, nil)
	if err != nil {
		t.Fatal(err)
	}
	p := result.Report.BodyPaths
	if len(p.Search.Attempts) != 2 || p.Search.Selection.ModelCalls != 2 || p.FunctionalCompleteness != 100 || p.Conditions.Passed != 3 || len(p.ConditionFeedback) != 1 {
		t.Fatal("flow feedback failed", p)
	}
	f := p.ConditionFeedback[0]
	if f.FeatureVersion != decision.ExecutionFlowFeatureVersion || f.HasFailure || f.OutputFailure == nil || f.OutputFailure.Result.Actual != -9007199254740995 || f.Proposed != 2 {
		t.Fatal("output-only failure lost its exact value or next selection", f)
	}
}
