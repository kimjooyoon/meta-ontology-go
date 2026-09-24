package lsp

import (
	"context"
	"encoding/json"

	"github.com/kimjooyoon/meta-ontology-go/internal/provenance"
)

const ExecutionPlanProvenanceSchemaPart01 = provenance.ExecutionPlanProvenanceBindingSchemaPart01

type ExecutionPlanProvenanceParamsPart01 struct {
	TextDocument    TextDocumentIdentifier                  `json:"textDocument"`
	Task            string                                  `json:"task"`
	WorkspaceDigest string                                  `json:"workspace_digest"`
	Model           string                                  `json:"model"`
	GatewayPolicy   provenance.GatewayPolicy                `json:"gateway_policy"`
	Lifecycle       provenance.ExecutionPlanLifecyclePart01 `json:"lifecycle"`
}

func (server *Server) executionPlanProvenanceRequest(
	ctx context.Context,
	request requestEnvelope,
) (*responseEnvelope, [][]byte, error) {
	var params ExecutionPlanProvenanceParamsPart01
	if decodeParams(request.Params, &params) != nil || params.TextDocument.URI == "" {
		return responseOrNil(request.ID, invalidParams, "Invalid execution-plan provenance parameters"), nil, nil
	}
	if err := server.refresh(ctx, params.TextDocument.URI); err != nil {
		return featureErrorResponse(request.ID, err, ctx)
	}

	var chain provenance.SelfImprovementProvenanceChainPart01
	server.mu.RLock()
	stored, exists := server.documents[params.TextDocument.URI]
	if exists {
		key := stored.cacheKey
		result := cloneParseResult(stored.result)
		stored = &document{
			version:  stored.version,
			text:     stored.text,
			cacheKey: key,
			result:   result,
		}
	}
	server.mu.RUnlock()

	if exists && stored.cacheKey.sourceDigest != "" {
		symbols := documentProvenanceSymbols(stored.result)
		references := documentProvenanceReferences(stored.result)
		chain = provenance.BuildSelfImprovementProvenanceChainPart01(
			documentProvenanceSymbolMapDigest(symbols),
			stored.cacheKey.sourceDigest,
			stored.result.semanticDigest,
			documentProvenanceReferenceMapDigest(references),
			"",
			"",
		)
	}

	binding := provenance.BindExecutionPlanToProvenancePart01(
		params.Task,
		params.WorkspaceDigest,
		params.Model,
		params.GatewayPolicy,
		params.Lifecycle,
		chain,
	)
	if err := binding.Validate(); err != nil {
		return featureErrorResponse(request.ID, err, ctx)
	}
	return resultResponse(request.ID, binding), nil, nil
}
