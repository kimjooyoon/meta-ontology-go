package decisionroute

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	decision "github.com/kimjooyoon/gooo-decision-runtime"
)

const (
	ProviderTinyGo = "tiny_go"

	TinyGoOperationAdd       = "add"
	TinyGoOperationSubtract  = "subtract"
	TinyGoOperationMultiply  = "multiply"
	TinyGoOperationLessThan  = "less_than"
	TinyGoOperationLessEqual = "less_equal"
	TinyGoOperationEqual     = "equal"
	TinyGoOperationAnd       = "and"
	TinyGoOperationOr        = "or"

	TinyGoFallbackLowConfidence       = "TINY_GO_LOW_CONFIDENCE"
	TinyGoFallbackOperationNotOffered = "TINY_GO_OPERATION_NOT_OFFERED"
)

var (
	ErrTinyGoModelLoad        = errors.New("tiny_go model bundle is invalid or inaccessible")
	ErrTinyGoInvalidRequest   = errors.New("tiny_go request intent or operation mapping is invalid")
	ErrTinyGoProviderModelSet = errors.New("tiny_go does not accept a Laya provider_model selector")
)

type tinyGoRuntime interface {
	Variant() string
	WeightsSHA256() string
	MetadataSHA256() string
	PredictInto(string, *decision.Workspace, *decision.Prediction) error
	PredictLabel(*decision.Prediction) string
}

// TinyGoProvider owns one immutable loaded model. Share it between goroutines,
// and give each concurrent worker its own workspace.
type TinyGoProvider struct {
	model tinyGoRuntime
}

// TinyGoWorker owns scratch arrays for one caller or concurrent worker.
// Do not call Resolve concurrently on the same worker.
type TinyGoWorker struct {
	model     tinyGoRuntime
	workspace decision.Workspace
}

// LoadTinyGoProvider loads one pinned model bundle. SDK errors may contain
// filesystem paths, so the adapter returns a path-free error at this boundary.
func LoadTinyGoProvider(metadataPath string) (*TinyGoProvider, error) {
	model, err := decision.Load(metadataPath)
	if err != nil {
		return nil, ErrTinyGoModelLoad
	}
	return &TinyGoProvider{model: model}, nil
}

// NewWorker allocates independent inference scratch for a concurrent caller.
func (provider *TinyGoProvider) NewWorker() (*TinyGoWorker, error) {
	if provider == nil || provider.model == nil {
		return nil, errors.New("tiny_go provider is not loaded")
	}
	return &TinyGoWorker{model: provider.model}, nil
}

// Resolve uses a fresh per-call workspace. For repeated calls, callers can
// reuse NewWorker and its private workspace instead.
func (provider *TinyGoProvider) Resolve(ctx context.Context, request Request) (Receipt, error) {
	if ctx == nil {
		return Receipt{}, errors.New("tiny_go context is required")
	}
	if err := ctx.Err(); err != nil {
		return Receipt{}, err
	}
	worker, err := provider.NewWorker()
	if err != nil {
		return Receipt{}, err
	}
	return worker.Resolve(ctx, request)
}

// Resolve classifies only Request.Intent. State, question prose, option
// descriptions, and any embedded candidate or training data are not passed to
// the model. The closed operation label is translated to a declared option ID.
func (worker *TinyGoWorker) Resolve(ctx context.Context, request Request) (Receipt, error) {
	if err := worker.validateResolveStart(ctx, request); err != nil {
		return Receipt{}, err
	}
	digest, operationIDs, err := validateTinyGoRequest(request)
	if err != nil {
		return Receipt{}, err
	}
	prediction, err := worker.predict(ctx, request.Intent)
	if err != nil {
		return Receipt{}, err
	}
	return worker.receipt(request, digest, operationIDs, &prediction)
}

func (worker *TinyGoWorker) validateResolveStart(ctx context.Context, request Request) error {
	if ctx == nil {
		return errors.New("tiny_go context is required")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if worker == nil || worker.model == nil {
		return errors.New("tiny_go worker is not initialized")
	}
	if request.ProviderModel != "" {
		return ErrTinyGoProviderModelSet
	}
	return nil
}

func (worker *TinyGoWorker) predict(ctx context.Context, intent string) (decision.Prediction, error) {
	var prediction decision.Prediction
	if err := ctx.Err(); err != nil {
		return prediction, err
	}
	if err := worker.model.PredictInto(intent, &worker.workspace, &prediction); err != nil {
		return decision.Prediction{}, fmt.Errorf("tiny_go inference failed: %w", err)
	}
	if err := ctx.Err(); err != nil {
		return decision.Prediction{}, err
	}
	return prediction, nil
}

func (worker *TinyGoWorker) receipt(request Request, digest string, operationIDs [decision.LabelCount]string, prediction *decision.Prediction) (Receipt, error) {
	label := worker.model.PredictLabel(prediction)
	operationIndex, ok := tinyGoOperationIndex(label)
	if !ok {
		return Receipt{}, errors.New("tiny_go runtime returned an unsupported operation label")
	}
	variant, weightsSHA256 := worker.model.Variant(), worker.model.WeightsSHA256()
	metadataSHA256 := worker.model.MetadataSHA256()
	probabilities := make(map[string]float64, len(request.Question.Options))
	for index, optionID := range operationIDs {
		if optionID != "" {
			probabilities[optionID] = float64(prediction.Probabilities[index])
		}
	}
	confidence := float64(prediction.Confidence)
	receipt := Receipt{
		Schema: ReceiptSchema, Provider: ProviderTinyGo,
		TinyGoVariant: variant, TinyGoWeightsSHA256: weightsSHA256, TinyGoMetadataSHA256: metadataSHA256,
		RequestSHA256: digest, Probabilities: probabilities, Confidence: &confidence,
	}
	if prediction.Abstained {
		receipt.Mode = "deterministic_fallback"
		receipt.Selected = request.Fallback
		receipt.FallbackReason = TinyGoFallbackLowConfidence
		return receipt, nil
	}
	selected := operationIDs[operationIndex]
	if selected == "" {
		receipt.Mode = "deterministic_fallback"
		receipt.Selected = request.Fallback
		receipt.FallbackReason = TinyGoFallbackOperationNotOffered
		return receipt, nil
	}
	receipt.Mode = ProviderTinyGo
	receipt.Selected = selected
	return receipt, nil
}

func validateTinyGoRequest(request Request) (string, [decision.LabelCount]string, error) {
	var operationIDs [decision.LabelCount]string
	if len(request.Question.Options) > decision.LabelCount {
		return "", operationIDs, fmt.Errorf("%w: at most %d operation choices are supported", ErrTinyGoInvalidRequest, decision.LabelCount)
	}
	if strings.TrimSpace(request.Intent) == "" || len(request.Intent) > decision.InputMaxBytes ||
		!utf8.ValidString(request.Intent) {
		return "", operationIDs, ErrTinyGoInvalidRequest
	}
	digest, err := Validate(request)
	if err != nil {
		return "", operationIDs, fmt.Errorf("%w: invalid typed request", ErrTinyGoInvalidRequest)
	}
	for _, option := range request.Question.Options {
		index, ok := tinyGoOperationIndex(option.Operation)
		if !ok {
			return "", operationIDs, fmt.Errorf("%w: every option must map to one supported operation", ErrTinyGoInvalidRequest)
		}
		if operationIDs[index] != "" {
			return "", operationIDs, fmt.Errorf("%w: operation mappings must be one-to-one", ErrTinyGoInvalidRequest)
		}
		operationIDs[index] = option.ID
	}
	return digest, operationIDs, nil
}

func tinyGoOperationIndex(value string) (int, bool) {
	for index, operation := range decision.Labels() {
		if value == operation {
			return index, true
		}
	}
	return 0, false
}
