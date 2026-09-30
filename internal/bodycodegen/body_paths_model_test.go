package bodycodegen

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	decision "github.com/kimjooyoon/gooo-decision-runtime"
)

// This synthetic constant-logit bundle tests calls and ABI isolation only.
// It is deliberately not evidence for language understanding or trained quality.
func writeTypedPathContractModel(t *testing.T, operationABI bool) string {
	t.Helper()
	labels := decision.PathLabels()
	modelSchema := decision.PathMetadataSchema
	if operationABI {
		labels = decision.Labels()
		modelSchema = decision.MetadataSchema
	}
	threshold := 1.0
	metadata := decision.Metadata{Schema: modelSchema, Variant: "fp32", FeatureDim: decision.FeatureDim,
		HiddenDim: decision.HiddenDim, MaxBytes: decision.InputMaxBytes, Labels: labels[:], Temperature: 1,
		WeightsFile: "weights.bin", ConfidenceThreshold: &threshold}
	var raw []byte
	for _, tensor := range []struct {
		name       string
		rows, cols int
	}{{"w1", decision.HiddenDim, decision.FeatureDim}, {"b1", 1, decision.HiddenDim},
		{"w2", decision.LabelCount, decision.HiddenDim}, {"b2", 1, decision.LabelCount}} {
		count := tensor.rows * tensor.cols
		metadata.Tensors = append(metadata.Tensors, decision.TensorMetadata{Name: tensor.name, Count: count,
			Rows: tensor.rows, Cols: tensor.cols, Encoding: "float32_le", Offset: int64(len(raw)), Bytes: int64(count * 4), Scale: 1})
		raw = append(raw, make([]byte, count*4)...)
	}
	sum := sha256.Sum256(raw)
	metadata.WeightsSHA256 = hex.EncodeToString(sum[:])
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "weights.bin"), raw, 0600); err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(metadata)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "model.json")
	if err := os.WriteFile(path, encoded, 0600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestTypedPathModelCalledOncePerDecisionBeforeTDD(t *testing.T) {
	source, document := typedPathFixture(t)
	result, err := GenerateWithTypedPaths(context.Background(), "fixture.gooo", source, "Combined", document,
		writeTypedPathContractModel(t, false))
	if err != nil {
		t.Fatal(err)
	}
	receipt := result.Report.BodyPaths
	if receipt.Search.Selection.ModelCalls != 3 || receipt.Search.Selection.ExternalCalls != 0 ||
		len(receipt.Search.Selection.Receipts) != 3 || receipt.Search.ModelAbstentionsObserved != 3 ||
		receipt.FunctionalCompleteness != 100 || receipt.Search.Selection.WeightsSHA256 == "" ||
		receipt.Search.Selection.MetadataSHA256 == "" {
		t.Fatalf("model call contract: %+v", receipt)
	}
	for _, prediction := range receipt.Search.Selection.Receipts {
		if prediction.PredictNS <= 0 || prediction.Proposed == "" {
			t.Fatal("prediction was not recorded before selection")
		}
	}
	if _, err := GenerateWithTypedPaths(context.Background(), "fixture.gooo", source, "Combined", document,
		writeTypedPathContractModel(t, true)); err == nil {
		t.Fatal("operator model was silently reinterpreted as a structural model")
	}
}
