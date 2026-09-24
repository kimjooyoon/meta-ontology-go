package lsp

import (
	"context"

	"github.com/kimjooyoon/meta-ontology-go/internal/provenance"
)

const SelfImprovementProvenanceChainSchemaPart01 = "gooo/lsp-self-improvement-provenance-chain/v1"

type SelfImprovementProvenanceChainResponsePart01 struct {
	Schema string `json:"schema"`
	SelfImprovementProvenanceChainProjectionPart01
}

func (value SelfImprovementProvenanceChainResponsePart01) Validate() error {
	if value.Schema != SelfImprovementProvenanceChainSchemaPart01 {
		return invalidSelfImprovementProvenanceChainResponsePart01("schema does not match")
	}
	return value.SelfImprovementProvenanceChainProjectionPart01.Validate()
}

func (server *Server) selfImprovementProvenanceChainRequest(
	ctx context.Context,
	request requestEnvelope,
) (*responseEnvelope, [][]byte, error) {
	var params DocumentProvenanceParams
	if decodeParams(request.Params, &params) != nil || params.TextDocument.URI == "" {
		return responseOrNil(request.ID, invalidParams, "Invalid self-improvement provenance parameters"), nil, nil
	}
	if err := server.refresh(ctx, params.TextDocument.URI); err != nil {
		return featureErrorResponse(request.ID, err, ctx)
	}
	server.mu.RLock()
	stored, exists := server.documents[params.TextDocument.URI]
	if exists {
		key := stored.cacheKey
		result := cloneParseResult(stored.result)
		stored = &document{
			version: stored.version, text: stored.text, cacheKey: key, result: result,
		}
	}
	server.mu.RUnlock()
	if !exists || stored.cacheKey.sourceDigest == "" {
		return resultResponse(request.ID, nil), nil, nil
	}

	symbols := documentProvenanceSymbols(stored.result)
	references := documentProvenanceReferences(stored.result)
	chain := provenance.BuildSelfImprovementProvenanceChainPart01(
		documentProvenanceSymbolMapDigest(symbols),
		stored.cacheKey.sourceDigest,
		stored.result.semanticDigest,
		documentProvenanceReferenceMapDigest(references),
		"",
		"",
	)
	projection, err := ProjectSelfImprovementProvenanceChainPart01(
		params.TextDocument.URI,
		stored.version,
		chain,
	)
	if err != nil {
		return featureErrorResponse(request.ID, err, ctx)
	}
	response := SelfImprovementProvenanceChainResponsePart01{
		Schema: SelfImprovementProvenanceChainSchemaPart01,
		SelfImprovementProvenanceChainProjectionPart01: projection,
	}
	if err := response.Validate(); err != nil {
		return featureErrorResponse(request.ID, err, ctx)
	}
	return resultResponse(request.ID, response), nil, nil
}

type invalidSelfImprovementProvenanceChainResponsePart01 string

func (err invalidSelfImprovementProvenanceChainResponsePart01) Error() string {
	return "invalid self-improvement provenance chain response: " + string(err)
}
