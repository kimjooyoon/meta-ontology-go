package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"

	decision "github.com/kimjooyoon/gooo-decision-runtime"
)

func TestRunBodyCodegenTinyModelRequiresFillPlanAndRejectsLayaConfiguration(t *testing.T) {
	t.Setenv("GOOO_LAYA_URL", "")
	t.Setenv("GOOO_LAYA_API_KEY", "")
	fixture := `package sample
namespace sample
entity Integer id "sample://entity/integer"
activity Choose(Integer) -> Integer computes "if input < 0 { return __GOOO_BODY_HOLE_value__ } else { return input }"
`
	plan := `{"schema":"gooo/body-codegen-ir-fill-plan/v1","intent":"Choose an arithmetic operation.","hole_id":"value","candidates":[{"id":"sum","expression":"input + 0"},{"id":"difference","expression":"input - 0"}],"test_cases":[{"input":-1,"expected":-1}]}`
	tests := []struct {
		name string
		args []string
		want int
	}{
		{name: "requires fill plan", args: []string{"--tiny-model", "model.json", "--activity", "Choose", "fixture.gooo"}, want: exitUsage},
		{name: "rejects fill search", args: []string{"--tiny-model", "model.json", "--fill-search", "plan.json", "--activity", "Choose", "fixture.gooo"}, want: exitUsage},
		{name: "rejects sample seed", args: []string{"--tiny-model", "model.json", "--fill-plan", "plan.json", "--sample-seed", "seed", "--activity", "Choose", "fixture.gooo"}, want: exitUsage},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			code := runBodyCodegen(test.args, mapSourceReader{"fixture.gooo": []byte(fixture), "plan.json": []byte(plan)}, &stdout, &stderr)
			if code != test.want || !strings.Contains(stderr.String(), "usage:") {
				t.Fatalf("code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
			}
		})
	}

	t.Setenv("GOOO_LAYA_URL", "http://laya.example.invalid/v1/systemone")
	var stdout, stderr bytes.Buffer
	code := runBodyCodegen([]string{"--tiny-model", "model.json", "--fill-plan", "plan.json", "--activity", "Choose", "fixture.gooo"},
		mapSourceReader{"fixture.gooo": []byte(fixture), "plan.json": []byte(plan)}, &stdout, &stderr)
	if code != exitUsage || !strings.Contains(stderr.String(), "GOOO_LAYA_URL") {
		t.Fatalf("Laya endpoint conflict was not rejected: code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
	}
	t.Setenv("GOOO_LAYA_URL", "")
	t.Setenv("GOOO_LAYA_API_KEY", "local-key")
	stdout.Reset()
	stderr.Reset()
	code = runBodyCodegen([]string{"--tiny-model", "model.json", "--fill-plan", "plan.json", "--activity", "Choose", "fixture.gooo"},
		mapSourceReader{"fixture.gooo": []byte(fixture), "plan.json": []byte(plan)}, &stdout, &stderr)
	if code != exitUsage || !strings.Contains(stderr.String(), "GOOO_LAYA_API_KEY") {
		t.Fatalf("Laya credential conflict was not rejected: code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
	}
}

func TestRunBodyCodegenTinyModelAcceptsSourceOwnedMultiHolePlan(t *testing.T) {
	t.Setenv("GOOO_LAYA_URL", "")
	t.Setenv("GOOO_LAYA_API_KEY", "")
	fixture, err := os.ReadFile("../../examples/body-codegen/source-ir-fill.gooo.fixture")
	if err != nil {
		t.Fatal(err)
	}
	modelPath := writeSyntheticTinyGoModel(t, "add")
	var stdout, stderr bytes.Buffer
	code := runBodyCodegen([]string{"--json", "--tiny-model", modelPath, "--activity", "Lift", "fixture.gooo"},
		mapSourceReader{"fixture.gooo": fixture}, &stdout, &stderr)
	if code != exitOK || stderr.Len() != 0 {
		t.Fatalf("source-owned tiny-model fill failed: code=%d stderr=%q stdout=%q", code, stderr.String(), stdout.String())
	}
	var payload struct {
		GoooSource string `json:"gooo_source"`
		Report     struct {
			BodyFill struct {
				Selected string  `json:"selected_candidate_id"`
				Accuracy float64 `json:"functional_accuracy_percent"`
				Decision struct {
					Provider string `json:"provider"`
				} `json:"decision"`
			} `json:"body_fill"`
		} `json:"report"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload.Report.BodyFill.Selected == "" || payload.Report.BodyFill.Accuracy != 100 || payload.Report.BodyFill.Decision.Provider != "tiny_go" {
		t.Fatalf("source plan did not yield a measured local decision: %+v", payload.Report.BodyFill)
	}
	if strings.Contains(payload.GoooSource, "source_fill") || strings.Contains(payload.GoooSource, "__GOOO_BODY_HOLE_") {
		t.Fatalf("generated source retained pending plan instructions:\n%s", payload.GoooSource)
	}
}

func TestRunBodyCodegenTinyModelFillsTypedPlanAndReportsSeparateAccounting(t *testing.T) {
	t.Setenv("GOOO_LAYA_URL", "")
	t.Setenv("GOOO_LAYA_API_KEY", "")
	fixture := `package sample
namespace sample
entity Integer id "sample://entity/integer"
activity Choose(Integer) -> Integer computes "if input < 0 { return __GOOO_BODY_HOLE_value__ } else { return input }"
`
	plan := `{"schema":"gooo/body-codegen-ir-fill-plan/v1","intent":"Add a neutral offset to preserve the value.","hole_id":"value","candidates":[{"id":"sum","expression":"input + 0"},{"id":"difference","expression":"input - 0"}],"test_cases":[{"input":-1,"expected":-1},{"input":2,"expected":2}]}`
	modelPath := writeSyntheticTinyGoModel(t, "add")
	reader := mapSourceReader{"fixture.gooo": []byte(fixture), "plan.json": []byte(plan)}
	var stdout, stderr bytes.Buffer
	code := runBodyCodegen([]string{"--json", "--tiny-model", modelPath, "--fill-plan", "plan.json", "--activity", "Choose", "fixture.gooo"},
		reader, &stdout, &stderr)
	if code != exitOK || stderr.Len() != 0 {
		t.Fatalf("tiny CLI failed: code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
	}
	var payload struct {
		Source string `json:"source"`
		Report struct {
			BodyFill struct {
				Decision struct {
					Provider string `json:"provider"`
					Mode     string `json:"mode"`
				} `json:"decision"`
				LocalPredictions int  `json:"local_model_predictions"`
				ExternalCalls    int  `json:"external_provider_calls"`
				CallsKnown       bool `json:"external_provider_calls_known"`
				Timing           struct {
					LayaMS float64  `json:"laya_decision_ms"`
					TinyMS float64  `json:"tiny_decision_ms"`
					LoadMS *float64 `json:"tiny_model_load_ms"`
				} `json:"timing"`
			} `json:"body_fill"`
			Completeness struct {
				Scope      map[string]any `json:"scope"`
				Dimensions []struct {
					ID          string `json:"id"`
					Status      string `json:"status"`
					Numerator   int    `json:"numerator"`
					Denominator int    `json:"denominator"`
				} `json:"dimensions"`
			} `json:"completeness_receipt"`
		} `json:"report"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		t.Fatalf("decode tiny report: %v; output=%s", err, stdout.String())
	}
	fill := payload.Report.BodyFill
	if fill.Decision.Provider != "tiny_go" || fill.Decision.Mode != "tiny_go" ||
		fill.LocalPredictions != 1 || fill.ExternalCalls != 0 || !fill.CallsKnown || fill.Timing.LayaMS != 0 ||
		fill.Timing.TinyMS <= 0 || fill.Timing.LoadMS == nil || *fill.Timing.LoadMS < 0 {
		t.Fatalf("tiny provider is misreported as external/Laya: %+v", fill)
	}
	if !strings.Contains(payload.Source, "return (input + 0)") {
		t.Fatalf("tiny choice was not emitted from the selected declared candidate: %s", payload.Source)
	}
	if payload.Report.Completeness.Scope["decision_provider"] != "tiny_go" ||
		payload.Report.Completeness.Scope["laya_provider"] != "" {
		t.Fatalf("tiny provider was conflated with Laya in completeness scope: %#v", payload.Report.Completeness.Scope)
	}
	for _, dimension := range payload.Report.Completeness.Dimensions {
		if dimension.ID == "external_network_boundary" && (dimension.Status != "PASS" || dimension.Numerator != 1) {
			t.Fatalf("tiny local decision was counted as external network use: %+v", dimension)
		}
		if dimension.ID == "laya_decision_observation" && dimension.Denominator != 0 {
			t.Fatalf("tiny local decision was counted as an eligible Laya request: %+v", dimension)
		}
	}
}

func TestRunBodyCodegenTinyModelRejectsPlanProviderSelectorAndPropagatesCancellation(t *testing.T) {
	t.Setenv("GOOO_LAYA_URL", "")
	t.Setenv("GOOO_LAYA_API_KEY", "")
	fixture := `package sample
namespace sample
entity Integer id "sample://entity/integer"
activity Choose(Integer) -> Integer computes "if input < 0 { return __GOOO_BODY_HOLE_value__ } else { return input }"
`
	planWithSelector := `{"schema":"gooo/body-codegen-ir-fill-plan/v1","intent":"Choose an operation.","hole_id":"value","provider_model":"english","candidates":[{"id":"sum","expression":"input + 0"},{"id":"difference","expression":"input - 0"}],"test_cases":[{"input":-1,"expected":-1}]}`
	modelPath := writeSyntheticTinyGoModel(t, "add")
	reader := mapSourceReader{"fixture.gooo": []byte(fixture), "plan.json": []byte(planWithSelector)}
	var stdout, stderr bytes.Buffer
	code := runBodyCodegen([]string{"--json", "--tiny-model", modelPath, "--fill-plan", "plan.json", "--activity", "Choose", "fixture.gooo"},
		reader, &stdout, &stderr)
	if code != exitFailure || !strings.Contains(stdout.String(), "provider_model selector") {
		t.Fatalf("tiny path did not reject a provider selector: code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
	}

	plan := strings.Replace(planWithSelector, `,"provider_model":"english"`, "", 1)
	reader["plan.json"] = []byte(plan)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	stdout.Reset()
	stderr.Reset()
	code = runBodyCodegenContext(ctx, []string{"--json", "--tiny-model", modelPath, "--fill-plan", "plan.json", "--activity", "Choose", "fixture.gooo"},
		reader, &stdout, &stderr)
	if code != exitFailure || !strings.Contains(stdout.String(), "context canceled") {
		t.Fatalf("signal cancellation did not close tiny generation: code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
	}
}

func TestRunBodyCodegenTinyModelLoadErrorDoesNotPrintCallerPath(t *testing.T) {
	t.Setenv("GOOO_LAYA_URL", "")
	t.Setenv("GOOO_LAYA_API_KEY", "")
	fixture := `package sample
namespace sample
entity Integer id "sample://entity/integer"
activity Choose(Integer) -> Integer computes "if input < 0 { return __GOOO_BODY_HOLE_value__ } else { return input }"
`
	plan := `{"schema":"gooo/body-codegen-ir-fill-plan/v1","intent":"Choose an operation.","hole_id":"value","candidates":[{"id":"sum","expression":"input + 0"},{"id":"difference","expression":"input - 0"}],"test_cases":[{"input":-1,"expected":-1}]}`
	modelPath := filepath.Join(t.TempDir(), "private-model-directory", "missing-model.json")
	reader := mapSourceReader{"fixture.gooo": []byte(fixture), "plan.json": []byte(plan)}
	var stdout, stderr bytes.Buffer
	code := runBodyCodegen([]string{"--tiny-model", modelPath, "--fill-plan", "plan.json", "--activity", "Choose", "fixture.gooo"},
		reader, &stdout, &stderr)
	if code != exitFailure || strings.Contains(stdout.String()+stderr.String(), modelPath) ||
		!strings.Contains(stderr.String(), tinyModelDiagnosticLabel) {
		t.Fatalf("load failure leaked caller model path: code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
	}
}

func writeSyntheticTinyGoModel(t *testing.T, winner string) string {
	t.Helper()
	directory := t.TempDir()
	labels := decision.Labels()
	metadata := decision.Metadata{
		Schema: decision.MetadataSchema, Variant: "fp32", FeatureDim: decision.FeatureDim,
		HiddenDim: decision.HiddenDim, MaxBytes: decision.InputMaxBytes, Labels: labels[:],
		Temperature: 1, WeightsFile: "weights.bin",
	}
	var weights []byte
	for _, tensor := range []struct {
		name  string
		rows  int
		cols  int
		count int
	}{
		{"w1", decision.HiddenDim, decision.FeatureDim, decision.HiddenDim * decision.FeatureDim},
		{"b1", 1, decision.HiddenDim, decision.HiddenDim},
		{"w2", decision.LabelCount, decision.HiddenDim, decision.LabelCount * decision.HiddenDim},
		{"b2", 1, decision.LabelCount, decision.LabelCount},
	} {
		raw := make([]byte, tensor.count*4)
		if tensor.name == "b2" {
			index := -1
			for i, label := range labels {
				if label == winner {
					index = i
					break
				}
			}
			if index < 0 {
				t.Fatalf("unknown synthetic label %q", winner)
			}
			binary.LittleEndian.PutUint32(raw[index*4:], math.Float32bits(10))
		}
		metadata.Tensors = append(metadata.Tensors, decision.TensorMetadata{
			Name: tensor.name, Count: tensor.count, Rows: tensor.rows, Cols: tensor.cols,
			Encoding: "float32_le", Offset: int64(len(weights)), Bytes: int64(len(raw)), Scale: 1,
		})
		weights = append(weights, raw...)
	}
	weightsDigest := sha256.Sum256(weights)
	metadata.WeightsSHA256 = hex.EncodeToString(weightsDigest[:])
	metadataBytes, err := json.Marshal(metadata)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(directory, metadata.WeightsFile), weights, 0o600); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(directory, "model.json")
	if err := os.WriteFile(path, metadataBytes, 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}
