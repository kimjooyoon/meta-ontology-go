package lsp

import (
	"context"
	"encoding/json"
	"errors"
	"sort"

	"github.com/kimjooyoon/meta-ontology-go/internal/cache"
)

const completionProvenanceSchema = "gooo/lsp-completion-provenance/v1"

const (
	completionProvenanceClosed  = "CLOSED"
	completionProvenanceUnknown = "UNKNOWN"
)

type CompletionProvenanceParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Position     Position               `json:"position"`
}

type completionProvenanceCandidate struct {
	Label         string `json:"label"`
	Kind          int    `json:"kind,omitempty"`
	Detail        string `json:"detail,omitempty"`
	Documentation string `json:"documentation,omitempty"`
	OriginDigest  string `json:"origin_digest"`
}

type completionProvenanceObservation struct {
	Schema                   string                          `json:"schema"`
	URI                      string                          `json:"uri"`
	Position                 Position                        `json:"position"`
	SourceDigest             string                          `json:"source_digest,omitempty"`
	SemanticDigest           string                          `json:"semantic_digest,omitempty"`
	ProfileDigest            string                          `json:"profile_digest,omitempty"`
	ToolchainDigest          string                          `json:"toolchain_digest,omitempty"`
	ContractDigest           string                          `json:"contract_digest,omitempty"`
	DocumentProvenanceDigest string                          `json:"document_provenance_digest,omitempty"`
	Candidates               []completionProvenanceCandidate `json:"candidates"`
	CandidateMapDigest       string                          `json:"candidate_map_digest"`
	Decision                 string                          `json:"decision"`
	Reason                   string                          `json:"reason"`
	NonAuthorizing           bool                            `json:"non_authorizing"`
	ObservationDigest        string                          `json:"observation_digest"`
}

func (server *Server) completionProvenanceRequest(ctx context.Context, request requestEnvelope) (*responseEnvelope, [][]byte, error) {
	var params CompletionProvenanceParams
	if decodeParams(request.Params, &params) != nil || params.TextDocument.URI == "" {
		return responseOrNil(request.ID, invalidParams, "Invalid completion provenance parameters"), nil, nil
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
		return resultResponse(request.ID, observeCompletionProvenance(server, params.TextDocument.URI, params.Position, value, key)), nil, nil
	}
	server.mu.RUnlock()
	return resultResponse(request.ID, nil), nil, nil
}

func observeCompletionProvenance(server *Server, uri string, position Position, value document, key documentCacheKey) completionProvenanceObservation {
	observation := completionProvenanceObservation{
		Schema: completionProvenanceSchema, URI: uri, Position: position,
		Candidates: []completionProvenanceCandidate{},
		Decision:   completionProvenanceUnknown, Reason: "MISSING_SOURCE_DIGEST",
		NonAuthorizing: true,

		SourceDigest:    key.sourceDigest,
		SemanticDigest:  value.result.semanticDigest,
		ProfileDigest:   key.profileDigest,
		ToolchainDigest: key.toolchainDigest,
		ContractDigest:  key.contractDigest}
	if value.result.semanticChecked && !value.result.semanticValid {
		observation.Reason = "SEMANTIC_INVALID"
		return finalizeCompletionProvenance(observation)
	}
	for _, required := range []struct {
		name, value string
	}{
		{"SOURCE_DIGEST", observation.SourceDigest},
		{"SEMANTIC_DIGEST", observation.SemanticDigest},
		{"PROFILE_DIGEST", observation.ProfileDigest},
		{"TOOLCHAIN_DIGEST", observation.ToolchainDigest},
		{"CONTRACT_DIGEST", observation.ContractDigest},
	} {
		if !knownLSPProvenanceDigest(required.value).Known() {
			observation.Reason = "MISSING_" + required.name
			return finalizeCompletionProvenance(observation)
		}
	}
	list := server.completionAt(uri, position, true)
	if list != nil {
		observation.Candidates = completionProvenanceCandidates(list.Items)
	}
	observation.DocumentProvenanceDigest = diagnosticDocumentProvenanceDigest(value, key)
	observation.Decision = completionProvenanceClosed
	observation.Reason = "COMPLETION_SURFACE_BOUND"
	return finalizeCompletionProvenance(observation)
}

func completionProvenanceCandidates(values []CompletionItem) []completionProvenanceCandidate {
	result := make([]completionProvenanceCandidate, 0, len(values))
	for _, value := range values {
		item := completionProvenanceCandidate{Label: value.Label, Kind: value.Kind, Detail: value.Detail, Documentation: value.Documentation}
		item.OriginDigest = completionProvenanceCandidateDigest(item)
		result = append(result, item)
	}
	sort.SliceStable(result, func(left, right int) bool {
		first, second := result[left], result[right]
		if first.Label != second.Label {
			return first.Label < second.Label
		}
		if first.Kind != second.Kind {
			return first.Kind < second.Kind
		}
		if first.Detail != second.Detail {
			return first.Detail < second.Detail
		}
		return first.Documentation < second.Documentation
	})
	return result
}

func completionProvenanceCandidateDigest(value completionProvenanceCandidate) string {
	value.OriginDigest = ""
	payload, _ := json.Marshal(value)
	return cache.HashBytes(payload).String()
}

func completionProvenanceCandidateMapDigest(values []completionProvenanceCandidate) string {
	payload, _ := json.Marshal(values)
	return cache.HashBytes(payload).String()
}

func finalizeCompletionProvenance(value completionProvenanceObservation) completionProvenanceObservation {
	value.CandidateMapDigest = completionProvenanceCandidateMapDigest(value.Candidates)
	value.ObservationDigest = completionProvenanceObservationDigest(value)
	return value
}

func completionProvenanceObservationDigest(value completionProvenanceObservation) string {
	value.ObservationDigest = ""
	payload, _ := json.Marshal(value)
	return cache.HashBytes(payload).String()
}

func validateCompletionProvenance(value completionProvenanceObservation) error {
	if value.Schema != completionProvenanceSchema || value.URI == "" || !value.NonAuthorizing || value.Reason == "" {
		return errors.New("completion provenance identity is invalid")
	}
	if value.Decision != completionProvenanceClosed && value.Decision != completionProvenanceUnknown {
		return errors.New("completion provenance decision is invalid")
	}
	for _, digest := range []string{value.SourceDigest, value.SemanticDigest, value.ProfileDigest, value.ToolchainDigest, value.ContractDigest, value.DocumentProvenanceDigest} {
		if digest != "" && !knownLSPProvenanceDigest(digest).Known() {
			return errors.New("completion provenance binding is invalid")
		}
	}
	if !knownLSPProvenanceDigest(value.CandidateMapDigest).Known() || value.CandidateMapDigest != completionProvenanceCandidateMapDigest(value.Candidates) {
		return errors.New("completion provenance candidate map is invalid")
	}
	for _, item := range value.Candidates {
		if item.Label == "" || !knownLSPProvenanceDigest(item.OriginDigest).Known() || item.OriginDigest != completionProvenanceCandidateDigest(item) {
			return errors.New("completion provenance candidate origin is invalid")
		}
	}
	if value.Decision == completionProvenanceClosed && (!knownLSPProvenanceDigest(value.SourceDigest).Known() || !knownLSPProvenanceDigest(value.SemanticDigest).Known() || !knownLSPProvenanceDigest(value.ProfileDigest).Known() || !knownLSPProvenanceDigest(value.ToolchainDigest).Known() || !knownLSPProvenanceDigest(value.ContractDigest).Known() || !knownLSPProvenanceDigest(value.DocumentProvenanceDigest).Known()) {
		return errors.New("completion provenance closed binding is incomplete")
	}
	if !knownLSPProvenanceDigest(value.ObservationDigest).Known() || value.ObservationDigest != completionProvenanceObservationDigest(value) {
		return errors.New("completion provenance observation is invalid")
	}
	return nil
}

func decodeCompletionProvenance(payload json.RawMessage) (completionProvenanceObservation, error) {
	var value completionProvenanceObservation
	if err := json.Unmarshal(payload, &value); err != nil {
		return completionProvenanceObservation{}, err
	}
	if err := validateCompletionProvenance(value); err != nil {
		return completionProvenanceObservation{}, err
	}
	return value, nil
}
