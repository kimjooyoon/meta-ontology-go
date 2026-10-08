package bodyexecution

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	decision "github.com/kimjooyoon/gooo-decision-runtime"
	"github.com/kimjooyoon/gooo-decision-runtime/jointdecision"
)

// All-zero synthetic weights exercise graph transport and finite continuation;
// they supply no learned solution. Ties retain deterministic mask order.
func zeroGraphModel(t *testing.T) string {
	t.Helper()
	meta := jointdecision.Metadata{Schema: jointdecision.SharedThreeSchema, Variant: "fp32",
		Feature: jointdecision.RecordGraphSharedFeatureVersion, Arithmetic: jointdecision.SeparateArithmeticVersion,
		FeatureDim: 768, HiddenDim: 8, MaxBytes: jointdecision.RecordGraphInputMaxBytes,
		Temperature: 1, WeightsFile: "weights.bin"}
	var weights []byte
	for i, name := range []string{"w1", "b1", "w2"} {
		rows, cols := [3]int{8, 1, 2}[i], [3]int{256, 8, 8}[i]
		meta.Tensors = append(meta.Tensors, decision.TensorMetadata{Name: name, Rows: rows, Cols: cols,
			Count: rows * cols, Encoding: "float32_le", Offset: int64(len(weights)), Bytes: int64(4 * rows * cols), Scale: 1})
		weights = append(weights, make([]byte, 4*rows*cols)...)
	}
	for i := range 8 {
		meta.Labels = append(meta.Labels, fmt.Sprintf("mask_%d", i))
	}
	meta.WeightsSHA = fmt.Sprintf("%x", sha256.Sum256(weights))
	root := t.TempDir()
	raw, err := json.Marshal(meta)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(root, "weights.bin"), weights, 0600); err != nil {
		t.Fatal(err)
	}
	name := filepath.Join(root, "model.json")
	if err = os.WriteFile(name, raw, 0600); err != nil {
		t.Fatal(err)
	}
	return name
}

func TestSourceGraphContractConstructsAndExecutesNativeRecords(t *testing.T) {
	source, err := os.ReadFile("../../examples/body-codegen/record-field-assembly.gooo.fixture")
	if err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile("../../examples/body-codegen/record-field-assembly-cases.json")
	if err != nil {
		t.Fatal(err)
	}
	suite, err := DecodeCompositionCases(raw)
	if err != nil {
		t.Fatal(err)
	}
	prior, err := GenerateComposition(context.Background(), "graph.gooo", source, suite, zeroGraphModel(t))
	if err != nil {
		t.Fatal(err)
	}
	assembly := prior.Steps[0].Generation.Report.RecordAssembly
	if assembly.ModelCalls != 1 || assembly.Context.FeatureVersion != jointdecision.RecordGraphSharedFeatureVersion ||
		assembly.FieldsPassed != 15 || len(assembly.Attempts) != 8 || assembly.Attempts[0].Mask != 0 {
		t.Fatal("graph prediction did not retain finite candidate continuation", assembly)
	}
	for range 2 {
		run, err := ExecuteComposition(context.Background(), "graph.gooo", source, prior, suite, nativeTool())
		if err != nil || run.FinitePassed != 14 || run.FiniteTotal != 14 || run.ModelCalls != 0 || !run.RuntimeReplayed {
			t.Fatal("graph model construction lost native execution or replay", run, err)
		}
	}
}
