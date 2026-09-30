package decisionroute

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"

	decision "github.com/kimjooyoon/gooo-decision-runtime"
)

func TestTinyGoProviderLoadsSyntheticBundleAndMapsClosedOperationToDeclaredID(t *testing.T) {
	provider := loadSyntheticTinyGoProvider(t, TinyGoOperationLessEqual, 0.1)
	request := tinyGoRequest("compare the amount with the limit", TinyGoOperationAdd, TinyGoOperationLessEqual)
	request.Question.Options[0].ID = "candidate-add-id"
	request.Question.Options[1].ID = "candidate-boundary-id"
	request.Fallback = "candidate-add-id"

	receipt, err := provider.Resolve(context.Background(), request)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if receipt.Provider != ProviderTinyGo || receipt.Mode != ProviderTinyGo ||
		receipt.Selected != "candidate-boundary-id" || receipt.FallbackReason != "" ||
		receipt.TinyGoPredictedOperation != TinyGoOperationLessEqual ||
		receipt.TinyGoPredictionApplied == nil || !*receipt.TinyGoPredictionApplied {
		t.Fatalf("unexpected provider receipt: %+v", receipt)
	}
	if receipt.TinyGoVariant != "fp32" || len(receipt.TinyGoWeightsSHA256) != 64 ||
		len(receipt.TinyGoMetadataSHA256) != 64 || receipt.TinyGoMetadataSHA256 != strings.ToLower(receipt.TinyGoMetadataSHA256) ||
		receipt.Model != "" || receipt.RequestedProviderModel != "" {
		t.Fatalf("tiny_go provenance overlaps or omits Laya fields: %+v", receipt)
	}
	if receipt.TinyGoMetadataSHA256 != provider.model.MetadataSHA256() {
		t.Fatalf("receipt metadata SHA %q does not match loaded model snapshot %q", receipt.TinyGoMetadataSHA256, provider.model.MetadataSHA256())
	}
	if len(receipt.Probabilities) != 2 || receipt.Probabilities["candidate-boundary-id"] < 0.99 {
		t.Fatalf("probabilities were not remapped from labels to declared IDs: %+v", receipt.Probabilities)
	}
	if receipt.Confidence == nil || *receipt.Confidence < 0.99 {
		t.Fatalf("model confidence was not preserved: %+v", receipt.Confidence)
	}
}

func TestTinyGoMetadataSHA256BindsConfigurationAndLoadedSnapshot(t *testing.T) {
	firstPath := writeSyntheticTinyGoBundle(t, TinyGoOperationAdd, 0.1)
	secondPath := writeSyntheticTinyGoBundle(t, TinyGoOperationAdd, 0.2)
	first, err := LoadTinyGoProvider(firstPath)
	if err != nil {
		t.Fatal(err)
	}
	second, err := LoadTinyGoProvider(secondPath)
	if err != nil {
		t.Fatal(err)
	}
	firstRaw, err := os.ReadFile(firstPath)
	if err != nil {
		t.Fatal(err)
	}
	secondRaw, err := os.ReadFile(secondPath)
	if err != nil {
		t.Fatal(err)
	}
	firstDigest := sha256.Sum256(firstRaw)
	secondDigest := sha256.Sum256(secondRaw)
	if first.model.WeightsSHA256() != second.model.WeightsSHA256() {
		t.Fatal("changing only the threshold unexpectedly changed weights digest")
	}
	if first.model.MetadataSHA256() != hex.EncodeToString(firstDigest[:]) ||
		second.model.MetadataSHA256() != hex.EncodeToString(secondDigest[:]) {
		t.Fatal("metadata digest did not match exact loaded JSON bytes")
	}
	if first.model.MetadataSHA256() == second.model.MetadataSHA256() {
		t.Fatal("metadata digest failed to bind changed threshold with identical weights")
	}

	wantSnapshot := first.model.MetadataSHA256()
	if err := os.WriteFile(firstPath, append(firstRaw, '\n'), 0o600); err != nil {
		t.Fatal(err)
	}
	if first.model.MetadataSHA256() != wantSnapshot {
		t.Fatal("metadata digest changed after the loaded metadata file was rewritten")
	}
}

func TestTinyGoProviderUsesOneToOneClosedEightOperationMapping(t *testing.T) {
	provider := loadSyntheticTinyGoProvider(t, TinyGoOperationOr, 0.1)
	labels := decision.Labels()
	request := Request{
		Schema: RequestSchema, State: "opaque-local-state",
		Intent:   "combine the two Boolean values",
		Question: Question{ID: "op-choice", Instructions: "opaque local instruction"},
		Fallback: "id-or",
	}
	for _, label := range labels {
		request.Question.Options = append(request.Question.Options, Option{
			ID: "id-" + label, Operation: label,
			Description: "opaque candidate source expression for " + label,
		})
	}
	receipt, err := provider.Resolve(context.Background(), request)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if receipt.Selected != "id-or" || len(receipt.Probabilities) != decision.LabelCount {
		t.Fatalf("all labels should map to declared unique IDs: %+v", receipt)
	}
	if receipt.TinyGoPredictedOperation != TinyGoOperationOr || receipt.TinyGoPredictionApplied == nil || !*receipt.TinyGoPredictionApplied {
		t.Fatalf("closed model prediction was not recorded as applied: %+v", receipt)
	}
	if len(tinyGoOperations()) != decision.LabelCount {
		t.Fatalf("closed operation contract has %d labels, want %d", len(tinyGoOperations()), decision.LabelCount)
	}
}

func TestTinyGoProviderFallsBackForAbstentionAndUnofferedOperation(t *testing.T) {
	t.Run("low confidence", func(t *testing.T) {
		provider := loadSyntheticTinyGoProvider(t, "", 0.9)
		request := tinyGoRequest("choose the best operation", TinyGoOperationAdd, TinyGoOperationSubtract)
		request.Fallback = "candidate-subtract"
		receipt, err := provider.Resolve(context.Background(), request)
		if err != nil {
			t.Fatalf("Resolve: %v", err)
		}
		if receipt.Mode != "deterministic_fallback" || receipt.Selected != request.Fallback ||
			receipt.FallbackReason != TinyGoFallbackLowConfidence || receipt.Provider != ProviderTinyGo {
			t.Fatalf("unexpected abstention fallback: %+v", receipt)
		}
		assertTinyGoPredictionApplied(t, receipt, TinyGoOperationAdd, false)
	})

	t.Run("operation not offered", func(t *testing.T) {
		provider := loadSyntheticTinyGoProvider(t, TinyGoOperationSubtract, 0.1)
		request := tinyGoRequest("subtract these values", TinyGoOperationAdd, TinyGoOperationMultiply)
		receipt, err := provider.Resolve(context.Background(), request)
		if err != nil {
			t.Fatalf("Resolve: %v", err)
		}
		if receipt.Mode != "deterministic_fallback" || receipt.Selected != request.Fallback ||
			receipt.FallbackReason != TinyGoFallbackOperationNotOffered {
			t.Fatalf("unexpected unoffered operation fallback: %+v", receipt)
		}
		assertTinyGoPredictionApplied(t, receipt, TinyGoOperationSubtract, false)
	})
}

func TestTinyGoPredictionAppliedIsExplicitInJSON(t *testing.T) {
	receipt := Receipt{TinyGoPredictedOperation: TinyGoOperationSubtract}
	apply := false
	receipt.TinyGoPredictionApplied = &apply
	encoded, err := json.Marshal(receipt)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(encoded), `"tiny_go_predicted_operation":"subtract"`) ||
		!strings.Contains(string(encoded), `"tiny_go_prediction_applied":false`) {
		t.Fatalf("receipt omitted the raw prediction or explicit false application: %s", encoded)
	}
}

func assertTinyGoPredictionApplied(t *testing.T, receipt Receipt, operation string, applied bool) {
	t.Helper()
	if receipt.TinyGoPredictedOperation != operation || receipt.TinyGoPredictionApplied == nil ||
		*receipt.TinyGoPredictionApplied != applied {
		t.Fatalf("prediction operation/application = %q/%v, want %q/%t", receipt.TinyGoPredictedOperation,
			receipt.TinyGoPredictionApplied, operation, applied)
	}
}

func TestTinyGoProviderFailsClosedOnUnsupportedRuntimeLabel(t *testing.T) {
	worker := &TinyGoWorker{model: &fakeTinyGoRuntime{label: "emit_source"}}
	request := tinyGoRequest("add these values", TinyGoOperationAdd, TinyGoOperationSubtract)
	receipt, err := worker.Resolve(context.Background(), request)
	if err == nil || receipt.Schema != "" {
		t.Fatalf("unsupported runtime label should fail closed: receipt=%+v err=%v", receipt, err)
	}
}

func TestTinyGoProviderRejectsInvalidMappingsBeforeInference(t *testing.T) {
	cases := []struct {
		name   string
		change func(*Request)
	}{
		{name: "unsupported", change: func(request *Request) { request.Question.Options[0].Operation = "emit_source" }},
		{name: "missing", change: func(request *Request) { request.Question.Options[1].Operation = "" }},
		{name: "duplicate operation", change: func(request *Request) { request.Question.Options[1].Operation = request.Question.Options[0].Operation }},
		{name: "duplicate option ID", change: func(request *Request) { request.Question.Options[1].ID = request.Question.Options[0].ID }},
		{name: "Laya model selector", change: func(request *Request) { request.ProviderModel = "english" }},
		{name: "empty intent", change: func(request *Request) { request.Intent = "  " }},
		{name: "too long", change: func(request *Request) { request.Intent = strings.Repeat("x", decision.InputMaxBytes+1) }},
		{name: "invalid UTF-8", change: func(request *Request) { request.Intent = string([]byte{0xff}) }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fake := &fakeTinyGoRuntime{label: TinyGoOperationAdd}
			worker := &TinyGoWorker{model: fake}
			request := tinyGoRequest("add the amount to the balance", TinyGoOperationAdd, TinyGoOperationSubtract)
			tc.change(&request)
			_, err := worker.Resolve(context.Background(), request)
			if err == nil {
				t.Fatal("invalid mapping/request was accepted")
			}
			if fake.calls != 0 {
				t.Fatalf("inference ran %d times for an invalid request", fake.calls)
			}
		})
	}
}

func TestTinyGoRejectsOptionOverflowBeforeFullValidation(t *testing.T) {
	fake := &fakeTinyGoRuntime{label: TinyGoOperationAdd}
	worker := &TinyGoWorker{model: fake}
	request := tinyGoRequest("valid intent", TinyGoOperationAdd, TinyGoOperationSubtract)
	request.Schema = "invalid-schema"
	labels := decision.Labels()
	for index := len(request.Question.Options); index <= decision.LabelCount; index++ {
		request.Question.Options = append(request.Question.Options, Option{
			ID: "candidate-" + string(rune('a'+index)), Description: "candidate",
			Operation: labels[index%decision.LabelCount],
		})
	}
	_, err := worker.Resolve(context.Background(), request)
	if !errors.Is(err, ErrTinyGoInvalidRequest) || !strings.Contains(err.Error(), "at most 8") {
		t.Fatalf("overflow error=%v, want early eight-option cap", err)
	}
	if fake.calls != 0 {
		t.Fatalf("inference ran %d times for an oversized option table", fake.calls)
	}
}

func TestTinyGoOperationIndexMatchesSDKLabelOrder(t *testing.T) {
	for expectedIndex, label := range decision.Labels() {
		index, ok := tinyGoOperationIndex(label)
		if !ok || index != expectedIndex {
			t.Fatalf("operation %q index=%d ok=%t, want %d", label, index, ok, expectedIndex)
		}
	}
	if _, ok := tinyGoOperationIndex("emit_source"); ok {
		t.Fatal("unsupported operation was assigned a fixed slot")
	}
}

func TestTinyGoResolveRejectsNilProviderAndWorker(t *testing.T) {
	var provider *TinyGoProvider
	request := tinyGoRequest("add these values", TinyGoOperationAdd, TinyGoOperationSubtract)
	if _, err := provider.NewWorker(); err == nil {
		t.Fatal("nil provider created a worker")
	}
	if _, err := provider.Resolve(context.Background(), request); err == nil {
		t.Fatal("nil provider resolved a request")
	}
	worker := &TinyGoWorker{}
	if _, err := worker.Resolve(context.Background(), request); err == nil {
		t.Fatal("nil-model worker resolved a request")
	}
}

func TestTinyGoWorkerPassesIntentOnlyAndOwnsReusableWorkspace(t *testing.T) {
	fake := &fakeTinyGoRuntime{label: TinyGoOperationAdd}
	provider := &TinyGoProvider{model: fake}
	worker, err := provider.NewWorker()
	if err != nil {
		t.Fatal(err)
	}
	otherWorker, err := provider.NewWorker()
	if err != nil {
		t.Fatal(err)
	}
	if &worker.workspace == &otherWorker.workspace {
		t.Fatal("concurrent workers share inference scratch")
	}
	request := tinyGoRequest("intent only: add amount to balance", TinyGoOperationAdd, TinyGoOperationSubtract)
	request.State = "secret state, hidden options, training failures, and holdout rows"
	request.Question.Instructions = "hidden prompt content"
	request.Question.Options[0].Description = "hidden candidate expression: amount + balance"
	request.Question.Options[1].Description = "hidden candidate expression: amount - balance"
	firstWorkspace := &worker.workspace
	receipt, err := worker.Resolve(context.Background(), request)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if fake.input != request.Intent || strings.Contains(fake.input, "hidden") ||
		strings.Contains(fake.input, "training") || strings.Contains(fake.input, "holdout") {
		t.Fatalf("tiny_go received text other than the explicit intent: %q", fake.input)
	}
	if &worker.workspace != firstWorkspace || receipt.Selected != "candidate-add" {
		t.Fatalf("worker did not retain its private workspace or map result: %+v", receipt)
	}
}

func TestTinyGoWorkerChecksCancellationBeforeAndAfterSynchronousInference(t *testing.T) {
	t.Run("provider rejects already canceled before allocating a worker", func(t *testing.T) {
		fake := &fakeTinyGoRuntime{label: TinyGoOperationAdd}
		provider := &TinyGoProvider{model: fake}
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		_, err := provider.Resolve(ctx, tinyGoRequest("add amount", TinyGoOperationAdd, TinyGoOperationSubtract))
		if !errors.Is(err, context.Canceled) || fake.calls != 0 {
			t.Fatalf("canceled request should not infer: err=%v calls=%d", err, fake.calls)
		}
	})

	t.Run("already canceled", func(t *testing.T) {
		fake := &fakeTinyGoRuntime{label: TinyGoOperationAdd}
		worker := &TinyGoWorker{model: fake}
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		_, err := worker.Resolve(ctx, tinyGoRequest("add amount", TinyGoOperationAdd, TinyGoOperationSubtract))
		if !errors.Is(err, context.Canceled) || fake.calls != 0 {
			t.Fatalf("canceled request should not infer: err=%v calls=%d", err, fake.calls)
		}
	})

	t.Run("canceled during synchronous inference", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		fake := &fakeTinyGoRuntime{label: TinyGoOperationAdd, afterPredict: cancel}
		worker := &TinyGoWorker{model: fake}
		receipt, err := worker.Resolve(ctx, tinyGoRequest("add amount", TinyGoOperationAdd, TinyGoOperationSubtract))
		if !errors.Is(err, context.Canceled) || fake.calls != 1 || receipt.Schema != "" {
			t.Fatalf("post-inference cancellation was not returned without a receipt: receipt=%+v err=%v calls=%d", receipt, err, fake.calls)
		}
	})
}

func TestTinyGoLoadErrorDoesNotExposeCallerFilesystemPath(t *testing.T) {
	secretPath := filepath.Join(t.TempDir(), "private-customer-path", "missing-model.json")
	_, err := LoadTinyGoProvider(secretPath)
	if !errors.Is(err, ErrTinyGoModelLoad) || strings.Contains(err.Error(), secretPath) || strings.Contains(err.Error(), "private-customer") {
		t.Fatalf("model load error exposed caller path: %v", err)
	}
}

func TestTinyGoIntentAndOperationFieldsAreDigestBoundButOmittedWhenEmpty(t *testing.T) {
	request := tinyGoRequest("", "", "")
	legacyDigest, err := Validate(request)
	if err != nil {
		t.Fatal(err)
	}
	request.Intent = "add this"
	request.Question.Options[0].Operation = TinyGoOperationAdd
	request.Question.Options[1].Operation = TinyGoOperationSubtract
	newDigest, err := Validate(request)
	if err != nil {
		t.Fatal(err)
	}
	if legacyDigest == newDigest {
		t.Fatal("explicit intent and operation mappings did not bind into request digest")
	}
}

type fakeTinyGoRuntime struct {
	label        string
	input        string
	calls        int
	afterPredict func()
}

func (fake *fakeTinyGoRuntime) Variant() string        { return "fp32" }
func (fake *fakeTinyGoRuntime) WeightsSHA256() string  { return strings.Repeat("a", 64) }
func (fake *fakeTinyGoRuntime) MetadataSHA256() string { return strings.Repeat("b", 64) }
func (fake *fakeTinyGoRuntime) PredictInto(input string, _ *decision.Workspace, output *decision.Prediction) error {
	fake.calls++
	fake.input = input
	output.Probabilities[0] = 1
	output.Confidence = 1
	if fake.afterPredict != nil {
		fake.afterPredict()
	}
	return nil
}
func (fake *fakeTinyGoRuntime) PredictLabel(_ *decision.Prediction) string { return fake.label }

func tinyGoRequest(intent string, firstOperation, secondOperation string) Request {
	return Request{
		Schema: RequestSchema, State: "state remains local",
		Intent: intent,
		Question: Question{ID: "operation", Instructions: "Choose a typed operation.", Options: []Option{
			{ID: "candidate-add", Description: "candidate A", Operation: firstOperation},
			{ID: "candidate-subtract", Description: "candidate B", Operation: secondOperation},
		}},
		Fallback: "candidate-add",
	}
}

func tinyGoOperations() []string {
	labels := decision.Labels()
	operations := make([]string, len(labels))
	copy(operations, labels[:])
	return operations
}

func loadSyntheticTinyGoProvider(t *testing.T, winner string, threshold float64) *TinyGoProvider {
	t.Helper()
	metadataPath := writeSyntheticTinyGoBundle(t, winner, threshold)
	provider, err := LoadTinyGoProvider(metadataPath)
	if err != nil {
		t.Fatalf("LoadTinyGoProvider: %v", err)
	}
	return provider
}

func writeSyntheticTinyGoBundle(t *testing.T, winner string, threshold float64) string {
	t.Helper()
	dir := t.TempDir()
	labels := decision.Labels()
	metadata := decision.Metadata{
		Schema: decision.MetadataSchema, Variant: "fp32",
		FeatureDim: decision.FeatureDim, HiddenDim: decision.HiddenDim,
		MaxBytes: decision.InputMaxBytes, Labels: labels[:],
		Temperature: 1, WeightsFile: "weights.bin", ConfidenceThreshold: &threshold,
	}
	var weights []byte
	for _, tensor := range []struct {
		name  string
		rows  int
		cols  int
		count int
	}{
		{name: "w1", rows: decision.HiddenDim, cols: decision.FeatureDim, count: decision.HiddenDim * decision.FeatureDim},
		{name: "b1", rows: 1, cols: decision.HiddenDim, count: decision.HiddenDim},
		{name: "w2", rows: decision.LabelCount, cols: decision.HiddenDim, count: decision.LabelCount * decision.HiddenDim},
		{name: "b2", rows: 1, cols: decision.LabelCount, count: decision.LabelCount},
	} {
		raw := make([]byte, tensor.count*4)
		if tensor.name == "b2" && winner != "" {
			index := -1
			for i, label := range labels {
				if label == winner {
					index = i
					break
				}
			}
			if index < 0 {
				t.Fatalf("unknown synthetic winner %q", winner)
			}
			binary.LittleEndian.PutUint32(raw[index*4:], math.Float32bits(10))
		}
		metadata.Tensors = append(metadata.Tensors, decision.TensorMetadata{
			Name: tensor.name, Count: tensor.count, Rows: tensor.rows, Cols: tensor.cols,
			Encoding: "float32_le", Offset: int64(len(weights)), Bytes: int64(len(raw)), Scale: 1,
		})
		weights = append(weights, raw...)
	}
	digest := sha256.Sum256(weights)
	metadata.WeightsSHA256 = hex.EncodeToString(digest[:])
	metadataRaw, err := json.Marshal(metadata)
	if err != nil {
		t.Fatalf("marshal synthetic model metadata: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, metadata.WeightsFile), weights, 0o600); err != nil {
		t.Fatalf("write synthetic weights: %v", err)
	}
	metadataPath := filepath.Join(dir, "model.json")
	if err := os.WriteFile(metadataPath, metadataRaw, 0o600); err != nil {
		t.Fatalf("write synthetic metadata: %v", err)
	}
	return metadataPath
}
