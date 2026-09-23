package lsp

import (
	"context"
	"encoding/json"
)

const documentProvenanceSchema = "gooo/lsp-document-provenance/v1"

type documentProvenance struct {
	Schema         string `json:"schema"`
	URI            string `json:"uri"`
	SourceDigest   string `json:"source_digest"`
	ProfileDigest  string `json:"profile_digest"`
	ToolchainDigest string `json:"toolchain_digest"`
	ContractDigest string `json:"contract_digest"`
}

func (server *Server) documentProvenanceRequest(ctx context.Context, request requestEnvelope) (*responseEnvelope, [][]byte, error) {
	var params DocumentProvenanceParams
	if decodeParams(request.Params, &params) != nil || params.TextDocument.URI == "" {
		return responseOrNil(request.ID, invalidParams, "Invalid document provenance parameters"), nil, nil
	}
	if err := server.refresh(ctx, params.TextDocument.URI); err != nil {
		return featureErrorResponse(request.ID, err, ctx)
	}
	server.mu.RLock()
	stored, exists := server.documents[params.TextDocument.URI]
	if exists {
		key := stored.cacheKey
		stored = &document{version: stored.version, text: stored.text, cacheKey: key}
	}
	server.mu.RUnlock()
	if !exists || stored.cacheKey.sourceDigest == "" {
		return resultResponse(request.ID, nil), nil, nil
	}
	return resultResponse(request.ID, documentProvenance{
		Schema: documentProvenanceSchema, URI: params.TextDocument.URI,
		SourceDigest: stored.cacheKey.sourceDigest, ProfileDigest: stored.cacheKey.profileDigest,
		ToolchainDigest: stored.cacheKey.toolchainDigest, ContractDigest: stored.cacheKey.contractDigest,
	}), nil, nil
}

func decodeDocumentProvenance(payload json.RawMessage) (documentProvenance, error) {
	var value documentProvenance
	if err := json.Unmarshal(payload, &value); err != nil {
		return documentProvenance{}, err
	}
	return value, nil
}
