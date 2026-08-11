package lsp

import (
	"sort"
	"unicode"
	"unicode/utf8"
)

func (server *Server) hover(params TextDocumentPositionParams) (*Hover, bool) {
	document, ok := server.documents[params.TextDocument.URI]
	if !ok {
		return nil, false
	}
	name, start, end, ok := wordAt(document.text, params.Position)
	if !ok {
		return nil, false
	}
	symbol, ok := symbolNamed(document.result.Symbols, name)
	if !ok {
		return nil, false
	}
	rangeValue, err := byteRange(document.text, start, end)
	if err != nil {
		return nil, false
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
	if document, ok := server.documents[uri]; ok {
		for _, symbol := range document.result.Symbols {
			items = append(items, CompletionItem{Label: symbol.Name, Kind: int(symbol.Kind), Detail: symbol.Detail})
		}
	}
	sort.SliceStable(items, func(left, right int) bool { return items[left].Label < items[right].Label })
	return &CompletionList{Items: uniqueCompletionItems(items)}
}

func (server *Server) definition(params TextDocumentPositionParams) []Location {
	document, ok := server.documents[params.TextDocument.URI]
	if !ok {
		return nil
	}
	name, _, _, ok := wordAt(document.text, params.Position)
	if !ok {
		return nil
	}
	if symbol, found := symbolNamed(document.result.Symbols, name); found {
		return []Location{{URI: params.TextDocument.URI, Range: symbol.SelectionRange}}
	}
	for _, reference := range document.result.References {
		if reference.Name != name {
			continue
		}
		if symbol, found := symbolNamed(document.result.Symbols, name); found {
			return []Location{{URI: params.TextDocument.URI, Range: symbol.SelectionRange}}
		}
	}
	return nil
}

func symbolNamed(symbols []Symbol, name string) (Symbol, bool) {
	for _, symbol := range symbols {
		if symbol.Name == name {
			return symbol, true
		}
	}
	return Symbol{}, false
}

func uniqueCompletionItems(items []CompletionItem) []CompletionItem {
	result := make([]CompletionItem, 0, len(items))
	seen := make(map[string]struct{}, len(items))
	for _, item := range items {
		if _, exists := seen[item.Label]; exists {
			continue
		}
		seen[item.Label] = struct{}{}
		result = append(result, item)
	}
	return result
}

func wordAt(source string, position Position) (string, int, int, bool) {
	offset, err := PositionToOffset(source, position)
	if err != nil {
		return "", 0, 0, false
	}
	start, end := identifierBounds(source, offset)
	if start == end {
		return "", 0, 0, false
	}
	return source[start:end], start, end, true
}

func identifierBounds(source string, offset int) (int, int) {
	if offset > 0 && (offset == len(source) || !identifierAt(source, offset)) {
		_, size := utf8.DecodeLastRuneInString(source[:offset])
		if size > 0 && identifierAt(source, offset-size) {
			offset -= size
		}
	}
	start := offset
	for start > 0 {
		value, size := utf8.DecodeLastRuneInString(source[:start])
		if !isIdentifier(value) {
			break
		}
		start -= size
	}
	end := offset
	for end < len(source) {
		value, size := utf8.DecodeRuneInString(source[end:])
		if !isIdentifier(value) {
			break
		}
		end += size
	}
	return start, end
}

func identifierAt(source string, offset int) bool {
	if offset < 0 || offset >= len(source) {
		return false
	}
	value, _ := utf8.DecodeRuneInString(source[offset:])
	return isIdentifier(value)
}

func isIdentifier(value rune) bool {
	return value == '_' || unicode.IsLetter(value) || unicode.IsDigit(value)
}

func byteRange(source string, start, end int) (Range, error) {
	startPosition, err := OffsetToPosition(source, start)
	if err != nil {
		return Range{}, err
	}
	endPosition, err := OffsetToPosition(source, end)
	if err != nil {
		return Range{}, err
	}
	return Range{Start: startPosition, End: endPosition}, nil
}
