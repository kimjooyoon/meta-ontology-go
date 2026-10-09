package bodycodegen

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/kimjooyoon/gooo-decision-runtime/jointdecision"
)

func writeRecordContractModel(t *testing.T) string {
	t.Helper()
	name := writeThreeContractModel(t)
	raw, err := os.ReadFile(name)
	if err != nil {
		t.Fatal(err)
	}
	raw = []byte(strings.Replace(string(raw), jointdecision.ThreeFeatureVersion, jointdecision.RecordFieldFeatureVersion, 1))
	if err = os.WriteFile(name, raw, 0600); err != nil {
		t.Fatal(err)
	}
	return name
}

func TestRecordFieldExpressionContextMatchesSourceAndFiniteReplay(t *testing.T) {
	source := recordAssemblyFixture(t)
	ctx := context.Background()
	exported, err := ExportRecordAssemblyContextWithFeature(ctx, "record.gooo", source, "Select", false, jointdecision.RecordFieldFeatureVersion)
	if err != nil || exported.ModelPredictions != 0 || exported.CandidateTests != 0 || exported.ExpandedPlan != nil {
		t.Fatal("source-only field export", err)
	}
	if exported.Context.Schema != "gooo/record-field-expression-context/v1" || exported.Context.FeatureVersion != jointdecision.RecordFieldFeatureVersion ||
		strings.Contains(exported.Context.Text, "value_case") || strings.Contains(exported.Context.Text, "queued") {
		t.Fatal("field context contract or case separation", exported.Context)
	}
	for i, part := range exported.Context.Parts {
		var encoded jointdecision.RecordChoice
		if json.Unmarshal([]byte(part), &encoded) != nil || encoded.Field != exported.Choices[i].Field ||
			encoded.First != exported.Choices[i].First || encoded.Second != exported.Choices[i].Second || encoded.Intent != exported.Choices[i].Intent {
			t.Fatal("source field expression or intent differs")
		}
	}
	var features [jointdecision.ThreeFeatureDim]float32
	if err := jointdecision.FeaturesIntoRecordThree(exported.Context.Text, &features); err != nil {
		t.Fatal(err)
	}
	g, err := NewTypedPathGenerator(writeRecordContractModel(t))
	if err != nil {
		t.Fatal(err)
	}
	result, err := g.GenerateSourceAssembly(ctx, "record.gooo", source, "Select")
	if err != nil {
		t.Fatal(err)
	}
	r := result.Report.RecordAssembly
	if r.ModelCalls != 1 || r.PredictNS < 0 || r.Prediction == nil || r.Context.Text != exported.Context.Text ||
		r.SelectedMask != 7 || r.FieldsPassed != 15 || len(r.Attempts) != 8 || r.Attempts[0].Mask != 0 ||
		result.Report.CompletenessReceipt.Scope["decision_mode"] != "local_field_expression_prediction_then_finite_tdd" {
		t.Fatal("record-specific inference and actual finite continuation", r)
	}
	if _, err := RealizeSourceAssembly(ctx, "record.gooo", source, result); err != nil {
		t.Fatal("record-specific source replay", err)
	}
	// Captured field-feature identity is checked against the source and model
	// contract during deterministic realization; no second prediction occurs.
	r.Context.Parts[0] = strings.Replace(r.Context.Parts[0], "title", "caption", 1)
	if _, err := RealizeSourceAssembly(ctx, "record.gooo", source, result); err == nil {
		t.Fatal("changed source field context accepted")
	}
}

func TestRecordFieldContextDeclinesWholeOversizeInputAndArity(t *testing.T) {
	ctx := context.Background()
	source := recordAssemblyFixture(t)
	g, err := NewTypedPathGenerator(writeRecordContractModel(t))
	if err != nil {
		t.Fatal(err)
	}
	large := []byte(strings.Replace(string(source), "Keep the original title.", strings.Repeat("한", 140), 1))
	result, err := g.GenerateSourceAssembly(ctx, "large.gooo", large, "Select")
	if err != nil {
		t.Fatal(err)
	}
	r := result.Report.RecordAssembly
	if r.Context.Status != "DECLINED_TO_DETERMINISTIC" || r.ModelCalls != 0 || r.Prediction != nil || r.SelectedMask != 7 ||
		!strings.Contains(r.Context.Parts[0], strings.Repeat("한", 140)) {
		t.Fatal("oversize input was shortened or inferred", r)
	}
	if _, err = RealizeSourceAssembly(ctx, "large.gooo", large, result); err != nil {
		t.Fatal(err)
	}
	context := recordModelContext(r.Choices[:2], jointdecision.RecordFieldFeatureVersion)
	if context.Status != "DECLINED_TO_DETERMINISTIC" || context.Reason != "THREE_FIELD_CHOICES_REQUIRED" || len(context.Parts) != 2 {
		t.Fatal("non-three field source was lost")
	}
	if _, err := ExportRecordAssemblyContextWithFeature(ctx, "record.gooo", source, "Select", false, "unknown"); err == nil {
		t.Fatal("unknown field features exported")
	}
}
