package lsp

import (
	"context"
)

func definitionTarget(source string, position Position, symbols []Symbol) (string, bool) {
	for _, symbol := range symbols {
		if symbol.hasIdentity && positionInRange(position, symbol.identityRange) {
			return symbol.ID, true
		}
	}
	target, _, _, ok := wordAt(source, position)
	return target, ok
}
func positionInRange(position Position, value Range) bool {
	if position.Line < value.Start.Line || position.Line > value.End.Line {
		return false
	}
	if position.Line == value.Start.Line && position.Character < value.Start.Character {
		return false
	}
	if position.Line == value.End.Line && position.Character >= value.End.Character {
		return false
	}
	return true
}
func resolveDefinitionSymbol(symbols []Symbol, target string) (Symbol, bool) {
	var match Symbol
	for _, symbol := range symbols {
		if symbol.Name != target && symbol.ID != target {
			continue
		}
		if match.Name != "" || match.ID != "" {
			return Symbol{}, false
		}
		match = symbol
	}
	if match.Name == "" && match.ID == "" {
		return Symbol{}, false
	}
	return match, true
}
func (server *Server) refresh(ctx context.Context, uri string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	server.mu.RLock()
	document, exists := server.documents[uri]
	if exists {
		version, source, cachedKey := document.version, document.text, document.cacheKey
		cachedResult := document.result
		server.mu.RUnlock()
		expectedKey := server.cacheKey(source)
		if cachedKey == expectedKey {
			server.recordRefreshObservationPart01(uri, expectedKey, cachedResult, refreshObservationPassPart01, "EXACT_CACHE_HIT", false, true, false)
			return nil
		}
		result, err := server.parse(ctx, uri, source)
		if err != nil {
			server.recordRefreshObservationPart01(uri, expectedKey, ParseResult{}, refreshObservationUnknownPart01, "PARSE_FAILED", true, false, false)
			return err
		}
		server.mu.Lock()
		current, stillOpen := server.documents[uri]
		if !stillOpen || current.version != version || current.text != source {
			server.mu.Unlock()
			server.recordRefreshObservationPart01(uri, expectedKey, ParseResult{}, refreshObservationUnknownPart01, "STALE_RESULT_SUPPRESSED", true, false, true)
			return ErrStaleResult
		}
		current.result = result
		current.cacheKey = expectedKey
		server.mu.Unlock()
		server.recordRefreshObservationPart01(uri, expectedKey, result, refreshObservationPassPart01, "PARSE_REFRESH", true, false, false)
		return nil
	}
	server.mu.RUnlock()
	return nil
}
func documentCopy(value *document) document {
	return document{version: value.version, text: value.text, result: cloneParseResult(value.result)}
}
