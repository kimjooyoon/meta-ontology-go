package bodycodegen

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	decision "github.com/kimjooyoon/gooo-decision-runtime"
	"github.com/kimjooyoon/gooo-decision-runtime/conditiondecision"
	"github.com/kimjooyoon/gooo-decision-runtime/executiondecision"
)

func writeCandidateArtifact(t *testing.T, raw []byte) string {
	t.Helper()
	name := filepath.Join(t.TempDir(), "candidate.json")
	if err := os.WriteFile(name, raw, 0600); err != nil {
		t.Fatal(err)
	}
	return name
}

// Deliberate weights isolate the compiler's observation wiring. Learned model
// quality is measured separately with the public study's unchanged artifacts.
func writeExecutionContractModel(t *testing.T) string {
	t.Helper()
	var weights [executiondecision.ParameterCount]float32
	weights[256], weights[311] = 8, 128
	weights[executiondecision.FeatureDim*executiondecision.HiddenDim] = -1
	weights[executiondecision.FeatureDim*executiondecision.HiddenDim+2*executiondecision.HiddenDim] = 2
	m, err := executiondecision.New(weights)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := m.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	return writeCandidateArtifact(t, raw)
}

func executionModelSource() []byte {
	return []byte(`package executionmodel
namespace executionmodel
entity Integer id "executionmodel://integer"
activity Choose(Integer) -> Integer computes "if input < 10 { return input } else { return 10 }" assembling {
 choice "comparison" operand_order at "0" intent "입력과 10을 비교한다. Compare input with 10."
 choice "branches" branch_layout at "0" intent "더 큰 값을 반환한다. Return the larger value."
 case "-9007199254740995" -> "10"
 case "0" -> "10"
 case "10" -> "10"
 case "9007199254740993" -> "9007199254740993"
 case "9007199254740995" -> "9007199254740995"
 case "18014398509481990" -> "18014398509481990"
 condition_case "comparison" input "-9007199254740995" -> "true"
 condition_case "comparison" input "10" -> "false"
 condition_case "comparison" input "9007199254740995" -> "false"
 attempts "4"
}
`)
}

func TestExecutionModelCompilerInterleavesOutputFeedback(t *testing.T) {
	ctx := context.Background()
	source := executionModelSource()
	doc, err := DecodeSourcePathDocument(ctx, "execution.gooo", source, "Choose", nil)
	if err != nil {
		t.Fatal(err)
	}
	name := writeExecutionContractModel(t)
	before, err := ExportTypedPathModelContext(ctx, "execution.gooo", source, "Choose", doc, name, decision.ExecutionFeatureVersion)
	if err != nil {
		t.Fatal(err)
	}
	info := before.ModelCompatibility.Model
	if info.FeatureVersion != decision.ExecutionFeatureVersion || info.ModelSchema != executiondecision.Schema || info.ResidentTensorBytes != 31016 || before.ModelPredictions != 0 || before.CandidateTests != 0 {
		t.Fatal("incorrect execution model identity or preflight work", before)
	}
	result, err := GenerateWithTypedPathFeedback(ctx, "execution.gooo", source, "Choose", doc, name, 1, 3, nil)
	if err != nil {
		t.Fatal(err)
	}
	p := result.Report.BodyPaths
	if p.Search.Status != "TRAINING_COMPLETE" || len(p.Search.Attempts) != 2 || p.Search.Selection.ModelCalls != 2 || p.Search.Selection.ModelVariant != "execution_fp32" || p.FunctionalCompleteness != 100 || p.Conditions.Passed != 3 {
		t.Fatal("output feedback did not drive native codegen", p)
	}
	if len(p.ConditionFeedback) != 1 || p.ConditionFeedback[0].HasFailure || p.ConditionFeedback[0].OutputFailure == nil || p.ConditionFeedback[0].OutputFailure.Result.Actual != -9007199254740995 || p.ConditionFeedback[0].Proposed != 2 {
		t.Fatal("condition success hid output failure", p.ConditionFeedback)
	}
	if !reflect.DeepEqual(before.Context, p.ModelContext) {
		t.Fatal("preflight and generation prepared different context")
	}
	for i, input := range before.Inputs {
		if input.Features != nil || input.ExecutionFeatures == nil || input.Bytes != 1280 || input.InputSHA != "sha256:"+p.ConditionProgress[0].Ranking.FeatureSHA[i] || input.InputSHA != candidateFeatureDigest(input.ExecutionFeatures[:]) {
			t.Fatal("execution array differs from actual ranking", i)
		}
		for _, value := range input.ExecutionFeatures[256:] {
			if value != 0 {
				t.Fatal("future output leaked into initial model input")
			}
		}
	}
	if err := os.Remove(name); err != nil {
		t.Fatal(err)
	}
	if err := VerifyTypedPathProjection(ctx, "execution.gooo", source, doc, result); err != nil {
		t.Fatal("saved selected program required a model", err)
	}
}

func TestBranchRoleModelCompilerUsesDeclaredVersion(t *testing.T) {
	m, err := conditiondecision.NewForFeatures([conditiondecision.ParameterCount]float32{}, decision.ConditionBranchFeatureVersion)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := m.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	name := writeCandidateArtifact(t, raw)
	source := executionModelSource()
	ctx := context.Background()
	doc, err := DecodeSourcePathDocument(ctx, "roles.gooo", source, "Choose", nil)
	if err != nil {
		t.Fatal(err)
	}
	before, err := ExportTypedPathModelContext(ctx, "roles.gooo", source, "Choose", doc, name, decision.ConditionBranchFeatureVersion)
	if err != nil {
		t.Fatal(err)
	}
	result, err := GenerateWithTypedPathFeedback(ctx, "roles.gooo", source, "Choose", doc, name, 1, 3, nil)
	if err != nil {
		t.Fatal(err)
	}
	ranking := result.Report.BodyPaths.ConditionProgress[0].Ranking
	if ranking.FeatureVersion != decision.ConditionBranchFeatureVersion || before.ModelCompatibility.Model.FeatureVersion != ranking.FeatureVersion {
		t.Fatal("v2 artifact silently described as v1")
	}
	for i, input := range before.Inputs {
		if input.Features == nil || input.ExecutionFeatures != nil || input.Bytes != 1024 || input.InputSHA != "sha256:"+ranking.FeatureSHA[i] {
			t.Fatal("v2 preflight differs from actual input", i)
		}
	}
	if before.Inputs[1].Features[236] == 0 || before.Inputs[1].Features[247] == 0 {
		t.Fatal("branch return roles absent")
	}
	if _, err := ExportTypedPathModelContext(ctx, "roles.gooo", source, "Choose", doc, name, decision.ConditionChannelFeatureVersion); err == nil {
		t.Fatal("mismatched explicit version accepted")
	}
}
