package lsp

func (server *Server) definition(params TextDocumentPositionParams) []Location {
	server.mu.RLock()
	document, ok := server.documents[params.TextDocument.URI]
	if ok {
		copyValue := documentCopy(document)
		document = &copyValue
	}
	server.mu.RUnlock()
	if !ok {
		return nil
	}
	if document.result.semanticChecked && !document.result.semanticValid {
		return nil
	}
	target, ok := definitionTargetForDocument(*document, params.Position)
	if !ok {
		return nil
	}
	symbol, found := resolveDefinitionSymbol(allSymbols(document.result), target)
	if !found {
		return nil
	}
	return []Location{{URI: params.TextDocument.URI, Range: symbol.SelectionRange}}
}
func symbolAtPosition(document document, position Position) (Symbol, bool) {
	for _, symbol := range allSymbols(document.result) {
		if positionInRange(position, symbol.SelectionRange) || (symbol.hasIdentity && positionInRange(position, symbol.identityRange)) {
			return symbol, true
		}
	}
	for _, reference := range document.result.References {
		if !positionInRange(position, reference.Range) || reference.ID == "" {
			continue
		}
		if symbol, ok := symbolByID(allSymbols(document.result), reference.ID); ok {
			return symbol, true
		}
	}
	name, _, _, ok := wordAt(document.text, position)
	if !ok {
		return Symbol{}, false
	}
	return symbolNamed(allSymbols(document.result), name)
}
func symbolByID(symbols []Symbol, id string) (Symbol, bool) {
	var match Symbol
	for _, symbol := range symbols {
		if symbol.ID != id {
			continue
		}
		if match.ID != "" {
			return Symbol{}, false
		}
		match = symbol
	}
	return match, match.ID != ""
}
func definitionTargetForDocument(document document, position Position) (string, bool) {
	for _, symbol := range allSymbols(document.result) {
		if symbol.hasIdentity && positionInRange(position, symbol.identityRange) {
			return symbol.ID, true
		}
	}
	for _, reference := range document.result.References {
		if reference.ID != "" && positionInRange(position, reference.Range) {
			return reference.ID, true
		}
	}
	return definitionTarget(document.text, position, allSymbols(document.result))
}
