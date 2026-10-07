package bodycodegen

import (
	"context"
	"fmt"

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
