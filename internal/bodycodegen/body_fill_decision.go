package bodycodegen

import (
	"context"
	"fmt"
	"reflect"

	"github.com/kimjooyoon/meta-ontology-go/internal/decisionroute"
)

func bodyFillUsesTinyGo(options IRBodyFillOptions, provider tinyGoBodyFillResolver) bool {
	return provider != nil || options.recordedDecision != nil && options.recordedDecision.Provider == decisionroute.ProviderTinyGo
}

// A recorded decision is consumed only by source replay. It must bind the
// complete reconstructed request and choose one of its declared option IDs.
// Historical model activity remains an observation; no provider is called here.
func resolveBodyFillDecision(ctx context.Context, request decisionroute.Request, endpoint, apiKey string,
	options IRBodyFillOptions, provider tinyGoBodyFillResolver) (decisionroute.Receipt, error) {
	if len(request.Question.Options) == 1 {
		return resolveSingleBodyFill(ctx, request, options.recordedDecision)
	}
	if options.recordedDecision == nil {
		if provider != nil {
			return provider.Resolve(ctx, request)
		}
		return decisionroute.Resolve(ctx, request, endpoint, apiKey)
	}
	if err := ctx.Err(); err != nil {
		return decisionroute.Receipt{}, err
	}
	prior := *options.recordedDecision
	digest, err := decisionroute.Validate(request)
	if err != nil || prior.Schema != decisionroute.ReceiptSchema || prior.RequestSHA256 != digest {
		return decisionroute.Receipt{}, fmt.Errorf("body-fill recorded decision does not bind the reconstructed request")
	}
	for _, option := range request.Question.Options {
		if prior.Selected == option.ID {
			return prior, nil
		}
	}
	return decisionroute.Receipt{}, fmt.Errorf("body-fill recorded decision chose an undeclared option")
}

func resolveSingleBodyFill(ctx context.Context, request decisionroute.Request, prior *decisionroute.Receipt) (decisionroute.Receipt, error) {
	if err := ctx.Err(); err != nil {
		return decisionroute.Receipt{}, err
	}
	digest, err := decisionroute.ValidateSingleton(request)
	if err != nil {
		return decisionroute.Receipt{}, err
	}
	r := decisionroute.Receipt{Schema: decisionroute.ReceiptSchema, Mode: "deterministic_singleton",
		Provider: "deterministic", Selected: request.Fallback, FallbackReason: "ONLY_VALID_CANDIDATE", RequestSHA256: digest}
	if prior != nil && !reflect.DeepEqual(*prior, r) {
		return decisionroute.Receipt{}, fmt.Errorf("singleton body-fill decision differs from the reconstructed request")
	}
	return r, nil
}
