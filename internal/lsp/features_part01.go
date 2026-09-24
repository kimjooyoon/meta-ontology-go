package lsp

import (
	"sort"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
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
	return &Hover{Contents: MarkupContent{Kind: "plaintext", Value: hoverSymbolDetail(*document, symbol)}, Range: &rangeValue}, true
}
func (server *Server) completion(uri string) *CompletionList {
	return server.completionAt(uri, Position{}, false)
}

func (server *Server) completionAt(uri string, position Position, usePosition bool) *CompletionList {
	keywords := syntax.CanonicalKeywordNames()
	items := make([]CompletionItem, 0, len(keywords))
	server.mu.RLock()
	document, ok := server.documents[uri]
	if ok {
		copyValue := documentCopy(document)
		document = &copyValue
	}
	server.mu.RUnlock()
	prefix := ""
	cursorWord := ""
	cursorInside := false
	if ok {
		if usePosition {
			word, _, end, found := wordAt(document.text, position)
			offset, offsetErr := PositionToOffset(document.text, position)
			if found && offset < end {
				cursorWord = word
				cursorInside = true
			}
			if !found || offsetErr != nil || offset != end {
				prefix = ""
			} else {
				prefix = word
			}
		}
	}
	for _, keyword := range keywords {
		if !completionMatchesPrefix(keyword, prefix) {
			continue
		}
		items = append(items, CompletionItem{Label: keyword, Kind: int(SymbolKeyword), Detail: "gooo keyword"})
	}
	if ok {
		expectedKind, contextAware := completionExpectedSymbolKind(document.text, position, usePosition)
		for _, symbol := range allSymbols(document.result) {
			if contextAware && symbol.Kind != expectedKind {
				if !(cursorInside && symbol.Kind == SymbolFunction && strings.Contains(strings.ToLower(symbol.Name), strings.ToLower(cursorWord))) {
					continue
				}
			}
			if !completionMatchesPrefix(symbol.Name, prefix) {
				continue
			}
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

func completionMatchesPrefix(label, prefix string) bool {
	return prefix == "" || strings.HasPrefix(strings.ToLower(label), strings.ToLower(prefix))
}
