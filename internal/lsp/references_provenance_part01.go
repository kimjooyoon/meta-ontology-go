package lsp

import (
	"context"
	"encoding/json"
	"errors"
	"sort"

	"github.com/kimjooyoon/meta-ontology-go/internal/cache"
)

const referencesProvenanceSchema = "gooo/lsp-references-provenance/v1"

const (
	referencesProvenanceClosed  = "CLOSED"
	referencesProvenanceUnknown = "UNKNOWN"
)

type referencesProvenanceParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Position     *Position              `json:"position"`
	Context      ReferenceContext       `json:"context"`
}

type referencesProvenanceLocation struct {
	URI          string `json:"uri"`
	Range        Range  `json:"range"`
	Role         string `json:"role"`
	OriginDigest string `json:"origin_digest"`
}

type referencesProvenanceObservation struct {
	Schema             string                         `json:"schema"`
	URI                string                         `json:"uri"`
	SourceDigest       string                         `json:"source_digest,omitempty"`
	SemanticDigest     string                         `json:"semantic_digest,omitempty"`
	ProfileDigest      string                         `json:"profile_digest,omitempty"`
	ToolchainDigest    string                         `json:"toolchain_digest,omitempty"`
	ContractDigest     string                         `json:"contract_digest,omitempty"`
	TargetName         string                         `json:"target_name,omitempty"`
	TargetSemanticID   string                         `json:"target_semantic_id,omitempty"`
	IncludeDeclaration bool                           `json:"include_declaration"`
	Locations          []referencesProvenanceLocation `json:"locations"`
	LocationMapDigest  string                         `json:"location_map_digest"`
	Decision           string                         `json:"decision"`
	Reason             string                         `json:"reason"`
	NonAuthorizing     bool                           `json:"non_authorizing"`
	ObservationDigest  string                         `json:"observation_digest"`
}

func (server *Server) referencesProvenanceRequest(ctx context.Context, request requestEnvelope) (*responseEnvelope, [][]byte, error) {
	var params referencesProvenanceParams
	if decodeParams(request.Params, &params) != nil || params.TextDocument.URI == "" || params.Position == nil {
		return responseOrNil(request.ID, invalidParams, "Invalid references provenance parameters"), nil, nil
	}
	if err := server.refresh(ctx, params.TextDocument.URI); err != nil {
		return featureErrorResponse(request.ID, err, ctx)
	}
	server.mu.RLock()
	stored, exists := server.documents[params.TextDocument.URI]
	if exists {
		key := stored.cacheKey
		document := documentCopy(stored)
		server.mu.RUnlock()
		targetID, targetName, err := referenceTargetForDocument(document, *params.Position)
		if err != nil {
			return responseOrNil(request.ID, invalidParams, "Invalid references provenance position"), nil, nil
		}
		observation := observeReferencesProvenance(params.TextDocument.URI, document, key, targetID, targetName, params.Context.IncludeDeclaration)
		return resultResponse(request.ID, observation), nil, nil
	}
	server.mu.RUnlock()
	return resultResponse(request.ID, nil), nil, nil
}

func observeReferencesProvenance(uri string, document document, key documentCacheKey, targetID, targetName string, includeDeclaration bool) referencesProvenanceObservation {
	value := referencesProvenanceObservation{
		Schema:             referencesProvenanceSchema,
		URI:                uri,
		SourceDigest:       key.sourceDigest,
		SemanticDigest:     document.result.semanticDigest,
		ProfileDigest:      key.profileDigest,
		ToolchainDigest:    key.toolchainDigest,
		ContractDigest:     key.contractDigest,
		TargetName:         targetName,
		TargetSemanticID:   targetID,
		IncludeDeclaration: includeDeclaration,
		Locations:          []referencesProvenanceLocation{},
		Decision:           referencesProvenanceUnknown,
		Reason:             "MISSING_SOURCE_DIGEST",
		NonAuthorizing:     true,
	}
	if document.result.semanticChecked && !document.result.semanticValid {
		value.Reason = "SEMANTIC_INVALID"
		return finalizeReferencesProvenance(value)
	}
	for _, required := range []struct {
		name  string
		value string
	}{
		{name: "SOURCE_DIGEST", value: value.SourceDigest},
		{name: "SEMANTIC_DIGEST", value: value.SemanticDigest},
		{name: "PROFILE_DIGEST", value: value.ProfileDigest},
		{name: "TOOLCHAIN_DIGEST", value: value.ToolchainDigest},
		{name: "CONTRACT_DIGEST", value: value.ContractDigest},
	} {
		if !knownLSPProvenanceDigest(required.value).Known() {
			value.Reason = "MISSING_" + required.name
			return finalizeReferencesProvenance(value)
		}
	}
	if targetName == "" {
		value.Reason = "TARGET_NOT_FOUND"
		return finalizeReferencesProvenance(value)
	}
	if targetID == "" {
		if referenceTargetAmbiguous(allSymbols(document.result), targetName) {
			value.Reason = "TARGET_AMBIGUOUS"
		} else {
			value.Reason = "TARGET_SEMANTIC_ID_MISSING"
		}
		return finalizeReferencesProvenance(value)
	}
	value.Locations = referencesProvenanceLocationsForTarget(
		value.URI, targetID, targetName, allSymbols(document.result), document.result.References, includeDeclaration,
	)
	value.Decision = referencesProvenanceClosed
	value.Reason = "REFERENCE_SURFACE_BOUND"
	return finalizeReferencesProvenance(value)
}

func referencesProvenanceLocationsForTarget(uri, targetID, targetName string, symbols []Symbol, references []Reference, includeDeclaration bool) []referencesProvenanceLocation {
	result := make([]referencesProvenanceLocation, 0, len(references)+1)
	for _, reference := range references {
		if targetID != "" {
			if reference.ID != targetID {
				continue
			}
		} else if reference.Name != targetName {
			continue
		}
		value := referencesProvenanceLocation{URI: uri, Range: reference.Range, Role: "reference"}
		value.OriginDigest = referencesProvenanceLocationDigest(value)
		result = append(result, value)
	}
	if includeDeclaration {
		for _, symbol := range symbols {
			if symbol.SelectionRange == (Range{}) {
				continue
			}
			if targetID != "" {
				if symbol.ID != targetID {
					continue
				}
			} else if symbol.Name != targetName {
				continue
			}
			value := referencesProvenanceLocation{URI: uri, Range: symbol.SelectionRange, Role: "declaration"}
			value.OriginDigest = referencesProvenanceLocationDigest(value)
			result = append(result, value)
			break
		}
	}
	sort.SliceStable(result, func(left, right int) bool {
		first, second := result[left], result[right]
		if first.URI != second.URI {
			return first.URI < second.URI
		}
		if first.Range.Start != second.Range.Start {
			return positionLess(first.Range.Start, second.Range.Start)
		}
		if first.Range.End != second.Range.End {
			return positionLess(first.Range.End, second.Range.End)
		}
		return first.Role < second.Role
	})
	return result
}

func referencesProvenanceLocationDigest(value referencesProvenanceLocation) string {
	value.OriginDigest = ""
	payload, _ := json.Marshal(value)
	return cache.HashBytes(payload).String()
}

func referencesProvenanceLocationMapDigest(values []referencesProvenanceLocation) string {
	payload, _ := json.Marshal(values)
	return cache.HashBytes(payload).String()
}

func finalizeReferencesProvenance(value referencesProvenanceObservation) referencesProvenanceObservation {
	value.LocationMapDigest = referencesProvenanceLocationMapDigest(value.Locations)
	value.ObservationDigest = referencesProvenanceObservationDigest(value)
	return value
}

func referencesProvenanceObservationDigest(value referencesProvenanceObservation) string {
	value.ObservationDigest = ""
	payload, _ := json.Marshal(value)
	return cache.HashBytes(payload).String()
}

func validateReferencesProvenance(value referencesProvenanceObservation) error {
	if value.Schema != referencesProvenanceSchema || value.URI == "" || !value.NonAuthorizing || value.Reason == "" {
		return errors.New("references provenance identity is invalid")
	}
	if value.Decision != referencesProvenanceClosed && value.Decision != referencesProvenanceUnknown {
		return errors.New("references provenance decision is invalid")
	}
	for _, digest := range []string{value.SourceDigest, value.SemanticDigest, value.ProfileDigest, value.ToolchainDigest, value.ContractDigest} {
		if digest != "" && !knownLSPProvenanceDigest(digest).Known() {
			return errors.New("references provenance binding is invalid")
		}
	}
	if !knownLSPProvenanceDigest(value.LocationMapDigest).Known() || value.LocationMapDigest != referencesProvenanceLocationMapDigest(value.Locations) {
		return errors.New("references provenance location map is invalid")
	}
	for _, location := range value.Locations {
		if location.URI != value.URI || (location.Role != "reference" && location.Role != "declaration") ||
			!knownLSPProvenanceDigest(location.OriginDigest).Known() || location.OriginDigest != referencesProvenanceLocationDigest(location) {
			return errors.New("references provenance location origin is invalid")
		}
	}
	if value.Decision == referencesProvenanceClosed {
		if value.TargetName == "" || value.TargetSemanticID == "" ||
			!knownLSPProvenanceDigest(value.SourceDigest).Known() || !knownLSPProvenanceDigest(value.SemanticDigest).Known() ||
			!knownLSPProvenanceDigest(value.ProfileDigest).Known() || !knownLSPProvenanceDigest(value.ToolchainDigest).Known() ||
			!knownLSPProvenanceDigest(value.ContractDigest).Known() {
			return errors.New("references provenance closed binding is incomplete")
		}
	}
	if !knownLSPProvenanceDigest(value.ObservationDigest).Known() || value.ObservationDigest != referencesProvenanceObservationDigest(value) {
		return errors.New("references provenance observation is invalid")
	}
	return nil
}