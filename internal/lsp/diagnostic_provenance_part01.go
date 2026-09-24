package lsp

import (
	"context"
	"encoding/json"
	"errors"
	"sort"

	"github.com/kimjooyoon/meta-ontology-go/internal/cache"
)

const diagnosticProvenanceSchema = "gooo/lsp-diagnostic-provenance/v1"

const (
	diagnosticProvenanceClosed  = "CLOSED"
	diagnosticProvenanceUnknown = "UNKNOWN"
)

type diagnosticProvenanceParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
}

type diagnosticProvenanceItem struct {
	Range                    Range              `json:"range"`
	Severity                 DiagnosticSeverity `json:"severity,omitempty"`
	Code                     string             `json:"code,omitempty"`
	Source                   string             `json:"source,omitempty"`
	Message                  string             `json:"message"`
	OriginDigest             string             `json:"origin_digest"`
	DocumentProvenanceDigest string             `json:"document_provenance_digest"`
}

type diagnosticProvenanceObservation struct {
	Schema                   string                     `json:"schema"`
	URI                      string                     `json:"uri"`
	SourceDigest             string                     `json:"source_digest,omitempty"`
	SemanticDigest           string                     `json:"semantic_digest,omitempty"`
	ProfileDigest            string                     `json:"profile_digest,omitempty"`
	ToolchainDigest          string                     `json:"toolchain_digest,omitempty"`
	ContractDigest           string                     `json:"contract_digest,omitempty"`
	DocumentProvenanceDigest string                     `json:"document_provenance_digest,omitempty"`
	Diagnostics              []diagnosticProvenanceItem `json:"diagnostics"`
	DiagnosticMapDigest      string                     `json:"diagnostic_map_digest"`
	Decision                 string                     `json:"decision"`
	Reason                   string                     `json:"reason"`
	NonAuthorizing           bool                       `json:"non_authorizing"`
	ObservationDigest        string                     `json:"observation_digest"`
}

func (server *Server) diagnosticProvenanceRequest(ctx context.Context, request requestEnvelope) (*responseEnvelope, [][]byte, error) {
	var params diagnosticProvenanceParams
	if decodeParams(request.Params, &params) != nil || params.TextDocument.URI == "" {
		return responseOrNil(request.ID, invalidParams, "Invalid diagnostic provenance parameters"), nil, nil
	}
	if err := server.refresh(ctx, params.TextDocument.URI); err != nil {
		return featureErrorResponse(request.ID, err, ctx)
	}
	server.mu.RLock()
	stored, exists := server.documents[params.TextDocument.URI]
	if exists {
		key := stored.cacheKey
		value := documentCopy(stored)
		server.mu.RUnlock()
		return resultResponse(request.ID, observeDiagnosticProvenance(params.TextDocument.URI, value, key)), nil, nil
	}
	server.mu.RUnlock()
	return resultResponse(request.ID, nil), nil, nil
}

func observeDiagnosticProvenance(uri string, value document, key documentCacheKey) diagnosticProvenanceObservation {
	observation := diagnosticProvenanceObservation{
		Schema:         diagnosticProvenanceSchema,
		URI:            uri,
		Diagnostics:    []diagnosticProvenanceItem{},
		Decision:       diagnosticProvenanceUnknown,
		Reason:         "MISSING_SOURCE_DIGEST",
		NonAuthorizing: true,

		SourceDigest:    key.sourceDigest,
		SemanticDigest:  value.result.semanticDigest,
		ProfileDigest:   key.profileDigest,
		ToolchainDigest: key.toolchainDigest,
		ContractDigest:  key.contractDigest}
	if value.result.semanticChecked && !value.result.semanticValid {
		observation.Reason = "SEMANTIC_INVALID"
		return finalizeDiagnosticProvenance(observation)
	}
	for _, required := range []struct {
		name  string
		value string
	}{
		{name: "SOURCE_DIGEST", value: observation.SourceDigest},
		{name: "SEMANTIC_DIGEST", value: observation.SemanticDigest},
		{name: "PROFILE_DIGEST", value: observation.ProfileDigest},
		{name: "TOOLCHAIN_DIGEST", value: observation.ToolchainDigest},
		{name: "CONTRACT_DIGEST", value: observation.ContractDigest},
	} {
		if !cache.Digest(required.value).Known() {
			observation.Reason = "MISSING_" + required.name
			return finalizeDiagnosticProvenance(observation)
		}
	}
	observation.DocumentProvenanceDigest = diagnosticDocumentProvenanceDigest(value, key)
	observation.Diagnostics = diagnosticProvenanceItems(value.result.Diagnostics, observation.DocumentProvenanceDigest)
	observation.Decision = diagnosticProvenanceClosed
	observation.Reason = "DIAGNOSTIC_SURFACE_BOUND"
	return finalizeDiagnosticProvenance(observation)
}

func diagnosticDocumentProvenanceDigest(value document, key documentCacheKey) string {
	provenance := documentProvenance{
		SourceDigest:    key.sourceDigest,
		SemanticDigest:  value.result.semanticDigest,
		ProfileDigest:   key.profileDigest,
		ToolchainDigest: key.toolchainDigest,
		ContractDigest:  key.contractDigest,
		Symbols:         documentProvenanceSymbols(value.result),
		References:      documentProvenanceReferences(value.result),
	}
	provenance.SymbolMapDigest = documentProvenanceSymbolMapDigest(provenance.Symbols)
	provenance.ReferenceMapDigest = documentProvenanceReferenceMapDigest(provenance.References)
	return documentProvenanceDigest(provenance)
}

func diagnosticProvenanceItems(values []Diagnostic, documentProvenanceDigest string) []diagnosticProvenanceItem {
	result := make([]diagnosticProvenanceItem, 0, len(values))
	for _, value := range values {
		item := diagnosticProvenanceItem{
			Range:    value.Range,
			Severity: value.Severity,
			Code:     value.Code,
			Source:   value.Source,
			Message:  value.Message,
			DocumentProvenanceDigest: documentProvenanceDigest,
		}
		item.OriginDigest = diagnosticProvenanceItemDigest(item)
		result = append(result, item)
	}
	sort.SliceStable(result, func(left, right int) bool {
		first, second := result[left], result[right]
		if first.Range.Start != second.Range.Start {
			return positionLess(first.Range.Start, second.Range.Start)
		}
		if first.Range.End != second.Range.End {
			return positionLess(first.Range.End, second.Range.End)
		}
		if first.Severity != second.Severity {
			return first.Severity < second.Severity
		}
		if first.Code != second.Code {
			return first.Code < second.Code
		}
		if first.Source != second.Source {
			return first.Source < second.Source
		}
		return first.Message < second.Message
	})
	return result
}

func diagnosticProvenanceItemDigest(value diagnosticProvenanceItem) string {
	value.OriginDigest = ""
	payload, _ := json.Marshal(value)
	return cache.HashBytes(payload).String()
}

func diagnosticProvenanceMapDigest(values []diagnosticProvenanceItem) string {
	payload, _ := json.Marshal(values)
	return cache.HashBytes(payload).String()
}

func finalizeDiagnosticProvenance(value diagnosticProvenanceObservation) diagnosticProvenanceObservation {
	value.DiagnosticMapDigest = diagnosticProvenanceMapDigest(value.Diagnostics)
	value.ObservationDigest = diagnosticProvenanceObservationDigest(value)
	return value
}

func diagnosticProvenanceObservationDigest(value diagnosticProvenanceObservation) string {
	value.ObservationDigest = ""
	payload, _ := json.Marshal(value)
	return cache.HashBytes(payload).String()
}

func validateDiagnosticProvenance(value diagnosticProvenanceObservation) error {
	if value.Schema != diagnosticProvenanceSchema || value.URI == "" || !value.NonAuthorizing || value.Reason == "" {
		return errors.New("diagnostic provenance identity is invalid")
	}
	if value.Decision != diagnosticProvenanceClosed && value.Decision != diagnosticProvenanceUnknown {
		return errors.New("diagnostic provenance decision is invalid")
	}
	for _, digest := range []string{value.SourceDigest, value.SemanticDigest, value.ProfileDigest, value.ToolchainDigest, value.ContractDigest, value.DocumentProvenanceDigest} {
		if digest != "" && !cache.Digest(digest).Known() {
			return errors.New("diagnostic provenance binding is invalid")
		}
	}
	if !cache.Digest(value.DiagnosticMapDigest).Known() || value.DiagnosticMapDigest != diagnosticProvenanceMapDigest(value.Diagnostics) {
		return errors.New("diagnostic provenance map is invalid")
	}
	for _, item := range value.Diagnostics {
		if !cache.Digest(item.OriginDigest).Known() || item.OriginDigest != diagnosticProvenanceItemDigest(item) || !cache.Digest(item.DocumentProvenanceDigest).Known() || item.DocumentProvenanceDigest != value.DocumentProvenanceDigest {
			return errors.New("diagnostic provenance item origin is invalid")
		}
	}
	if value.Decision == diagnosticProvenanceClosed {
		if !cache.Digest(value.SourceDigest).Known() || !cache.Digest(value.SemanticDigest).Known() ||
			!cache.Digest(value.ProfileDigest).Known() || !cache.Digest(value.ToolchainDigest).Known() ||
			!cache.Digest(value.ContractDigest).Known() || !cache.Digest(value.DocumentProvenanceDigest).Known() {
			return errors.New("diagnostic provenance closed binding is incomplete")
		}
	}
	if !cache.Digest(value.ObservationDigest).Known() || value.ObservationDigest != diagnosticProvenanceObservationDigest(value) {
		return errors.New("diagnostic provenance observation is invalid")
	}
	return nil
}
