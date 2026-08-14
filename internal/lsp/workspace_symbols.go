package lsp

import (
	"encoding/json"
	"sort"
	"strings"
)

type workspaceSymbolParamsWire struct {
	Query *string `json:"query"`
}

func (server *Server) workspaceSymbolRequest(request requestEnvelope) (*responseEnvelope, [][]byte, error) {
	params, ok := decodeWorkspaceSymbolParams(request.Params)
	if !ok {
		return responseOrNil(request.ID, invalidParams, "Invalid workspace symbol parameters"), nil, nil
	}
	return resultResponse(request.ID, server.workspaceSymbols(params.Query)), nil, nil
}

func decodeWorkspaceSymbolParams(raw json.RawMessage) (WorkspaceSymbolParams, bool) {
	if len(raw) == 0 || string(raw) == "null" {
		return WorkspaceSymbolParams{}, false
	}
	var wire workspaceSymbolParamsWire
	if err := json.Unmarshal(raw, &wire); err != nil || wire.Query == nil {
		return WorkspaceSymbolParams{}, false
	}
	return WorkspaceSymbolParams{Query: *wire.Query}, true
}

func (server *Server) workspaceSymbols(query string) []WorkspaceSymbol {
	server.mu.RLock()
	result := make([]WorkspaceSymbol, 0)
	for uri, document := range server.documents {
		for _, symbol := range document.result.Symbols {
			if !workspaceSymbolMatches(symbol, query) {
				continue
			}
			result = append(result, workspaceSymbol(uri, symbol))
		}
	}
	server.mu.RUnlock()
	return canonicalWorkspaceSymbols(result)
}

func workspaceSymbolMatches(symbol Symbol, query string) bool {
	query = strings.ToLower(strings.TrimSpace(query))
	if query == "" {
		return true
	}
	return strings.Contains(strings.ToLower(symbol.Name), query) ||
		strings.Contains(strings.ToLower(symbol.ID), query)
}

func workspaceSymbol(uri string, symbol Symbol) WorkspaceSymbol {
	return WorkspaceSymbol{
		ID: symbol.ID, Name: symbol.Name, Detail: symbol.Detail, Kind: symbol.Kind,
		Location: Location{URI: uri, Range: symbol.Range},
	}
}

func canonicalWorkspaceSymbols(symbols []WorkspaceSymbol) []WorkspaceSymbol {
	result := make([]WorkspaceSymbol, len(symbols))
	copy(result, symbols)
	sort.SliceStable(result, func(left, right int) bool {
		first, second := result[left], result[right]
		if first.Location.URI != second.Location.URI {
			return first.Location.URI < second.Location.URI
		}
		if first.Name != second.Name {
			return first.Name < second.Name
		}
		if first.ID != second.ID {
			return first.ID < second.ID
		}
		if first.Location.Range.Start != second.Location.Range.Start {
			return positionLess(first.Location.Range.Start, second.Location.Range.Start)
		}
		if first.Location.Range.End != second.Location.Range.End {
			return positionLess(first.Location.Range.End, second.Location.Range.End)
		}
		if first.Detail != second.Detail {
			return first.Detail < second.Detail
		}
		return first.Kind < second.Kind
	})
	return result
}
