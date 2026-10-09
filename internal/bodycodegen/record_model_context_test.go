package bodycodegen

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/kimjooyoon/gooo-decision-runtime/jointdecision"
)

func recordContextOwnModel(t *testing.T) ([]byte, string) {
	t.Helper()
	source, err := os.ReadFile("../../examples/scalar-identity/source.gooo.fixture")
	if err != nil {
		t.Fatal(err)
	}
	return source, "../../examples/scalar-identity/model/model.json"
}

func TestRecordModelPreflightMatchesActualRankingContext(t *testing.T) {
	source, model := recordContextOwnModel(t)
	ctx := context.Background()
	exported, err := ExportRecordAssemblyModelContext(ctx, "r.gooo", source, "Describe", model,
		RecordModelContextOptions{FeatureVersion: jointdecision.RecordGraphSharedFeatureVersion, IncludePlan: true, ValueFlow: true})
	if err != nil {
		t.Fatal(err)
	}
	g, err := NewTypedPathGenerator(model)
	if err != nil {
		t.Fatal(err)
	}
	selected, err := g.GenerateSourceAssembly(ctx, "r.gooo", source, "Describe")
	if err != nil {
		t.Fatal(err)
	}
	r := selected.Report.RecordAssembly
	if exported.Context.SHA256 != r.Context.SHA256 || exported.Context.Text != r.Context.Text ||
		exported.OriginalSourceSHA256 != r.OriginalSourceSHA256 || exported.ContractSHA256 != r.ContractSHA256 ||
		exported.ModelCompatibility.Model.MetadataSHA256 != r.Model.MetadataSHA256 ||
		exported.ModelCompatibility.Model.ResidentTensorBytes != 2096 || exported.ExpandedPlan == nil || exported.ValueFlow == nil ||
		exported.ModelPredictions != 0 || exported.CandidateTests != 0 || r.ModelCalls != 1 || r.FieldsPassed != 12 {
		t.Fatal("preflight differs from actual model input or performs selection")
	}
}

func TestRecordModelPreflightRejectsInvalidArtifactAndCancelledContext(t *testing.T) {
	source, model := recordContextOwnModel(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	for _, c := range []context.Context{nil, ctx} {
		if _, err := ExportRecordAssemblyModelContext(c, "r.gooo", source, "Describe", model, RecordModelContextOptions{}); err == nil {
			t.Fatal("missing or cancelled context accepted")
		}
	}
	if _, err := ExportRecordAssemblyModelContext(context.Background(), "r.gooo", source, "Describe", "", RecordModelContextOptions{}); err == nil {
		t.Fatal("implicit model accepted")
	}
	raw, err := os.ReadFile(model)
	if err != nil {
		t.Fatal(err)
	}
	weights, err := os.ReadFile(filepath.Join(filepath.Dir(model), "weights.bin"))
	if err != nil {
		t.Fatal(err)
	}
	weights[0] ^= 1
	dir := t.TempDir()
	if err = os.WriteFile(filepath.Join(dir, "model.json"), raw, 0600); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(dir, "weights.bin"), weights, 0600); err != nil {
		t.Fatal(err)
	}
	if _, err = ExportRecordAssemblyModelContext(context.Background(), "r.gooo", source, "Describe", filepath.Join(dir, "model.json"), RecordModelContextOptions{}); err == nil {
		t.Fatal("changed weights advertised as compatible")
	}
}

func TestRecordContextWithoutModelRetainsOriginalExport(t *testing.T) {
	source, _ := recordContextOwnModel(t)
	exported, err := ExportRecordAssemblyContextWithFeature(context.Background(), "r.gooo", source, "Describe", false,
		jointdecision.RecordGraphSharedFeatureVersion)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(exported)
	if err != nil {
		t.Fatal(err)
	}
	var fields map[string]json.RawMessage
	if err = json.Unmarshal(raw, &fields); err != nil {
		t.Fatal(err)
	}
	if _, present := fields["model_compatibility"]; present || exported.ModelPredictions != 0 || exported.CandidateTests != 0 {
		t.Fatal("model-free export contract changed")
	}
}
