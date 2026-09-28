package lsp

import (
	"context"
	"encoding/json"
	"errors"
	"sort"
)

const storyProvenanceSchema = "gooo/lsp-story-provenance/v1"

type StoryProvenanceParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	SemanticID   string                 `json:"semantic_id"`
}

type storyProvenanceOrigin struct {
	Kind         string          `json:"kind"`
	Name         string          `json:"name,omitempty"`
	SemanticID   string          `json:"semantic_id,omitempty"`
	Range        json.RawMessage `json:"range,omitempty"`
	OriginDigest string          `json:"origin_digest,omitempty"`
}

type storyProvenanceObservation struct {
	Schema           string                  `json:"schema"`
	URI              string                  `json:"uri"`
	SemanticID       string                  `json:"semantic_id"`
	Status           string                  `json:"status"`
	Reason           string                  `json:"reason"`
	SourceDigest     string                  `json:"source_digest,omitempty"`
	SemanticDigest   string                  `json:"semantic_digest,omitempty"`
	ProfileDigest    string                  `json:"profile_digest,omitempty"`
	ToolchainDigest  string                  `json:"toolchain_digest,omitempty"`
	ContractDigest   string                  `json:"contract_digest,omitempty"`
	DeclarationFound bool                    `json:"declaration_found"`
	Origins          []storyProvenanceOrigin `json:"origins"`
	NonAuthorizing   bool                    `json:"non_authorizing"`
	StoryDigest      string                  `json:"story_digest"`
}

func (server *Server) storyProvenanceRequest(ctx context.Context, request requestEnvelope) (*responseEnvelope, [][]byte, error) {
	_ = ctx
	var params StoryProvenanceParams
	if err := json.Unmarshal(request.Params, &params); err != nil {
		return nil, nil, err
	}
	if params.TextDocument.URI == "" || params.SemanticID == "" {
		return nil, nil, errors.New("story provenance requires textDocument URI and semantic_id")
	}

	observation := storyProvenanceObservation{
		Schema:         storyProvenanceSchema,
		URI:            params.TextDocument.URI,
		SemanticID:     params.SemanticID,
		Status:         "unknown",
		Reason:         "MISSING_DOCUMENT",
		Origins:        []storyProvenanceOrigin{},
		NonAuthorizing: true,
	}
	document, ok := server.documents[params.TextDocument.URI]
	if !ok || document == nil {
		observation.StoryDigest = storyProvenanceDigest(observation)
		return resultResponse(request.ID, observation), nil, nil
	}
	observation.SourceDigest = document.cacheKey.sourceDigest
	observation.ProfileDigest = document.cacheKey.profileDigest
	observation.ToolchainDigest = document.cacheKey.toolchainDigest
	observation.ContractDigest = document.cacheKey.contractDigest
	observation.SemanticDigest = document.result.semanticDigest
	if observation.SourceDigest == "" || observation.SemanticDigest == "" {
		observation.Reason = "MISSING_SEMANTIC_IDENTITY"
		observation.StoryDigest = storyProvenanceDigest(observation)
		return resultResponse(request.ID, observation), nil, nil
	}

	origins := make([]storyProvenanceOrigin, 0)
	origins = appendStoryProvenanceOrigins(origins, document.result.Symbols, "symbol", params.SemanticID)
	origins = appendStoryProvenanceOrigins(origins, document.result.References, "reference", params.SemanticID)
	sort.Slice(origins, func(i, j int) bool {
		if origins[i].Kind != origins[j].Kind {
			return origins[i].Kind < origins[j].Kind
		}
		if origins[i].Name != origins[j].Name {
			return origins[i].Name < origins[j].Name
		}
		return origins[i].OriginDigest < origins[j].OriginDigest
	})
	observation.Origins = origins
	if len(origins) == 0 {
		observation.Reason = "UNKNOWN_SEMANTIC_ID"
	} else {
		observation.Reason = "MISSING_PROVENANCE_LEDGER"
		observation.DeclarationFound = true
	}
	observation.StoryDigest = storyProvenanceDigest(observation)
	return resultResponse(request.ID, observation), nil, nil
}

func appendStoryProvenanceOrigins(origins []storyProvenanceOrigin, values any, kind, semanticID string) []storyProvenanceOrigin {
	raw, err := json.Marshal(values)
	if err != nil {
		return origins
	}
	var entries []map[string]any
	if err := json.Unmarshal(raw, &entries); err != nil {
		return origins
	}
	for _, entry := range entries {
		if storyProvenanceField(entry, "semantic_id", "semanticID", "id") != semanticID {
			continue
		}
		origin := storyProvenanceOrigin{
			Kind:         kind,
			Name:         storyProvenanceField(entry, "name"),
			SemanticID:   semanticID,
			OriginDigest: storyProvenanceField(entry, "origin_digest", "originDigest", "provenance_digest", "provenanceDigest"),
		}
		if value, ok := entry["range"]; ok {
			if encoded, err := json.Marshal(value); err == nil {
				origin.Range = encoded
			}
		}
		origins = append(origins, origin)
	}
	return origins
}

func storyProvenanceField(entry map[string]any, keys ...string) string {
	for _, key := range keys {
		if value, ok := entry[key].(string); ok {
			return value
		}
	}
	return ""
}

func storyProvenanceDigest(observation storyProvenanceObservation) string {
	observation.StoryDigest = ""
	encoded, err := json.Marshal(observation)
	if err != nil {
		return ""
	}
	return digestText(string(encoded))
}
