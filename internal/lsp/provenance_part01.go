package lsp

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/analysisprovenance"
)

const documentProvenanceSchema = "gooo/lsp-document-provenance/v3"

type documentProvenance struct {
	Schema             string                        `json:"schema"`
	URI                string                        `json:"uri"`
	SubjectDigest      string                        `json:"subject_digest"`
	SourceDigest       string                        `json:"source_digest"`
	SemanticDigest     string                        `json:"semantic_digest"`
	ProfileDigest      string                        `json:"profile_digest"`
	ToolchainDigest    string                        `json:"toolchain_digest"`
	ContractDigest     string                        `json:"contract_digest"`
	SymbolMapDigest    string                        `json:"symbol_map_digest"`
	ReferenceMapDigest string                        `json:"reference_map_digest"`
	Symbols            []documentProvenanceSymbol    `json:"symbols"`
	References         []documentProvenanceReference `json:"references"`
	ProvenanceDigest   string                        `json:"provenance_digest"`
}

func documentProvenanceDigest(value documentProvenance) string {
	return analysisprovenance.DocumentDigestWithSymbolMapAndReferences(value.SourceDigest, value.SemanticDigest, value.ProfileDigest, value.ToolchainDigest, value.ContractDigest, value.SymbolMapDigest, value.ReferenceMapDigest)
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
		result := cloneParseResult(stored.result)
		stored = &document{version: stored.version, text: stored.text, cacheKey: key, result: result}
	}
	server.mu.RUnlock()
	if !exists || stored.cacheKey.sourceDigest == "" {
		return resultResponse(request.ID, nil), nil, nil
	}
	provenance := documentProvenance{
		Schema: documentProvenanceSchema, URI: params.TextDocument.URI,
		SubjectDigest: stored.cacheKey.sourceDigest,
		SourceDigest:  stored.cacheKey.sourceDigest, SemanticDigest: stored.result.semanticDigest, ProfileDigest: stored.cacheKey.profileDigest,
		ToolchainDigest: stored.cacheKey.toolchainDigest, ContractDigest: stored.cacheKey.contractDigest,
		Symbols:    documentProvenanceSymbols(stored.result),
		References: documentProvenanceReferences(stored.result),
	}
	provenance.SymbolMapDigest = documentProvenanceSymbolMapDigest(provenance.Symbols)
	provenance.ReferenceMapDigest = documentProvenanceReferenceMapDigest(provenance.References)
	provenance.ProvenanceDigest = documentProvenanceDigest(provenance)
	return resultResponse(request.ID, provenance), nil, nil
}

func decodeDocumentProvenance(payload json.RawMessage) (documentProvenance, error) {
	var value documentProvenance
	if err := json.Unmarshal(payload, &value); err != nil {
		return documentProvenance{}, err
	}
	if err := validateDocumentProvenance(value); err != nil {
		return documentProvenance{}, err
	}
	return value, nil
}

func validateDocumentProvenance(value documentProvenance) error {
	if value.Schema != documentProvenanceSchema || value.URI == "" ||
		!knownLSPProvenanceDigest(value.SubjectDigest) || value.SubjectDigest != value.SourceDigest ||
		!knownLSPProvenanceDigest(value.SourceDigest) || !knownLSPProvenanceDigest(value.SemanticDigest) || !knownLSPProvenanceDigest(value.ProfileDigest) ||
		!knownLSPProvenanceDigest(value.ToolchainDigest) || !knownLSPProvenanceDigest(value.ContractDigest) ||
		!knownLSPProvenanceDigest(value.SymbolMapDigest) || value.SymbolMapDigest != documentProvenanceSymbolMapDigest(value.Symbols) ||
		!knownLSPProvenanceDigest(value.ReferenceMapDigest) || value.ReferenceMapDigest != documentProvenanceReferenceMapDigest(value.References) ||
		!knownLSPProvenanceDigest(value.ProvenanceDigest) || value.ProvenanceDigest != documentProvenanceDigest(value) {
		return errors.New("document provenance identity is invalid")
	}
	if err := validateDocumentProvenanceSymbols(value.Symbols); err != nil {
		return err
	}
	if err := validateDocumentProvenanceReferences(value.References); err != nil {
		return err
	}
	return nil
}
