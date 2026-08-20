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
	server.mu.RLock()
	document, exists := server.documents[uri]
	if exists {
		version, source := document.version, document.text
		server.mu.RUnlock()
		result, err := server.parse(ctx, uri, source)
		if err != nil {
			return err
		}
		server.mu.Lock()
		defer server.mu.Unlock()
		current, stillOpen := server.documents[uri]
		if !stillOpen || current.version != version || current.text != source {
			return ErrStaleResult
		}
		current.result = result
		return nil
	}
	server.mu.RUnlock()
	return nil
}
func documentCopy(value *document) document {
	return document{version: value.version, text: value.text, result: cloneParseResult(value.result)}
}
