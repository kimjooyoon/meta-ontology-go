package toolchainrelease

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/kimjooyoon/gooo-decision-runtime/contractdecision"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

// Fresh zero-weight context export performs no fit, prediction or native run.
func TestRelativeV3PreflightReaderBindsOriginalSourceAndAllArrays(t *testing.T) {
	source, err := os.ReadFile(filepath.Join("../../../../", relativeV3Root+"cases.gooo.fixture"))
	if err != nil {
		t.Fatal(err)
	}
	d, err := bodycodegen.DecodeSourcePathDocument(context.Background(), "relative-v3.gooo", source, "Choose", nil)
	if err != nil {
		t.Fatal(err)
	}
	m, err := contractdecision.NewSourceRelativeV3([contractdecision.InteractionRequirementParameterCount]float32{})
	if err != nil {
		t.Fatal(err)
	}
	model, err := m.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	name := filepath.Join(t.TempDir(), "model.json")
	if err = os.WriteFile(name, model, 0600); err != nil {
		t.Fatal(err)
	}
	exported, err := bodycodegen.ExportTypedPathModelContext(context.Background(), "relative-v3.gooo", source, "Choose", d, name, "")
	if err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(exported)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = validateRelativeV3Preflight(raw, source, model); err != nil {
		t.Fatal(err)
	}
	for _, change := range []func(*bodycodegen.TypedPathContextExport){
		func(r *bodycodegen.TypedPathContextExport) { r.ModelPredictions = 1 },
		func(r *bodycodegen.TypedPathContextExport) { r.Inputs[0].SourceRelativeOutputs.Counts[0][0]++ },
		func(r *bodycodegen.TypedPathContextExport) { r.Inputs[0].SourceRelativeFeatures[0]++ },
		func(r *bodycodegen.TypedPathContextExport) { r.Inputs[0].InputSHA = "changed" },
		func(r *bodycodegen.TypedPathContextExport) { r.ContractCases.Features[0][0]++ },
		func(r *bodycodegen.TypedPathContextExport) { r.ContractConditions.Features[0][0]++ },
		func(r *bodycodegen.TypedPathContextExport) { r.Context.ModelFingerprint = "changed" },
	} {
		var r bodycodegen.TypedPathContextExport
		if err = json.Unmarshal(raw, &r); err != nil {
			t.Fatal(err)
		}
		change(&r)
		bad, err := json.Marshal(r)
		if err != nil {
			t.Fatal(err)
		}
		if _, err = validateRelativeV3Preflight(bad, source, model); err == nil {
			t.Fatal("altered preflight accepted")
		}
	}
}
