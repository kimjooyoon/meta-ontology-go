package lsp

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
