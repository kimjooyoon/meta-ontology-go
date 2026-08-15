package lsp

import "sort"

func (server *Server) documentSymbolRequest(request requestEnvelope) (*responseEnvelope, [][]byte, error) {
	var params DocumentSymbolParams
	if decodeParams(request.Params, &params) != nil || params.TextDocument.URI == "" {
		return responseOrNil(request.ID, invalidParams, "Invalid document symbol parameters"), nil, nil
	}
	symbols, exists := server.documentSymbols(params.TextDocument.URI)
	if !exists {
		return resultResponse(request.ID, nil), nil, nil
	}
	return resultResponse(request.ID, symbols), nil, nil
}

func (server *Server) documentSymbols(uri string) ([]DocumentSymbol, bool) {
	server.mu.RLock()
	document, exists := server.documents[uri]
	if exists {
		result := canonicalDocumentSymbols(allSymbols(document.result))
		server.mu.RUnlock()
		return result, true
	}
	server.mu.RUnlock()
	return nil, false
}

func allSymbols(result ParseResult) []Symbol {
	all := make([]Symbol, 0, len(result.Headers)+len(result.Symbols))
	all = append(all, result.Headers...)
	all = append(all, result.Symbols...)
	return all
}

func symbolDetail(symbol Symbol) string {
	if symbol.ID == "" {
		return symbol.Detail
	}
	if symbol.Detail == "" {
		return "semantic ID: " + symbol.ID
	}
	return symbol.Detail + " (semantic ID: " + symbol.ID + ")"
}

func canonicalDocumentSymbols(symbols []Symbol) []DocumentSymbol {
	result := make([]DocumentSymbol, 0, len(symbols))
	for _, symbol := range symbols {
		result = append(result, DocumentSymbol{
			ID: symbol.ID, Name: symbol.Name, Detail: symbolDetail(symbol), Kind: symbol.Kind,
			Range: symbol.Range, SelectionRange: symbol.SelectionRange,
		})
	}
	sort.SliceStable(result, func(left, right int) bool {
		first, second := result[left], result[right]
		if first.Range.Start != second.Range.Start {
			return positionLess(first.Range.Start, second.Range.Start)
		}
		if first.Range.End != second.Range.End {
			return positionLess(first.Range.End, second.Range.End)
		}
		if first.SelectionRange.Start != second.SelectionRange.Start {
			return positionLess(first.SelectionRange.Start, second.SelectionRange.Start)
		}
		if first.SelectionRange.End != second.SelectionRange.End {
			return positionLess(first.SelectionRange.End, second.SelectionRange.End)
		}
		if first.ID != second.ID {
			return first.ID < second.ID
		}
		if first.Name != second.Name {
			return first.Name < second.Name
		}
		if first.Detail != second.Detail {
			return first.Detail < second.Detail
		}
		return first.Kind < second.Kind
	})
	return result
}
