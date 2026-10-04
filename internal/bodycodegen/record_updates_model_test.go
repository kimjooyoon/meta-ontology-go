package bodycodegen

import (
	"context"
	"testing"

	"github.com/kimjooyoon/gooo-decision-runtime/jointdecision"
)

func TestRecordUpdateUsesOneSharedPredictionAndDeterministicReconstruction(t *testing.T) {
	source, ctx := recordUpdatesFixture(t), context.Background()
	exported, err := ExportRecordAssemblyContextWithFeature(ctx, "updates.gooo", source, "Select", false, jointdecision.RecordSharedFeatureVersion)
	if err != nil || exported.ModelPredictions != 0 || exported.CandidateTests != 0 || exported.ExpandedPlan != nil {
		t.Fatal("context", err)
	}
	if exported.Choices[2].Kind != "field_update" || exported.Choices[2].Second != `copy.title + ":" + copy.state` {
		t.Fatal(exported.Choices)
	}
	g, err := NewTypedPathGenerator(writeSharedRecordContractModel(t, "qat_ternary"))
	if err != nil {
		t.Fatal(err)
	}
	result, err := g.GenerateSourceAssembly(ctx, "updates.gooo", source, "Select")
	if err != nil {
		t.Fatal(err)
	}
	r := result.Report.RecordAssembly
	if r.ModelCalls != 1 || r.Prediction == nil || r.Context.Text != exported.Context.Text || r.FieldsPassed != 15 {
		t.Fatal(r)
	}
	realized, err := RealizeSourceAssembly(ctx, "updates.gooo", source, result)
	if err != nil || realized.ModelCalls != 0 {
		t.Fatal("reconstruction", err)
	}
	r.Choices[0].Kind = "field_value"
	if _, err = RealizeSourceAssembly(ctx, "updates.gooo", source, result); err == nil {
		t.Fatal("choice kind does not reconstruct")
	}
}
