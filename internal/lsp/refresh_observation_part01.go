package lsp

import (
	"context"
	"errors"
	"strings"
)

const (
	refreshObservationSchemaPart01  = "gooo/lsp-refresh-observation/v1"
	refreshObservationPassPart01    = "PASS"
	refreshObservationUnknownPart01 = "UNKNOWN"
)

// RefreshObservationPart01 is an observational receipt for one URI's exact
// refresh identity. It never grants reuse authority; UNKNOWN is retained when
// semantic evidence, parsing, or freshness evidence is incomplete.
type RefreshObservationPart01 struct {
	Schema            string `json:"schema"`
	URI               string `json:"uri"`
	Decision          string `json:"decision"`
	Reason            string `json:"reason"`
	SourceDigest      string `json:"source_digest"`
	SemanticDigest    string `json:"semantic_digest,omitempty"`
	ProfileDigest     string `json:"profile_digest"`
	ToolchainDigest   string `json:"toolchain_digest"`
	ContractDigest    string `json:"contract_digest"`
	ParseCalls        int    `json:"parse_calls"`
	CacheHits         int    `json:"cache_hits"`
	StaleResultCount  int    `json:"stale_result_count"`
	MissingStageIndex int    `json:"missing_stage_index"`
}

// ObserveRefreshPart01 exposes an explicit observation boundary without
// changing the existing LSP refresh path. The returned receipt is evidence
// only; it never authorizes reuse or promotion.
func (server *Server) ObserveRefreshPart01(ctx context.Context, uri string) (RefreshObservationPart01, error) {
	if err := ctx.Err(); err != nil {
		return RefreshObservationPart01{}, err
	}
	server.mu.RLock()
	document, exists := server.documents[uri]
	if exists {
		source, cachedKey, cachedResult := document.text, document.cacheKey, document.result
		server.mu.RUnlock()
		expectedKey := server.cacheKey(source)
		if cachedKey == expectedKey {
			server.recordRefreshObservationPart01(uri, expectedKey, cachedResult, refreshObservationPassPart01, "EXACT_CACHE_HIT", false, true, false)
			return server.refreshObservationPart01(uri), nil
		}
		if err := server.refresh(ctx, uri); err != nil {
			reason := "PARSE_FAILED"
			stale := errors.Is(err, ErrStaleResult)
			if stale {
				reason = "STALE_RESULT_SUPPRESSED"
			}
			server.recordRefreshObservationPart01(uri, expectedKey, ParseResult{}, refreshObservationUnknownPart01, reason, true, false, stale)
			return server.refreshObservationPart01(uri), err
		}
		server.mu.RLock()
		current, stillOpen := server.documents[uri]
		if stillOpen {
			cachedKey, cachedResult = current.cacheKey, current.result
		}
		server.mu.RUnlock()
		if stillOpen && cachedKey == expectedKey {
			server.recordRefreshObservationPart01(uri, expectedKey, cachedResult, refreshObservationPassPart01, "PARSE_REFRESH", true, false, false)
		} else {
			server.recordRefreshObservationPart01(uri, expectedKey, ParseResult{}, refreshObservationUnknownPart01, "MISSING_REFRESH_EVIDENCE", true, false, false)
		}
		return server.refreshObservationPart01(uri), nil
	}
	server.mu.RUnlock()
	return RefreshObservationPart01{}, errors.New("lsp: document is not open")
}

func (server *Server) recordRefreshObservationPart01(uri string, key documentCacheKey, result ParseResult, decision, reason string, parseCall, cacheHit, stale bool) {
	server.mu.Lock()
	defer server.mu.Unlock()
	if server.refreshObservations == nil {
		server.refreshObservations = make(map[string]RefreshObservationPart01)
	}
	observation := server.refreshObservations[uri]
	if observation.Schema != refreshObservationSchemaPart01 ||
		observation.SourceDigest != key.sourceDigest ||
		observation.ProfileDigest != key.profileDigest ||
		observation.ToolchainDigest != key.toolchainDigest ||
		observation.ContractDigest != key.contractDigest {
		observation = RefreshObservationPart01{
			Schema: refreshObservationSchemaPart01, URI: uri,
			SourceDigest: key.sourceDigest, ProfileDigest: key.profileDigest,
			ToolchainDigest: key.toolchainDigest, ContractDigest: key.contractDigest,
			MissingStageIndex: 1,
		}
	}
	if parseCall {
		observation.ParseCalls++
	}
	if cacheHit {
		observation.CacheHits++
	}
	if stale {
		observation.StaleResultCount++
	}
	observation.Decision = decision
	observation.Reason = reason
	if decision == refreshObservationPassPart01 && strings.TrimSpace(result.semanticDigest) != "" {
		observation.SemanticDigest = result.semanticDigest
		observation.MissingStageIndex = -1
	} else {
		observation.Decision = refreshObservationUnknownPart01
		if strings.TrimSpace(observation.Reason) == "" || decision == refreshObservationPassPart01 {
			observation.Reason = "MISSING_SEMANTIC_DIGEST"
		}
		observation.SemanticDigest = ""
		observation.MissingStageIndex = 1
	}
	server.refreshObservations[uri] = observation
}

func (server *Server) RefreshObservationPart01(uri string) (RefreshObservationPart01, bool) {
	server.mu.RLock()
	defer server.mu.RUnlock()
	observation, ok := server.refreshObservations[uri]
	return observation, ok
}

func (server *Server) refreshObservationPart01(uri string) RefreshObservationPart01 {
	observation, _ := server.RefreshObservationPart01(uri)
	return observation
}

func (value RefreshObservationPart01) ValidPart01() bool {
	if value.Schema != refreshObservationSchemaPart01 || strings.TrimSpace(value.URI) == "" ||
		!refreshObservationDigestPart01(value.SourceDigest) ||
		!refreshObservationDigestPart01(value.ProfileDigest) ||
		!refreshObservationDigestPart01(value.ToolchainDigest) ||
		!refreshObservationDigestPart01(value.ContractDigest) ||
		value.ParseCalls < 0 || value.CacheHits < 0 || value.StaleResultCount < 0 ||
		value.MissingStageIndex < -1 {
		return false
	}
	if value.Decision == refreshObservationPassPart01 {
		return value.MissingStageIndex == -1 && strings.TrimSpace(value.SemanticDigest) != ""
	}
	return value.Decision == refreshObservationUnknownPart01 && value.MissingStageIndex >= 0 && strings.TrimSpace(value.Reason) != ""
}

func refreshObservationDigestPart01(value string) bool {
	return strings.HasPrefix(value, "sha256:") && len(strings.TrimPrefix(value, "sha256:")) == 64
}
