package lsp

import (
	"sort"
)

func (server *Server) hover(params TextDocumentPositionParams) (*Hover, bool) {
	server.mu.RLock()
	document, ok := server.documents[params.TextDocument.URI]
	if ok {
		copyValue := documentCopy(document)
		document = &copyValue
	}
	server.mu.RUnlock()
	if !ok {
		return nil, false
	}
	symbol, ok := symbolAtPosition(*document, params.Position)
	if !ok {
		return nil, false
	}
	rangeValue := symbol.SelectionRange
	if name, start, end, found := wordAt(document.text, params.Position); found {
		if candidate, candidateOK := symbolNamed(allSymbols(document.result), name); candidateOK && candidate.Name == symbol.Name {
			if value, err := byteRange(document.text, start, end); err == nil {
				rangeValue = value
			}
		}
	}
	return &Hover{Contents: MarkupContent{Kind: "plaintext", Value: symbol.Detail}, Range: &rangeValue}, true
}
func (server *Server) completion(uri string) *CompletionList {
	items := []CompletionItem{
		{Label: "activity", Kind: int(SymbolKeyword), Detail: "gooo keyword"},
		{Label: "entity", Kind: int(SymbolKeyword), Detail: "gooo keyword"},
		{Label: "namespace", Kind: int(SymbolKeyword), Detail: "gooo keyword"},
		{Label: "package", Kind: int(SymbolKeyword), Detail: "gooo keyword"},
	}
	server.mu.RLock()
	document, ok := server.documents[uri]
	if ok {
		copyValue := documentCopy(document)
		document = &copyValue
	}
	server.mu.RUnlock()
	if ok {
		for _, symbol := range allSymbols(document.result) {
			item := CompletionItem{Label: symbol.Name, Kind: int(symbol.Kind), Detail: symbol.Detail}
			if symbol.ID != "" {
				item.Documentation = "semantic ID: " + symbol.ID
			}
			items = append(items, item)
		}
	}
	sort.SliceStable(items, func(left, right int) bool { return items[left].Label < items[right].Label })
	return &CompletionList{Items: uniqueCompletionItems(items)}
}
